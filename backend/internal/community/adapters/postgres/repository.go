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

const visibleTitle = `EXISTS(SELECT 1 FROM chapters chapter JOIN releases release ON release.chapter_id=chapter.id JOIN revisions revision ON revision.id=release.current_revision_id JOIN projects project ON project.id=release.project_id AND project.status='active' JOIN teams team ON team.id=release.team_id AND team.is_approved WHERE chapter.title_id=$1 AND release.status='published' AND revision.status='published')`

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
	if failure := transaction.QueryRow(ctx, `SELECT is_verified,role FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&verified, &role); failure != nil {
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
func (repository *Repository) Discussion(ctx context.Context, title, release, after string) (application.Discussion, error) {
	result := application.Discussion{Comments: []application.Comment{}}
	var valid bool
	if failure := repository.pool.QueryRow(ctx, `SELECT `+visibleTitle, title).Scan(&valid); failure != nil {
		return result, failure
	}
	if !valid {
		return result, domain.ErrNotFound
	}
	if failure := repository.pool.QueryRow(ctx, `SELECT COALESCE(avg(score)::float8,0),count(*) FROM ratings WHERE title_id=$1`, title).Scan(&result.Score, &result.Ratings); failure != nil {
		return result, failure
	}
	rows, failure := repository.pool.Query(ctx, `WITH roots AS(
	 SELECT id,created_at FROM comments WHERE title_id=$1 AND COALESCE(release_id,'')=$2 AND parent_id IS NULL
	 AND ($3::varchar='' OR (created_at,id)<(SELECT created_at,id FROM comments WHERE id=$3::varchar AND title_id=$1 AND COALESCE(release_id,'')=$2))
	 ORDER BY created_at DESC,id DESC LIMIT 20
	), selected AS(
	 SELECT id FROM roots UNION ALL SELECT reply.id FROM roots CROSS JOIN LATERAL(
	 SELECT id FROM comments WHERE parent_id=roots.id ORDER BY created_at,id LIMIT 10)reply
	)
	SELECT comment.id,comment.user_id,comment.title_id,COALESCE(comment.release_id,''),COALESCE(comment.parent_id,''),
	CASE WHEN comment.is_hidden OR comment.deleted_at IS NOT NULL THEN '' ELSE comment.body END,
	comment.is_spoiler,comment.is_hidden OR comment.deleted_at IS NOT NULL,comment.created_at,
	(SELECT count(*) FROM comments reply WHERE reply.parent_id=comment.id),(SELECT display_name FROM users WHERE id=comment.user_id)
	FROM comments comment WHERE comment.id IN(SELECT id FROM selected) ORDER BY comment.created_at,comment.id`, title, release, after)
	if failure != nil {
		return result, failure
	}
	defer rows.Close()
	rootCount := 0
	var oldest application.Comment
	for rows.Next() {
		var item application.Comment
		if failure = rows.Scan(&item.ID, &item.UserID, &item.TitleID, &item.ReleaseID, &item.ParentID, &item.Body, &item.Spoiler, &item.Hidden, &item.CreatedAt, &item.Replies, &item.DisplayName); failure != nil {
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
		var replyUser string
		if input.ParentID != "" {
			var title, release, parent string
			var hidden bool
			failure := transaction.QueryRow(ctx, `SELECT title_id,COALESCE(release_id,''),COALESCE(parent_id,''),user_id,is_hidden OR deleted_at IS NOT NULL FROM comments WHERE id=$1 FOR SHARE`, input.ParentID).Scan(&title, &release, &parent, &replyUser, &hidden)
			if failure != nil || hidden || !domain.CanReply(input.TitleID, input.ReleaseID, title, release, parent != "") {
				return domain.ErrInvalid
			}
		}
		if _, failure := transaction.Exec(ctx, `INSERT INTO comments(id,user_id,title_id,release_id,parent_id,body,is_spoiler)VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),$6,$7)`, input.ID, user, input.TitleID, input.ReleaseID, input.ParentID, strings.TrimSpace(input.Body), input.Spoiler); failure != nil {
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
func (repository *Repository) Reports(ctx context.Context, user, after string) ([]application.Report, error) {
	items := []application.Report{}
	failure := repository.write(ctx, user, true, func(transaction pgx.Tx) error {
		rows, failure := transaction.Query(ctx, `SELECT id,COALESCE(user_id,''),COALESCE(title_id,''),COALESCE(release_id,''),COALESCE(page_id,''),target_type,target_id,category,reason,status,resolution_note,created_at,contact_email FROM reports WHERE status='pending' AND ($1::varchar='' OR id>$1::varchar)ORDER BY id LIMIT 50`, after)
		if failure != nil {
			return failure
		}
		defer rows.Close()
		for rows.Next() {
			var item application.Report
			if failure = rows.Scan(&item.ID, &item.UserID, &item.TitleID, &item.ReleaseID, &item.PageID, &item.TargetType, &item.TargetID, &item.Category, &item.Reason, &item.Status, &item.Resolution, &item.CreatedAt, &item.ContactEmail); failure != nil {
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
		return audit(ctx, transaction, user, "resolve_report", id, reason)
	})
}
func (repository *Repository) Visibility(ctx context.Context, user, id string, hidden bool, reason string) error {
	return repository.write(ctx, user, true, func(transaction pgx.Tx) error {
		result, failure := transaction.Exec(ctx, `UPDATE comments SET is_hidden=$2,updated_at=now() WHERE id=$1`, id, hidden)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return audit(ctx, transaction, user, "comment_visibility", id, reason)
	})
}
func audit(ctx context.Context, transaction pgx.Tx, user, action, id, reason string) error {
	metadata, _ := json.Marshal(map[string]string{"reason": reason})
	_, failure := transaction.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,$3,'community',$4,$5)`, uuid.NewString(), user, action, id, metadata)
	return failure
}
