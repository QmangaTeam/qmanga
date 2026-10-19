package postgres

import (
	"context"
	"qmanga/internal/identity/domain"
	"time"
)

func (repository *Repository) CreateAccountForImport(ctx context.Context, account domain.Account, hash string, expires time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,role,is_verified,display_name) VALUES($1,$2,$3,$4,$5,$6)`, account.ID, account.Email, account.PasswordHash, account.Role, account.Verified, account.DisplayName); err != nil {
		return translate(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM registration_reading_imports WHERE expires_at<now()`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO registration_reading_imports(user_id,receipt_hash,expires_at) VALUES($1,$2,$3)`, account.ID, hash, expires); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (repository *Repository) RegistrationImportAllowed(ctx context.Context, actor, hash string, now time.Time) (bool, error) {
	var allowed bool
	err := repository.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registration_reading_imports WHERE user_id=$1 AND receipt_hash=$2 AND expires_at>$3 AND completed_at IS NULL)`, actor, hash, now).Scan(&allowed)
	return allowed, err
}
func (repository *Repository) CompleteRegistrationImport(ctx context.Context, actor, hash string, now time.Time) error {
	result, err := repository.pool.Exec(ctx, `UPDATE registration_reading_imports SET completed_at=coalesce(completed_at,$3) WHERE user_id=$1 AND receipt_hash=$2 AND expires_at>$3`, actor, hash, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return domain.ErrForbidden
	}
	return nil
}
