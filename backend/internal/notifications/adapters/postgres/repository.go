package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/notifications/application"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }
func (repository *Repository) List(ctx context.Context, user, after string) ([]application.Notification, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT id,type,title_kk,body_kk,COALESCE(data->>'release_id',''),COALESCE(data->>'title_slug',''),is_read,created_at FROM notifications WHERE user_id=$1 AND ($2::varchar='' OR (created_at,id)<(SELECT created_at,id FROM notifications WHERE id=$2::varchar AND user_id=$1)) ORDER BY created_at DESC,id DESC LIMIT 50`, user, after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Notification{}
	for rows.Next() {
		var item application.Notification
		if failure = rows.Scan(&item.ID, &item.Kind, &item.Title, &item.Body, &item.ReleaseID, &item.TitleSlug, &item.Read, &item.CreatedAt); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) Read(ctx context.Context, user, id string) error {
	_, failure := repository.pool.Exec(ctx, `UPDATE notifications SET is_read=true WHERE user_id=$1 AND id=$2`, user, id)
	return failure
}
func (repository *Repository) Deliver(ctx context.Context) (bool, error) {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return false, failure
	}
	defer transaction.Rollback(ctx)
	var id, kind string
	failure = transaction.QueryRow(ctx, `SELECT id,event_type FROM outbox_events WHERE status='pending' AND event_type IN('release_published','comment_reply')ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &kind)
	if errors.Is(failure, pgx.ErrNoRows) {
		return false, nil
	}
	if failure != nil {
		return false, failure
	}
	if kind == "release_published" {
		_, failure = transaction.Exec(ctx, `INSERT INTO notifications(id,user_id,type,title_kk,body_kk,data,event_id,dedupe_key) SELECT gen_random_uuid()::varchar,library.user_id,'new_release','Жаңа тарау',title.name_kk||' · '||chapter.label||'-тарау',jsonb_build_object('release_id',release.id,'title_slug',title.slug),event.id,'release:'||release.id FROM outbox_events event JOIN releases release ON release.id=event.payload->>'release_id' AND release.current_revision_id=event.payload->>'revision_id' AND release.status='published' JOIN chapters chapter ON chapter.id=release.chapter_id JOIN titles title ON title.id=chapter.title_id JOIN library_entries library ON library.title_id=title.id AND library.notify_new_releases AND library.created_at<=event.created_at AND (library.preferred_team_id IS NULL OR library.preferred_team_id=release.team_id) WHERE event.id=$1 ON CONFLICT(user_id,dedupe_key)DO NOTHING`, id)
	} else {
		_, failure = transaction.Exec(ctx, `INSERT INTO notifications(id,user_id,type,title_kk,body_kk,data,event_id,dedupe_key)SELECT gen_random_uuid()::varchar,recipient.id,'comment_reply','Пікіріңізге жауап берілді',title.name_kk,jsonb_build_object('release_id',COALESCE(event.payload->>'release_id',''),'title_slug',title.slug),event.id,'reply:'||(event.payload->>'comment_id') FROM outbox_events event JOIN users recipient ON recipient.id=event.payload->>'user_id' JOIN titles title ON title.id=event.payload->>'title_id' JOIN comments comment ON comment.id=event.payload->>'comment_id' AND NOT comment.is_hidden AND comment.deleted_at IS NULL WHERE event.id=$1 ON CONFLICT(user_id,dedupe_key)DO NOTHING`, id)
	}
	if failure != nil {
		return false, failure
	}
	if _, failure = transaction.Exec(ctx, `UPDATE outbox_events SET status='sent',sent_at=now() WHERE id=$1`, id); failure != nil {
		return false, failure
	}
	return true, transaction.Commit(ctx)
}
