package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"time"
)

func (r *Repository) SaveGoogleFlow(ctx context.Context, f application.GoogleFlow) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM google_auth_flows WHERE expires_at < now()`)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO google_auth_flows(state_hash,nonce,verifier,return_path,nickname,expires_at) VALUES($1,$2,$3,$4,$5,$6)`, f.Hash, f.Nonce, f.Verifier, f.ReturnPath, f.Nickname, f.ExpiresAt)
	return err
}
func (r *Repository) ConsumeGoogleFlow(ctx context.Context, hash string, now time.Time) (application.GoogleFlow, error) {
	var f application.GoogleFlow
	err := r.pool.QueryRow(ctx, `DELETE FROM google_auth_flows WHERE state_hash=$1 AND expires_at>$2 RETURNING nonce,verifier,return_path,nickname,expires_at`, hash, now).Scan(&f.Nonce, &f.Verifier, &f.ReturnPath, &f.Nickname, &f.ExpiresAt)
	return f, translate(err)
}
func (r *Repository) GoogleAccount(ctx context.Context, identity application.GoogleIdentity, candidate domain.Account) (domain.Account, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Account{}, err
	}
	defer tx.Rollback(ctx)
	var a domain.Account
	err = tx.QueryRow(ctx, `SELECT u.id,u.email,u.password_hash,u.role,u.is_verified,u.display_name FROM external_identities i JOIN users u ON u.id=i.user_id WHERE i.issuer=$1 AND i.subject=$2 AND NOT u.is_blocked`, identity.Issuer, identity.Subject).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.Verified, &a.DisplayName)
	if err == nil {
		return a, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return a, err
	}
	// An email match never grants account access or links an external identity.
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=$1) OR EXISTS(SELECT 1 FROM external_identities WHERE issuer=$2 AND subject=$3)`, identity.Email, identity.Issuer, identity.Subject).Scan(&exists); err != nil {
		return a, err
	}
	if exists {
		return a, domain.ErrConflict
	}
	if candidate.DisplayName == "" {
		return a, application.ErrNicknameRequired
	}
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,role,is_verified,display_name) VALUES($1,$2,$3,'reader',true,$4)`, candidate.ID, candidate.Email, candidate.PasswordHash, candidate.DisplayName); err != nil {
		return a, translate(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO external_identities(issuer,subject,user_id) VALUES($1,$2,$3)`, identity.Issuer, identity.Subject, candidate.ID); err != nil {
		return a, translate(err)
	}
	return candidate, tx.Commit(ctx)
}
