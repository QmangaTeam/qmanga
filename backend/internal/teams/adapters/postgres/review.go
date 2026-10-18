package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/teams/domain"
)

func (repository *Repository) Reject(ctx context.Context, actor, id, reason string) error {
	return repository.withAccess(ctx, actor, id, domain.Approve, func(transaction pgx.Tx) error {
		result, failure := transaction.Exec(ctx, "UPDATE teams SET review_status='rejected',review_note=$2 WHERE id=$1 AND NOT is_approved AND review_status='pending'", id, reason)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		_, failure = transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,'reject','team',$3,jsonb_build_object('reason',$4::text))", uuid.NewString(), actor, id, reason)
		return failure
	})
}
func (repository *Repository) Update(ctx context.Context, actor string, team domain.Team) error {
	return repository.withAccess(ctx, actor, team.ID, domain.Manage, func(transaction pgx.Tx) error {
		_, failure := transaction.Exec(ctx, "UPDATE teams SET slug=$2,name=$3,description=$4,telegram_url=$5,review_status=CASE WHEN is_approved THEN 'approved' ELSE 'pending' END,review_note='' WHERE id=$1", team.ID, team.Slug, team.Name, team.Description, team.TelegramURL)
		return failure
	})
}
