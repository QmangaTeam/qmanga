package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }
func translate(failure error) error {
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(failure, &postgresError) && postgresError.Code == "23505" {
		return domain.ErrConflict
	}
	return failure
}
func (repository *Repository) CreateAccount(ctx context.Context, account domain.Account) error {
	_, failure := repository.pool.Exec(ctx, "INSERT INTO users(id,email,password_hash,role,is_verified) VALUES($1,$2,$3,$4,$5)", account.ID, account.Email, account.PasswordHash, account.Role, account.Verified)
	return translate(failure)
}
func (repository *Repository) AccountByEmail(ctx context.Context, email string) (domain.Account, error) {
	var account domain.Account
	failure := repository.pool.QueryRow(ctx, "SELECT id,email,password_hash,role,is_verified FROM users WHERE lower(email)=$1", email).Scan(&account.ID, &account.Email, &account.PasswordHash, &account.Role, &account.Verified)
	return account, translate(failure)
}
func (repository *Repository) SaveSession(ctx context.Context, session application.Session) error {
	_, failure := repository.pool.Exec(ctx, "INSERT INTO sessions(id,user_id,csrf_hash,expires_at) VALUES($1,$2,$3,$4)", session.Hash, session.UserID, session.CSRFHash, session.ExpiresAt)
	return failure
}
func (repository *Repository) SessionAccount(ctx context.Context, hash string, now time.Time) (application.Session, domain.Account, error) {
	var session application.Session
	var account domain.Account
	failure := repository.pool.QueryRow(ctx, `SELECT session.id,session.user_id,session.csrf_hash,session.expires_at,account.id,account.email,account.password_hash,account.role,account.is_verified
 FROM sessions session JOIN users account ON account.id=session.user_id WHERE session.id=$1 AND session.expires_at>$2`, hash, now).Scan(
		&session.Hash, &session.UserID, &session.CSRFHash, &session.ExpiresAt, &account.ID, &account.Email, &account.PasswordHash, &account.Role, &account.Verified)
	if errors.Is(failure, pgx.ErrNoRows) {
		return session, account, domain.ErrCredentials
	}
	return session, account, failure
}
func (repository *Repository) DeleteSession(ctx context.Context, hash string) error {
	_, failure := repository.pool.Exec(ctx, "DELETE FROM sessions WHERE id=$1", hash)
	return failure
}
func (repository *Repository) SaveToken(ctx context.Context, token application.Token) error {
	_, failure := repository.pool.Exec(ctx, `INSERT INTO tokens(id,user_id,token_type,expires_at) VALUES($1,$2,$3,$4)
 ON CONFLICT(user_id,token_type) DO UPDATE SET id=excluded.id,expires_at=excluded.expires_at,created_at=now()`, token.Hash, token.UserID, token.Kind, token.ExpiresAt)
	return failure
}
func (repository *Repository) ConsumeToken(ctx context.Context, hash, kind string, now time.Time, passwordHash string) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var userID string
	failure = transaction.QueryRow(ctx, "DELETE FROM tokens WHERE id=$1 AND token_type=$2 AND expires_at>$3 RETURNING user_id", hash, kind, now).Scan(&userID)
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrExpired
	}
	if failure != nil {
		return failure
	}
	if kind == "verify" {
		_, failure = transaction.Exec(ctx, "UPDATE users SET is_verified=true,updated_at=now() WHERE id=$1", userID)
	} else if kind == "reset" {
		_, failure = transaction.Exec(ctx, "UPDATE users SET password_hash=$2,updated_at=now() WHERE id=$1", userID, passwordHash)
		if failure == nil {
			_, failure = transaction.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", userID)
		}
	} else {
		return domain.ErrInvalid
	}
	if failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
