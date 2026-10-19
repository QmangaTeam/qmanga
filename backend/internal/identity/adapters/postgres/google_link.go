package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"time"
)

func (r *Repository) GoogleLinked(ctx context.Context, actor string) (bool, error) {
	var linked bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM external_identities WHERE issuer='https://accounts.google.com' AND user_id=$1)`, actor).Scan(&linked)
	return linked, err
}
func (r *Repository) LinkGoogleIdentity(ctx context.Context, actor, session string, identity application.GoogleIdentity, now time.Time) error {
	email, err := domain.NormalizeEmail(identity.Email)
	if err != nil || email != identity.Email || !identity.Verified || identity.Issuer != "https://accounts.google.com" || identity.Subject == "" {
		return domain.ErrCredentials
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	// Serialize against account blocking, reset/revocation and logout until the identity is committed.
	err = tx.QueryRow(ctx, `SELECT u.id FROM users u JOIN sessions s ON s.user_id=u.id WHERE u.id=$1 AND s.id=$2 AND s.expires_at>GREATEST($3,clock_timestamp()) AND u.is_verified AND NOT u.is_blocked FOR UPDATE OF u,s`, actor, session, now).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	if err != nil {
		return err
	}
	var owner string
	err = tx.QueryRow(ctx, `SELECT user_id FROM external_identities WHERE issuer=$1 AND subject=$2`, identity.Issuer, identity.Subject).Scan(&owner)
	if err == nil {
		if owner != actor {
			return domain.ErrConflict
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	var linked string
	err = tx.QueryRow(ctx, `INSERT INTO external_identities(issuer,subject,user_id) SELECT $1,$2,u.id FROM users u JOIN sessions s ON s.user_id=u.id WHERE u.id=$3 AND s.id=$4 AND s.expires_at>GREATEST($5,clock_timestamp()) AND u.is_verified AND NOT u.is_blocked RETURNING user_id`, identity.Issuer, identity.Subject, actor, session, now).Scan(&linked)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	if err != nil {
		return translate(err)
	}
	return tx.Commit(ctx)
}
