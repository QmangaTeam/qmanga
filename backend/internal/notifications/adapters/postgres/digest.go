package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/url"
	"qmanga/internal/notifications/application"
	"strings"
	"time"
)

// A locked durable receipt prevents simultaneous delivery. SMTP acceptance and DB
// acknowledgement cannot be atomic: a crash between them may repeat a digest.
// Failure retries after an hour. Empty weeks advance the receipt without email.
func (r *Repository) DeliverDigest(ctx context.Context, sender application.DigestSender, origin string) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	_, err = tx.Exec(ctx, `INSERT INTO notification_digest_state(user_id) SELECT p.user_id FROM account_preferences p JOIN users u ON u.id=p.user_id AND u.is_verified AND NOT u.is_blocked WHERE p.email_digest AND p.digest_after<=now() ON CONFLICT DO NOTHING`)
	if err != nil {
		return false, err
	}
	var user, email string
	var since time.Time
	err = tx.QueryRow(ctx, `SELECT u.id,u.email,GREATEST(COALESCE(s.last_sent_at,now()-interval '7 days'),now()-interval '7 days') FROM notification_digest_state s JOIN account_preferences p ON p.user_id=s.user_id JOIN users u ON u.id=p.user_id WHERE p.email_digest AND p.digest_after<=now() AND u.is_verified AND NOT u.is_blocked AND s.available_at<=now() ORDER BY s.available_at,s.user_id FOR UPDATE OF s,p,u SKIP LOCKED LIMIT 1`).Scan(&user, &email, &since)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// Digest is based on visible published releases, current library choices and
	// durable publication events, not notification read status or invented activity.
	rows, err := tx.Query(ctx, `SELECT DISTINCT t.name_kk,c.label,t.slug FROM library_entries l JOIN titles t ON t.id=l.title_id JOIN chapters c ON c.title_id=t.id JOIN releases r ON r.chapter_id=c.id AND r.status='published' JOIN outbox_events e ON e.event_type='release_published' AND e.payload->>'release_id'=r.id AND e.payload->>'revision_id'=r.current_revision_id WHERE l.user_id=$1 AND NOT l.is_removed AND l.notify_new_releases AND (l.preferred_team_id IS NULL OR l.preferred_team_id=r.team_id) AND l.created_at<=e.created_at AND e.created_at>$2 AND e.created_at<=now() ORDER BY t.name_kk,c.label,t.slug LIMIT 100`, user, since)
	if err != nil {
		return true, err
	}
	var body strings.Builder
	for rows.Next() {
		var title, label, slug string
		if err = rows.Scan(&title, &label, &slug); err != nil {
			rows.Close()
			return true, err
		}
		fmt.Fprintf(&body, "%s · %s-тарау\n%s/titles/%s\n\n", title, label, strings.TrimRight(origin, "/"), url.PathEscape(slug))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return true, err
	}
	var sendErr error
	if body.Len() > 0 {
		sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		sendErr = sender.Send(sendCtx, email, "Qmanga · Апталық жаңа тараулар", body.String())
		cancel()
	}
	// A cancelled send is left unacknowledged and safely retried next time.
	if ctx.Err() != nil {
		return true, ctx.Err()
	}
	if sendErr != nil {
		_, err = tx.Exec(ctx, `UPDATE notification_digest_state SET available_at=now()+interval '1 hour' WHERE user_id=$1`, user)
	} else {
		_, err = tx.Exec(ctx, `UPDATE notification_digest_state SET last_sent_at=now(),available_at=now()+interval '7 days' WHERE user_id=$1`, user)
	}
	if err != nil {
		return true, err
	}
	if err = tx.Commit(ctx); err != nil {
		return true, err
	}
	return true, sendErr
}
