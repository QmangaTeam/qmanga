package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/url"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
)

const preferenceColumns = `version,bio,public_profile,show_library,show_activity,allow_follow,notify_releases,notify_replies,notify_followers,email_digest`

func scanPreferences(row pgx.Row) (application.Preferences, error) {
	var p application.Preferences
	err := row.Scan(&p.Version, &p.Bio, &p.PublicProfile, &p.ShowLibrary, &p.ShowActivity, &p.AllowFollow, &p.NotifyReleases, &p.NotifyReplies, &p.NotifyFollowers, &p.EmailDigest)
	return p, err
}
func (r *Repository) Preferences(ctx context.Context, actor string) (application.Preferences, error) {
	p, err := scanPreferences(r.pool.QueryRow(ctx, `SELECT `+preferenceColumns+` FROM account_preferences WHERE user_id=$1`, actor))
	if errors.Is(err, pgx.ErrNoRows) {
		return application.DefaultPreferences(), nil
	}
	return p, err
}
func (r *Repository) SavePreferences(ctx context.Context, actor string, p application.Preferences) (application.Preferences, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)
	// All profile privacy writes and follows lock the target user first.
	var id string
	if err = tx.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 AND NOT is_blocked FOR UPDATE`, actor).Scan(&id); err != nil {
		return p, translate(err)
	}
	args := []any{actor, p.Version, p.Bio, p.PublicProfile, p.ShowLibrary, p.ShowActivity, p.AllowFollow, p.NotifyReleases, p.NotifyReplies, p.NotifyFollowers, p.EmailDigest}
	query := `UPDATE account_preferences SET version=version+1,bio=$3,public_profile=$4,show_library=$5,show_activity=$6,allow_follow=$7,notify_releases=$8,notify_replies=$9,notify_followers=$10,email_digest=$11 WHERE user_id=$1 AND version=$2 RETURNING ` + preferenceColumns
	if p.Version == 0 {
		query = `INSERT INTO account_preferences(user_id,version,bio,public_profile,show_library,show_activity,allow_follow,notify_releases,notify_replies,notify_followers,email_digest) VALUES($1,$2+1,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(user_id) DO NOTHING RETURNING ` + preferenceColumns
	}
	saved, err := scanPreferences(tx.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return p, domain.ErrConflict
	}
	if err != nil {
		return p, err
	}
	return saved, tx.Commit(ctx)
}
func (r *Repository) FollowProfile(ctx context.Context, actor, target string, following bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Lock in stable order to avoid reciprocal-follow deadlocks and serialize blocking.
	rows, err := tx.Query(ctx, `SELECT id FROM users WHERE id=ANY($1::varchar[]) ORDER BY id FOR UPDATE`, []string{actor, target})
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var verified bool
	if err = tx.QueryRow(ctx, `SELECT is_verified FROM users WHERE id=$1 AND NOT is_blocked`, actor).Scan(&verified); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		return err
	}
	if !following {
		_, err = tx.Exec(ctx, `DELETE FROM account_follows WHERE follower_id=$1 AND followee_id=$2`, actor, target)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if !verified {
		return domain.ErrForbidden
	}
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT COALESCE(p.public_profile AND p.allow_follow,false) FROM users u LEFT JOIN account_preferences p ON p.user_id=u.id WHERE u.id=$1 AND NOT u.is_blocked`, target).Scan(&allowed)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	result, err := tx.Exec(ctx, `INSERT INTO account_follows(follower_id,followee_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, actor, target)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 1 {
		_, err = tx.Exec(ctx, `INSERT INTO outbox_events(id,event_type,payload) VALUES($1,'profile_followed',jsonb_build_object('user_id',$2::text,'actor_id',$3::text))`, uuid.NewString(), target, actor)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *Repository) PublicProfile(ctx context.Context, viewer, target string) (application.PublicProfile, error) {
	p := application.PublicProfile{Library: []application.ProfileLibrary{}, Activity: []application.ProfileActivity{}}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT u.id,u.display_name,u.created_at,COALESCE(a.bio,''),COALESCE(a.public_profile,false),COALESCE(a.show_library,false),COALESCE(a.show_activity,false),COALESCE(a.allow_follow,true),u.id=$2,EXISTS(SELECT 1 FROM account_follows WHERE follower_id=$2 AND followee_id=u.id),(SELECT count(*) FROM account_follows f JOIN users x ON x.id=f.follower_id WHERE f.followee_id=u.id AND NOT x.is_blocked),(SELECT count(*) FROM account_follows f JOIN users x ON x.id=f.followee_id WHERE f.follower_id=u.id AND NOT x.is_blocked) FROM users u LEFT JOIN account_preferences a ON a.user_id=u.id WHERE u.id=$1 AND NOT u.is_blocked AND (u.id=$2 OR COALESCE(a.public_profile,false))`, target, viewer).Scan(&p.ID, &p.DisplayName, &p.CreatedAt, &p.Bio, &p.PublicProfile, &p.ShowLibrary, &p.ShowActivity, &p.AllowFollow, &p.IsOwner, &p.IsFollowing, &p.FollowersCount, &p.FollowingCount)
	if err != nil {
		return p, translate(err)
	}
	p.AvatarURL = "/media/avatars/user/" + url.PathEscape(p.ID)
	if p.ShowLibrary || p.IsOwner {
		rows, e := tx.Query(ctx, `SELECT t.id,t.slug,t.name_kk,l.status,COALESCE(t.cover_asset_id,'') FROM library_entries l JOIN titles t ON t.id=l.title_id WHERE l.user_id=$1 AND NOT l.is_removed AND (t.is_listed OR EXISTS(SELECT 1 FROM chapters c JOIN releases r ON r.chapter_id=c.id JOIN revisions v ON v.id=r.current_revision_id AND v.release_id=r.id JOIN projects pr ON pr.id=r.project_id AND pr.title_id=c.title_id AND pr.team_id=r.team_id JOIN teams tm ON tm.id=r.team_id WHERE c.title_id=t.id AND r.status='published' AND r.language='kk' AND v.status='published')) ORDER BY l.updated_at DESC,t.id LIMIT 40`, target)
		if e != nil {
			return p, e
		}
		for rows.Next() {
			var item application.ProfileLibrary
			var cover string
			if e = rows.Scan(&item.TitleID, &item.Slug, &item.Name, &item.Status, &cover); e != nil {
				rows.Close()
				return p, e
			}
			if cover != "" {
				item.CoverURL = "/api/v1/media/" + url.PathEscape(cover)
			}
			p.Library = append(p.Library, item)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return p, e
		}
	}
	if p.ShowActivity || p.IsOwner {
		rows, e := tx.Query(ctx, `SELECT t.name_kk,c.label,(a.day::timestamp AT TIME ZONE 'UTC'),t.slug FROM reading_activity a JOIN releases r ON r.id=a.release_id JOIN revisions v ON v.id=r.current_revision_id AND v.release_id=r.id AND v.status='published' JOIN chapters c ON c.id=r.chapter_id JOIN projects pr ON pr.id=r.project_id AND pr.title_id=c.title_id AND pr.team_id=r.team_id JOIN teams tm ON tm.id=r.team_id JOIN titles t ON t.id=c.title_id WHERE a.user_id=$1 AND r.status='published' AND r.language='kk' ORDER BY a.day DESC,r.id LIMIT 40`, target)
		if e != nil {
			return p, e
		}
		for rows.Next() {
			var item application.ProfileActivity
			var name, label, slug string
			if e = rows.Scan(&name, &label, &item.CreatedAt, &slug); e != nil {
				rows.Close()
				return p, e
			}
			item.Label = name + " · " + label
			item.Href = "/titles/" + url.PathEscape(slug)
			p.Activity = append(p.Activity, item)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return p, e
		}
	}
	return p, tx.Commit(ctx)
}
