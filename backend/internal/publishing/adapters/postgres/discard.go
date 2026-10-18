package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/domain"
)

// DiscardDraft removes only an editable revision. Published history and shared
// chapter records remain, and immutable object bytes are left to retention.
func (repository *Repository) DiscardDraft(ctx context.Context, actor, id string, version int64) error {
	return repository.withRevision(ctx, actor, id, "submit", func(tx pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		if revision.Version() != version {
			return domain.ErrConflict
		}
		var status, current string
		if err := tx.QueryRow(ctx, `SELECT status,coalesce(current_revision_id,'') FROM releases WHERE id=$1 FOR UPDATE`, revision.ReleaseID()).Scan(&status, &current); err != nil {
			return err
		}
		if current == id {
			return domain.ErrImmutable
		}
		if _, err := tx.Exec(ctx, `UPDATE jobs SET status='cancelled',leased_until=NULL,updated_at=now() WHERE type='process_page' AND status IN ('pending','leased','failed') AND payload->>'page_id' IN (SELECT id FROM pages WHERE revision_id=$1)`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM revisions WHERE id=$1`, id); err != nil {
			return err
		}
		if status == "draft" && current == "" {
			if _, err := tx.Exec(ctx, `DELETE FROM releases WHERE id=$1 AND NOT EXISTS(SELECT 1 FROM revisions WHERE release_id=$1)`, revision.ReleaseID()); err != nil {
				return err
			}
		}
		metadata, _ := json.Marshal(map[string]any{"revision_id": id, "page_count": revision.PageCount()})
		_, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,'discard_draft','release',$3,$4)`, uuid.NewString(), actor, revision.ReleaseID(), metadata)
		return err
	})
}
