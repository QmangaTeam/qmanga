package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/domain"
)

func (repository *Repository) SaveAttribution(ctx context.Context, actor, id string, version int64, input domain.Attribution) error {
	if !input.Valid() {
		return domain.ErrTransition
	}
	return repository.withRevision(ctx, actor, id, "upload", func(transaction pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		if revision.Version() != version {
			return domain.ErrConflict
		}
		_, failure := transaction.Exec(ctx, "UPDATE revisions SET source_note=$2,credits=$3,rights_confirmed=true,rights_confirmed_by=$4,version=version+1,updated_at=now() WHERE id=$1", id, input.Source, input.Credits, actor)
		return failure
	})
}
