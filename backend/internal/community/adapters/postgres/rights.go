package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/community/application"
	"qmanga/internal/community/domain"
)

func receiptHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (repository *Repository) RightsReport(ctx context.Context, receipt application.ReportReceipt, input application.RightsRequest) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	if _, failure = transaction.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", input.Email); failure != nil {
		return failure
	}
	var count int
	if failure = transaction.QueryRow(ctx, "SELECT count(*) FROM reports WHERE contact_email=$1 AND created_at>now()-interval '1 day'", input.Email).Scan(&count); failure != nil {
		return failure
	}
	if count >= 5 {
		return domain.ErrLimited
	}
	var title string
	failure = transaction.QueryRow(ctx, "SELECT id FROM titles WHERE slug=$1", input.TitleSlug).Scan(&title)
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if failure != nil {
		return failure
	}
	if failure = available(ctx, transaction, title, ""); failure != nil {
		return failure
	}
	if _, failure = transaction.Exec(ctx, "INSERT INTO reports(id,title_id,category,reason,target_type,target_id,contact_email,receipt_hash)VALUES($1,$2,'rights',$3,'title',$2,$4,$5)", receipt.ID, title, input.Reason, input.Email, receiptHash(receipt.Token)); failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) RightsStatus(ctx context.Context, id, token string) (application.Report, error) {
	var item application.Report
	failure := repository.pool.QueryRow(ctx, "SELECT id,status,resolution_note,created_at FROM reports WHERE id=$1 AND user_id IS NULL AND receipt_hash=$2", id, receiptHash(token)).Scan(&item.ID, &item.Status, &item.Resolution, &item.CreatedAt)
	if errors.Is(failure, pgx.ErrNoRows) {
		return item, domain.ErrNotFound
	}
	return item, failure
}
func (repository *Repository) MyReports(ctx context.Context, user, after string) ([]application.Report, error) {
	items := []application.Report{}
	rows, failure := repository.pool.Query(ctx, "SELECT id,COALESCE(title_id,''),category,reason,status,resolution_note,created_at FROM reports WHERE user_id=$1 AND ($2='' OR (created_at,id)<(SELECT created_at,id FROM reports WHERE id=$2 AND user_id=$1)) ORDER BY created_at DESC,id DESC LIMIT 50", user, after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	for rows.Next() {
		var item application.Report
		if failure = rows.Scan(&item.ID, &item.TitleID, &item.Category, &item.Reason, &item.Status, &item.Resolution, &item.CreatedAt); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
