package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/community/application"
	"qmanga/internal/community/domain"
	"strings"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }

const visibleTitle = `(EXISTS(SELECT 1 FROM titles WHERE id=$1 AND is_listed) OR EXISTS(SELECT 1 FROM chapters chapter JOIN releases release ON release.chapter_id=chapter.id JOIN revisions revision ON revision.id=release.current_revision_id AND revision.release_id=release.id JOIN projects project ON project.id=release.project_id AND project.title_id=chapter.title_id AND project.team_id=release.team_id JOIN teams team ON team.id=release.team_id WHERE chapter.title_id=$1 AND release.status='published' AND release.language='kk' AND revision.status='published'))`

func available(ctx context.Context, transaction pgx.Tx, title, release string) error {
	var valid bool
	failure := transaction.QueryRow(ctx, `SELECT `+visibleTitle+` AND ($2::varchar='' OR EXISTS(SELECT 1 FROM releases release JOIN chapters chapter ON chapter.id=release.chapter_id WHERE release.id=$2::varchar AND chapter.title_id=$1 AND release.status='published'))`, title, release).Scan(&valid)
	if failure != nil {
		return failure
	}
	if !valid {
		return domain.ErrNotFound
	}
	return nil
}
func actor(ctx context.Context, transaction pgx.Tx, id string, moderator bool) error {
	var verified bool
	var role string
	if failure := transaction.QueryRow(ctx, `SELECT is_verified,role FROM users WHERE id=$1 AND NOT is_blocked FOR UPDATE`, id).Scan(&verified, &role); failure != nil {
		return domain.ErrForbidden
	}
	if !verified || moderator && (role != "admin" && role != "moderator") {
		return domain.ErrForbidden
	}
	return nil
}
func (repository *Repository) write(ctx context.Context, user string, moderator bool, operation func(pgx.Tx) error) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	if failure = actor(ctx, transaction, user, moderator); failure != nil {
		return failure
	}
	if failure = operation(transaction); failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) Discussion(ctx context.Context, title, release, after string, focuses ...string) (application.Discussion, error) {
	result := application.Discussion{Comments: []application.Comment{}}
	focus := ""
	if len(focuses) > 0 {
		focus = focuses[0]
	}
	var valid bool
	if failure := repository.pool.QueryRow(ctx, `SELECT `+visibleTitle, title).Scan(&valid); failure != nil {
		return result, failure
	}
	if !valid {
		return result, domain.ErrNotFound
	}
	if failure := repository.pool.QueryRow(ctx, `SELECT COALESCE(avg(score)::float8,0),count(*) FROM ratings JOIN users rating_user ON rating_user.id=ratings.user_id AND NOT rating_user.is_blocked WHERE title_id=$1`, title).Scan(&result.Score, &result.Ratings); failure != nil {
		return result, failure
	}
	rows, failure := repository.pool.Query(ctx, `WITH roots AS(
	 SELECT id,created_at FROM comments WHERE title_id=$1 AND COALESCE(release_id,'')=$2 AND parent_id IS NULL
 AND ($4='' OR id=(SELECT COALESCE(root_id,parent_id,id) FROM comments WHERE id=$4 AND title_id=$1 AND COALESCE(release_id,'')=$2))
	 AND ($3::varchar='' OR (created_at,id)<(SELECT created_at,id FROM comments WHERE id=$3::varchar AND title_id=$1 AND COALESCE(release_id,'')=$2))
	 ORDER BY created_at DESC,id DESC LIMIT 20
	), selected AS(
	 SELECT id FROM roots UNION SELECT id FROM comments WHERE id IN($4,(SELECT parent_id FROM comments WHERE id=$4)) AND title_id=$1 AND COALESCE(release_id,'')=$2 UNION ALL SELECT reply.id FROM roots CROSS JOIN LATERAL(
	 SELECT id FROM comments WHERE (root_id=roots.id OR (root_id IS NULL AND parent_id=roots.id)) ORDER BY created_at,id LIMIT 10)reply
	)
	 SELECT comment.id,comment.user_id,comment.title_id,COALESCE(comment.release_id,''),COALESCE(comment.parent_id,''),COALESCE(comment.root_id,comment.parent_id,''),
	CASE WHEN comment.is_hidden OR comment.deleted_at IS NOT NULL THEN '' ELSE comment.body END,
	comment.is_spoiler,comment.is_hidden OR comment.deleted_at IS NOT NULL,comment.created_at,
	(SELECT count(*) FROM comments reply WHERE reply.root_id=comment.id OR (reply.root_id IS NULL AND reply.parent_id=comment.id)),(SELECT display_name FROM users WHERE id=comment.user_id),
 COALESCE((SELECT id FROM (SELECT id,created_at FROM comments reply WHERE reply.root_id=comment.id OR (reply.root_id IS NULL AND reply.parent_id=comment.id) ORDER BY created_at,id LIMIT 10) contiguous ORDER BY created_at DESC,id DESC LIMIT 1),'')
	FROM comments comment WHERE comment.id IN(SELECT id FROM selected) ORDER BY comment.created_at,comment.id`, title, release, after, focus)
	if failure != nil {
		return result, failure
	}
	defer rows.Close()
	rootCount := 0
	var oldest application.Comment
	for rows.Next() {
		var item application.Comment
		if failure = rows.Scan(&item.ID, &item.UserID, &item.TitleID, &item.ReleaseID, &item.ParentID, &item.RootID, &item.Body, &item.Spoiler, &item.Hidden, &item.CreatedAt, &item.Replies, &item.DisplayName, &item.ReplyCursor); failure != nil {
			return result, failure
		}
		result.Comments = append(result.Comments, item)
		if item.ParentID == "" {
			rootCount++
			if oldest.ID == "" || item.CreatedAt.Before(oldest.CreatedAt) || item.CreatedAt.Equal(oldest.CreatedAt) && item.ID < oldest.ID {
				oldest = item
			}
		}
	}
	if rootCount == 20 {
		result.Next = oldest.ID
	}
	return result, rows.Err()
}
func (repository *Repository) Comment(ctx context.Context, user string, input domain.Comment) (string, error) {
	failure := repository.write(ctx, user, false, func(transaction pgx.Tx) error {
		if failure := available(ctx, transaction, input.TitleID, input.ReleaseID); failure != nil {
			return failure
		}
		var count int
		if failure := transaction.QueryRow(ctx, `SELECT count(*) FROM comments WHERE user_id=$1 AND created_at>now()-interval '1 minute'`, user).Scan(&count); failure != nil {
			return failure
		}
		if count >= 5 {
			return domain.ErrLimited
		}
		var replyUser, rootID string
		if input.ParentID != "" {
			var title, release string
			var hidden bool
			failure := transaction.QueryRow(ctx, `SELECT title_id,COALESCE(release_id,''),COALESCE(root_id,parent_id,id),user_id,is_hidden OR deleted_at IS NOT NULL FROM comments WHERE id=$1 FOR SHARE`, input.ParentID).Scan(&title, &release, &rootID, &replyUser, &hidden)
			if failure != nil || hidden || !domain.CanReply(input.TitleID, input.ReleaseID, title, release) {
				return domain.ErrInvalid
			}
		}
		if _, failure := transaction.Exec(ctx, `INSERT INTO comments(id,user_id,title_id,release_id,parent_id,root_id,body,is_spoiler)VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7,$8)`, input.ID, user, input.TitleID, input.ReleaseID, input.ParentID, rootID, strings.TrimSpace(input.Body), input.Spoiler); failure != nil {
			return failure
		}
		if replyUser != "" && replyUser != user {
			payload, _ := json.Marshal(map[string]string{"user_id": replyUser, "comment_id": input.ID, "title_id": input.TitleID, "release_id": input.ReleaseID})
			_, failure := transaction.Exec(ctx, `INSERT INTO outbox_events(id,event_type,payload)VALUES($1,'comment_reply',$2)`, uuid.NewString(), payload)
			return failure
		}
		return nil
	})
	return input.ID, failure
}
func (repository *Repository) Edit(ctx context.Context, user, id, body string, spoiler bool) error {
	return repository.write(ctx, user, false, func(transaction pgx.Tx) error {
		var item domain.Comment
		if failure := transaction.QueryRow(ctx, `SELECT user_id,is_hidden OR deleted_at IS NOT NULL FROM comments WHERE id=$1 FOR UPDATE`, id).Scan(&item.UserID, &item.Hidden); failure != nil {
			return domain.ErrNotFound
		}
		if failure := item.Edit(user, body, spoiler); failure != nil {
			return failure
		}
		_, failure := transaction.Exec(ctx, `UPDATE comments SET body=$2,is_spoiler=$3,updated_at=now() WHERE id=$1`, id, item.Body, item.Spoiler)
		return failure
	})
}
func (repository *Repository) Delete(ctx context.Context, user, id string) error {
	return repository.write(ctx, user, false, func(transaction pgx.Tx) error {
		result, failure := transaction.Exec(ctx, `UPDATE comments SET body='',deleted_at=now(),updated_at=now() WHERE id=$1 AND user_id=$2`, id, user)
		if failure == nil && result.RowsAffected() == 0 {
			return domain.ErrForbidden
		}
		return failure
	})
}
func (repository *Repository) Rating(ctx context.Context, user, title string, score int) error {
	return repository.write(ctx, user, false, func(transaction pgx.Tx) error {
		if failure := available(ctx, transaction, title, ""); failure != nil {
			return failure
		}
		if score == 0 {
			_, failure := transaction.Exec(ctx, `DELETE FROM ratings WHERE user_id=$1 AND title_id=$2`, user, title)
			return failure
		}
		_, failure := transaction.Exec(ctx, `INSERT INTO ratings(id,user_id,title_id,score)VALUES($1,$2,$3,$4)ON CONFLICT(user_id,title_id)DO UPDATE SET score=EXCLUDED.score,updated_at=now()`, uuid.NewString(), user, title, score)
		return failure
	})
}
func (repository *Repository) MyRating(ctx context.Context, user, title string) (int, error) {
	var score int
	failure := repository.pool.QueryRow(ctx, `SELECT score FROM ratings WHERE user_id=$1 AND title_id=$2`, user, title).Scan(&score)
	if errors.Is(failure, pgx.ErrNoRows) {
		return 0, nil
	}
	return score, failure
}
func (repository *Repository) Report(ctx context.Context, user string, input application.Report) (string, error) {
	failure := repository.write(ctx, user, false, func(transaction pgx.Tx) error {
		if failure := available(ctx, transaction, input.TitleID, input.ReleaseID); failure != nil {
			return failure
		}
		var count int
		if failure := transaction.QueryRow(ctx, `SELECT count(*) FROM reports WHERE user_id=$1 AND created_at>now()-interval '1 hour'`, user).Scan(&count); failure != nil {
			return failure
		}
		if count >= 10 {
			return domain.ErrLimited
		}
		var valid bool
		switch input.TargetType {
		case "title":
			valid = input.TargetID == input.TitleID
		case "release":
			valid = input.TargetID == input.ReleaseID && input.ReleaseID != ""
		case "comment":
			if failure := transaction.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id=$1 AND title_id=$2 AND COALESCE(release_id,'')=$3)`, input.TargetID, input.TitleID, input.ReleaseID).Scan(&valid); failure != nil {
				return failure
			}
		case "page":
			if failure := transaction.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pages page JOIN releases release ON release.current_revision_id=page.revision_id WHERE release.id=$1 AND page.id=$2 AND release.status='published')`, input.ReleaseID, input.TargetID).Scan(&valid); failure != nil {
				return failure
			}
		default:
			return domain.ErrInvalid
		}
		if !valid {
			return domain.ErrNotFound
		}
		_, failure := transaction.Exec(ctx, `INSERT INTO reports(id,user_id,title_id,release_id,page_id,category,reason,target_type,target_id)VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),$6,$7,$8,$9)`, input.ID, user, input.TitleID, input.ReleaseID, input.PageID, input.Category, strings.TrimSpace(input.Reason), input.TargetType, input.TargetID)
		return failure
	})
	return input.ID, failure
}
func (repository *Repository) Reports(ctx context.Context, user, after string, filters ...string) ([]application.Report, error) {
	teamFilter, kind := "", ""
	if len(filters) > 0 {
		teamFilter = filters[0]
	}
	if len(filters) > 1 {
		kind = filters[1]
	}
	items := []application.Report{}
	failure := repository.write(ctx, user, true, func(transaction pgx.Tx) error {
		rows, failure := transaction.Query(ctx, `SELECT report.id,COALESCE(report.user_id,''),COALESCE(report.title_id,''),COALESCE(report.release_id,''),COALESCE(report.page_id,''),report.target_type,report.target_id,report.category,report.reason,report.status,report.resolution_note,report.created_at,report.contact_email,
 COALESCE(title.name_kk,''),COALESCE(title.slug,''),COALESCE(team.name,''),COALESCE(team.slug,''),COALESCE(chapter.label,''),COALESCE(release.status,''),COALESCE(page.revision_id,release.current_revision_id,''),COALESCE(page.position,0),
 COALESCE(comment.body,''),COALESCE(author.display_name,''),COALESCE(comment.is_hidden,false),comment.deleted_at IS NOT NULL,
 COALESCE((SELECT jsonb_agg(decision ORDER BY decision.created_at DESC) FROM (
 SELECT audit.action,COALESCE(audit.metadata->'after'->>'status','') AS status,COALESCE(staff.display_name,'') AS actor_name,COALESCE(audit.metadata->>'reason','') AS reason,audit.created_at
 FROM audit_logs audit LEFT JOIN users staff ON staff.id=audit.user_id
 WHERE audit.entity_type='community' AND (audit.entity_id=report.target_id OR audit.entity_id IN(SELECT previous.id FROM reports previous WHERE previous.target_type=report.target_type AND previous.target_id=report.target_id))
 ORDER BY audit.created_at DESC,audit.id DESC LIMIT 10) decision),'[]'::jsonb)
 FROM reports report LEFT JOIN titles title ON title.id=report.title_id
 LEFT JOIN releases release ON release.id=report.release_id LEFT JOIN chapters chapter ON chapter.id=release.chapter_id
 LEFT JOIN teams team ON team.id=release.team_id
 LEFT JOIN pages page ON page.id=CASE WHEN report.target_type='page' THEN report.target_id ELSE report.page_id END
 LEFT JOIN comments comment ON report.target_type='comment' AND comment.id=report.target_id LEFT JOIN users author ON author.id=comment.user_id
 WHERE report.status='pending' AND ($1::varchar='' OR report.id>$1::varchar) AND ($2='' OR team.id=$2 OR team.name ILIKE '%'||$2||'%') AND ($3='' OR report.target_type=$3) ORDER BY report.id LIMIT 50`, after, teamFilter, kind)
		if failure != nil {
			return failure
		}
		defer rows.Close()
		for rows.Next() {
			var item application.Report
			var history []byte
			if failure = rows.Scan(&item.ID, &item.UserID, &item.TitleID, &item.ReleaseID, &item.PageID, &item.TargetType, &item.TargetID, &item.Category, &item.Reason, &item.Status, &item.Resolution, &item.CreatedAt, &item.ContactEmail, &item.TitleName, &item.TitleSlug, &item.TeamName, &item.TeamSlug, &item.ChapterLabel, &item.ReleaseStatus, &item.RevisionID, &item.PagePosition, &item.CommentBody, &item.CommentAuthor, &item.CommentHidden, &item.CommentDeleted, &history); failure != nil {
				return failure
			}
			if failure = json.Unmarshal(history, &item.History); failure != nil {
				return failure
			}
			items = append(items, item)
		}
		return rows.Err()
	})
	return items, failure
}
func (repository *Repository) Resolve(ctx context.Context, user, id, status, reason string) error {
	return repository.write(ctx, user, true, func(transaction pgx.Tx) error {
		result, failure := transaction.Exec(ctx, `UPDATE reports SET status=$2,resolution_note=$3,resolved_at=now() WHERE id=$1 AND status='pending'`, id, status, reason)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return audit(ctx, transaction, user, "resolve_report", id, reason, map[string]any{"status": "pending"}, map[string]any{"status": status})
	})
}
func (repository *Repository) Visibility(ctx context.Context, user, id string, hidden bool, reason string) error {
	return repository.write(ctx, user, true, func(transaction pgx.Tx) error {
		var before bool
		if failure := transaction.QueryRow(ctx, `SELECT is_hidden FROM comments WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, id).Scan(&before); failure != nil {
			return domain.ErrNotFound
		}
		result, failure := transaction.Exec(ctx, `UPDATE comments SET is_hidden=$2,updated_at=now() WHERE id=$1`, id, hidden)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		action := "restore_comment"
		if hidden {
			action = "hide_comment"
		}
		return audit(ctx, transaction, user, action, id, reason, map[string]any{"hidden": before}, map[string]any{"hidden": hidden})
	})
}
func audit(ctx context.Context, transaction pgx.Tx, user, action, id, reason string, before, after map[string]any) error {
	metadata, _ := json.Marshal(map[string]any{"reason": reason, "before": before, "after": after})
	_, failure := transaction.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,$3,'community',$4,$5)`, uuid.NewString(), user, action, id, metadata)
	return failure
}
