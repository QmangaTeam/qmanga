package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/notifications/application"
	"time"
)

// The transaction lock spans send + acknowledgement so concurrent workers never
// send the same item concurrently. Delivery is at least once after a crash.
func (r *Repository) DeliverModeration(ctx context.Context, sender application.ModerationSender) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var message application.ModerationMessage
	err = tx.QueryRow(ctx, `SELECT id,kind,entity_id,label,detail,path FROM moderation_telegram_queue WHERE delivered_at IS NULL AND available_at<=now() ORDER BY available_at,id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&message.ID, &message.Kind, &message.EntityID, &message.Label, &message.Detail, &message.Path)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	sendErr := sender.Send(sendCtx, message)
	cancel()
	if sendErr != nil {
		_, err = tx.Exec(ctx, `UPDATE moderation_telegram_queue SET attempts=attempts+1,available_at=now()+make_interval(secs=>LEAST(3600,30*power(2,LEAST(attempts,7)))::int) WHERE id=$1`, message.ID)
	} else {
		_, err = tx.Exec(ctx, `UPDATE moderation_telegram_queue SET delivered_at=now(),attempts=attempts+1 WHERE id=$1`, message.ID)
	}
	if err != nil {
		return true, err
	}
	if err = tx.Commit(ctx); err != nil {
		return true, err
	}
	return true, sendErr
}
