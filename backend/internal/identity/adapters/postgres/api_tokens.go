package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"time"
)

func (r *Repository) CreateAPIToken(ctx context.Context, t application.APIToken) error {
	result, e := r.pool.Exec(ctx, `INSERT INTO api_tokens(id,user_id,name,prefix,token_hash,created_at,expires_at,direct_publish) SELECT $1,id,$3,$4,$5,$6,$7,$8 FROM users WHERE id=$2 AND role='admin' AND is_verified AND NOT is_blocked`, t.ID, t.UserID, t.Name, t.Prefix, t.Hash, t.CreatedAt, t.ExpiresAt, t.DirectPublish)
	if e != nil {
		return translate(e)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrForbidden
	}
	return nil
}
func (r *Repository) ListAPITokens(ctx context.Context, owner string) ([]application.APIToken, error) {
	rows, e := r.pool.Query(ctx, `SELECT t.id,t.name,t.prefix,t.created_at,t.expires_at,t.last_used_at,t.revoked_at,t.direct_publish FROM api_tokens t JOIN users u ON u.id=t.user_id WHERE u.id=$1 AND u.role='admin' AND u.is_verified AND NOT u.is_blocked ORDER BY t.created_at DESC,t.id`, owner)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	items := []application.APIToken{}
	for rows.Next() {
		var t application.APIToken
		if e = rows.Scan(&t.ID, &t.Name, &t.Prefix, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt, &t.RevokedAt, &t.DirectPublish); e != nil {
			return nil, e
		}
		items = append(items, t)
	}
	return items, rows.Err()
}
func (r *Repository) RevokeAPIToken(ctx context.Context, owner, id string, now time.Time) error {
	result, e := r.pool.Exec(ctx, `UPDATE api_tokens t SET revoked_at=COALESCE(t.revoked_at,$3) FROM users u WHERE t.id=$2 AND t.user_id=$1 AND u.id=t.user_id AND u.role='admin' AND u.is_verified AND NOT u.is_blocked`, owner, id, now)
	if e != nil {
		return e
	}
	if result.RowsAffected() != 1 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Repository) AuthenticateAPIToken(ctx context.Context, hash string, now time.Time) (application.APIToken, domain.Account, error) {
	var t application.APIToken
	var a domain.Account
	e := r.pool.QueryRow(ctx, `UPDATE api_tokens t SET last_used_at=$2 FROM users u WHERE t.token_hash=$1 AND t.revoked_at IS NULL AND t.expires_at>$2 AND u.id=t.user_id AND u.role='admin' AND u.is_verified AND NOT u.is_blocked RETURNING t.id,t.user_id,t.direct_publish,t.expires_at,u.id,u.email,u.role,u.is_verified,u.display_name`, hash, now).Scan(&t.ID, &t.UserID, &t.DirectPublish, &t.ExpiresAt, &a.ID, &a.Email, &a.Role, &a.Verified, &a.DisplayName)
	if errors.Is(e, pgx.ErrNoRows) {
		e = domain.ErrCredentials
	}
	return t, a, e
}
