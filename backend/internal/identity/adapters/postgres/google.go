package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"time"
)

func (r *Repository) SaveGoogleFlow(ctx context.Context, f application.GoogleFlow) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM google_auth_flows WHERE expires_at < now()`)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO google_auth_flows(state_hash,nonce,verifier,return_path,nickname,expires_at,link_user_id,link_session_hash) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''))`, f.Hash, f.Nonce, f.Verifier, f.ReturnPath, f.Nickname, f.ExpiresAt, f.LinkUserID, f.LinkSessionHash)
	return err
}
func (r *Repository) ConsumeGoogleFlow(ctx context.Context, hash string, now time.Time) (application.GoogleFlow, error) {
	var f application.GoogleFlow
	err := r.pool.QueryRow(ctx, `DELETE FROM google_auth_flows WHERE state_hash=$1 AND expires_at>$2 RETURNING nonce,verifier,return_path,nickname,expires_at,COALESCE(link_user_id,''),COALESCE(link_session_hash,'')`, hash, now).Scan(&f.Nonce, &f.Verifier, &f.ReturnPath, &f.Nickname, &f.ExpiresAt, &f.LinkUserID, &f.LinkSessionHash)
	return f, translate(err)
}
func (r *Repository) GoogleAccount(ctx context.Context, identity application.GoogleIdentity, candidate domain.Account) (domain.Account, error) {
	return r.googleAccount(ctx, identity, candidate, "", time.Time{})
}
func (r *Repository) GoogleAccountForImport(ctx context.Context, identity application.GoogleIdentity, candidate domain.Account, receiptHash string, expires time.Time) (domain.Account, error) {
	return r.googleAccount(ctx, identity, candidate, receiptHash, expires)
}
func (r *Repository) googleAccount(ctx context.Context, identity application.GoogleIdentity, candidate domain.Account, receiptHash string, expires time.Time) (domain.Account, error) {
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
	// A verified Google email links a password-registered account and signs it in
	// (owner decision 19 October 2026, TASK-62): profile, role, ID and edits stay untouched.
	var member domain.Account
	var blocked bool
	found := tx.QueryRow(ctx, `SELECT id,email,password_hash,role,is_verified,is_blocked,display_name FROM users WHERE lower(email)=$1 FOR UPDATE`, identity.Email).Scan(&member.ID, &member.Email, &member.PasswordHash, &member.Role, &member.Verified, &blocked, &member.DisplayName)
	if errors.Is(found, pgx.ErrNoRows) {
		if candidate.DisplayName == "" {
			return a, application.ErrNicknameRequired
		}
		if _, err = tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,role,is_verified,display_name) VALUES($1,$2,$3,'reader',true,$4)`, candidate.ID, candidate.Email, candidate.PasswordHash, candidate.DisplayName); err != nil {
			return a, translate(err)
		}
	} else if found != nil {
		return a, found
	} else if blocked {
		return a, domain.ErrConflict
	}
	user := member
	if member.ID == "" {
		user = candidate
	}
	if _, err = tx.Exec(ctx, `INSERT INTO external_identities(issuer,subject,user_id) VALUES($1,$2,$3)`, identity.Issuer, identity.Subject, user.ID); err != nil {
		var unique *pgconn.PgError
		if errors.As(err, &unique) && unique.Code == "23505" && user.ID == member.ID {
			// A concurrent login may have linked or signed this identity in; read the bound row.
			if bound := tx.QueryRow(ctx, `SELECT u.id,u.email,u.password_hash,u.role,u.is_verified,u.display_name FROM external_identities i JOIN users u ON u.id=i.user_id WHERE i.issuer=$1 AND i.subject=$2 AND NOT u.is_blocked`, identity.Issuer, identity.Subject).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.Verified, &a.DisplayName); bound == nil {
				return a, tx.Commit(ctx)
			}
		}
		return user, translate(err)
	}
	if user.ID == candidate.ID && receiptHash != "" {
		if _, err = tx.Exec(ctx, `DELETE FROM registration_reading_imports WHERE expires_at<now()`); err != nil {
			return a, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO registration_reading_imports(user_id,receipt_hash,expires_at) VALUES($1,$2,$3)`, user.ID, receiptHash, expires); err != nil {
			return a, err
		}
	}
	return user, tx.Commit(ctx)
}
