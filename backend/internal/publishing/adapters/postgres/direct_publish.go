package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/domain"
)

// DirectPublish never persists the intermediate submitted state. All access,
// page readiness, attribution and publication effects share this transaction.
func (repository *Repository) DirectPublish(ctx context.Context, actor, tokenID, id string, version int64) error {
	return repository.withRevision(ctx, actor, id, "submit", func(tx pgx.Tx, revision domain.Revision) error {
		// withRevision already holds the actor row FOR SHARE and the team FOR UPDATE.
		// Hold the token against revocation and recheck every privilege under locks.
		var locked string
		if err := lockDirectToken(ctx, tx, actor, tokenID); err != nil {
			return err
		}

		// Prevent project suspension racing the authorization snapshot.
		err := tx.QueryRow(ctx, `SELECT project.id FROM projects project JOIN releases release ON release.project_id=project.id WHERE release.id=$1 AND project.status='active' FOR SHARE OF project`, revision.ReleaseID()).Scan(&locked)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		if err != nil {
			return err
		}
		if revision.State() == domain.Published {
			// A lost-response retry must not restore a withdrawn/superseded release.
			err = tx.QueryRow(ctx, `SELECT id FROM releases WHERE id=$1 AND current_revision_id=$2 AND status='published' FOR UPDATE`, revision.ReleaseID(), id).Scan(&locked)
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrConflict
			}
			return err
		}
		if revision.Version() != version {
			return domain.ErrConflict
		}
		var confirmed bool
		if err = tx.QueryRow(ctx, `SELECT rights_confirmed AND length(trim(source_note))>=8 AND length(trim(credits))>=2 AND rights_confirmed_by IS NOT NULL FROM revisions WHERE id=$1`, id).Scan(&confirmed); err != nil {
			return err
		}
		if !confirmed {
			return domain.ErrTransition
		}
		if err = revision.Submit(); err != nil {
			return err
		}
		if err = publishRevision(ctx, tx, actor, revision, "Direct publication by API token"); err != nil {
			return err
		}
		metadata, _ := json.Marshal(map[string]string{"api_token_id": tokenID, "revision_id": id})
		_, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,'direct_publish','release',$3,$4)`, uuid.NewString(), actor, revision.ReleaseID(), metadata)
		return err
	})
}

// The caller holds the actor row FOR SHARE before acquiring this token lock.
func lockDirectToken(ctx context.Context, tx pgx.Tx, actor, tokenID string) error {
	var locked string
	err := tx.QueryRow(ctx, `SELECT token.id FROM api_tokens token JOIN users account ON account.id=token.user_id WHERE token.id=$1 AND token.user_id=$2 AND token.direct_publish AND token.revoked_at IS NULL AND token.expires_at>clock_timestamp() AND account.role='admin' AND account.is_verified AND NOT account.is_blocked FOR SHARE OF token`, tokenID, actor).Scan(&locked)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	if err != nil {
		return err
	}
	return nil
}

// SelectCoverDirect snapshots a ready page derivative into a dedicated cover.
// Session and ordinary bearer callers never reach this privileged operation.
func (repository *Repository) SelectCoverDirect(ctx context.Context, actor, tokenID, project, page string) error {
	if page == "" {
		return domain.ErrTransition
	}
	return repository.withProject(ctx, actor, project, "submit", func(tx pgx.Tx) error {
		if err := lockDirectToken(ctx, tx, actor, tokenID); err != nil {
			return err
		}
		var locked string
		err := tx.QueryRow(ctx, `SELECT id FROM projects WHERE id=$1 AND status='active' FOR SHARE`, project).Scan(&locked)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		if err != nil {
			return err
		}
		if err = selectCover(ctx, tx, actor, project, page); err != nil {
			return err
		}
		metadata, _ := json.Marshal(map[string]string{"api_token_id": tokenID, "page_id": page})
		_, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,'direct_select_cover','project',$3,$4)`, uuid.NewString(), actor, project, metadata)
		return err
	})
}
