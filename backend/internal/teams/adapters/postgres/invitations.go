package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/teams/application"
	"qmanga/internal/teams/domain"
)

func (repository *Repository) Invitations(ctx context.Context, actor, team string, offset ...int) ([]application.InvitationSummary, error) {
	access, failure := repository.Access(ctx, actor, team)
	if failure != nil {
		return nil, failure
	}
	if !access.Allows(domain.Manage) {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, `SELECT id,email,role,expires_at,CASE WHEN revoked_at IS NOT NULL THEN 'revoked' WHEN expires_at<=now() THEN 'expired' ELSE 'pending' END FROM team_invitations WHERE team_id=$1 ORDER BY created_at DESC,id LIMIT 100 OFFSET `+pageOffset(offset), team)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.InvitationSummary{}
	for rows.Next() {
		var item application.InvitationSummary
		if failure = rows.Scan(&item.ID, &item.Email, &item.Role, &item.ExpiresAt, &item.Status); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) RevokeInvitation(ctx context.Context, actor, team, id string) error {
	return repository.withAccess(ctx, actor, team, domain.Manage, func(transaction pgx.Tx) error {
		result, failure := transaction.Exec(ctx, "UPDATE team_invitations SET revoked_at=coalesce(revoked_at,now()) WHERE id=$1 AND team_id=$2", id, team)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		return nil
	})
}
