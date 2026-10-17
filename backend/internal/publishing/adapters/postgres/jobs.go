package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

func (repository *Repository) LeaseJob(ctx context.Context, claim string) (application.Job, bool, error) {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return application.Job{}, false, failure
	}
	defer transaction.Rollback(ctx)
	if _, failure = transaction.Exec(ctx, "UPDATE jobs SET status='failed',last_error='lease_exhausted',updated_at=now() WHERE type='process_page' AND status='leased' AND leased_until<now() AND attempts>=max_attempts"); failure != nil {
		return application.Job{}, false, failure
	}
	if _, failure = transaction.Exec(ctx, "UPDATE pages SET status='failed',failure_code='processing_failed' WHERE status='queued' AND EXISTS(SELECT 1 FROM jobs WHERE type='process_page' AND payload->>'page_id'=pages.id AND status='failed')"); failure != nil {
		return application.Job{}, false, failure
	}
	job := application.Job{ClaimID: claim}
	failure = transaction.QueryRow(ctx, `SELECT job.id,job.attempts,page.id,page.revision_id,page.status,page.image_format,page.source_key,page.checksum,page.position,page.width,page.height,page.byte_size FROM jobs job JOIN pages page ON page.id=job.payload->>'page_id' JOIN revisions revision ON revision.id=page.revision_id WHERE job.type='process_page' AND job.attempts<job.max_attempts AND ((job.status='pending' AND job.available_at<=now())OR(job.status='leased' AND job.leased_until<now())) AND revision.status IN('draft','processing','ready') ORDER BY job.created_at FOR UPDATE OF job SKIP LOCKED LIMIT 1`).Scan(&job.ID, &job.Attempts, &job.Page.ID, &job.Page.RevisionID, &job.Page.Status, &job.Page.Format, &job.Page.SourceKey, &job.Page.Checksum, &job.Page.Position, &job.Page.Width, &job.Page.Height, &job.Page.Bytes)
	if errors.Is(failure, pgx.ErrNoRows) {
		return job, false, transaction.Commit(ctx)
	}
	if failure != nil {
		return job, false, failure
	}
	job.PageID = job.Page.ID
	job.RevisionID = job.Page.RevisionID
	job.Attempts++
	if _, failure = transaction.Exec(ctx, "UPDATE jobs SET status='leased',lease_token=$2,leased_until=now()+interval '5 minutes',attempts=attempts+1,updated_at=now() WHERE id=$1", job.ID, claim); failure != nil {
		return job, false, failure
	}
	return job, true, transaction.Commit(ctx)
}
func (repository *Repository) CompleteJob(ctx context.Context, job application.Job, variants []application.Variant) error {
	if failure := application.ValidateVariants(job.Page, variants); failure != nil {
		return failure
	}
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	revision, failure := lockRevision(ctx, transaction, job.RevisionID)
	if failure != nil {
		return failure
	}
	if revision.State() != domain.Processing {
		return domain.ErrImmutable
	}
	var locked string
	if failure = transaction.QueryRow(ctx, "SELECT id FROM jobs WHERE id=$1 AND lease_token=$2 AND status='leased' AND leased_until>now() FOR UPDATE", job.ID, job.ClaimID).Scan(&locked); failure != nil {
		return domain.ErrConflict
	}
	for _, variant := range variants {
		_, failure = transaction.Exec(ctx, "INSERT INTO page_variants(id,page_id,tier,object_key,width,height,byte_size,offset_y,checksum,mime)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", variant.ID, job.PageID, variant.Tier, variant.ObjectKey, variant.Width, variant.Height, variant.Bytes, variant.Top, variant.Checksum, variant.Mime)
		if failure != nil {
			return failure
		}
	}
	if failure = revision.PageReady(); failure != nil {
		return failure
	}
	if _, failure = transaction.Exec(ctx, "UPDATE pages SET status='ready',failure_code='' WHERE id=$1", job.PageID); failure != nil {
		return failure
	}
	if failure = persist(ctx, transaction, revision); failure != nil {
		return failure
	}
	if _, failure = transaction.Exec(ctx, "UPDATE jobs SET status='done',leased_until=NULL,updated_at=now() WHERE id=$1", job.ID); failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) FailJob(ctx context.Context, job application.Job, code string) error {
	_, failure := repository.pool.Exec(ctx, `UPDATE jobs SET status=CASE WHEN attempts>=max_attempts THEN 'failed' ELSE 'pending' END,available_at=now()+interval '10 seconds'*attempts,last_error=$3,leased_until=NULL,updated_at=now() WHERE id=$1 AND lease_token=$2 AND status='leased'`, job.ID, job.ClaimID, code)
	return failure
}
