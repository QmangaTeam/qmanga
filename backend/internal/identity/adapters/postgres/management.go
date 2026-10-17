package postgres

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
)

func (repository *Repository) staff(ctx context.Context, actor string) (string, error) {
	var role string
	if failure := repository.pool.QueryRow(ctx, "SELECT role FROM users WHERE id=$1 AND is_verified AND NOT is_blocked AND role IN('admin','moderator')", actor).Scan(&role); failure != nil {
		return "", domain.ErrForbidden
	}
	return role, nil
}
func (repository *Repository) Accounts(ctx context.Context, actor, query, after string) ([]application.ManagedAccount, error) {
	if _, failure := repository.staff(ctx, actor); failure != nil {
		return nil, failure
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,display_name,role,is_blocked,management_version FROM users WHERE ($1='' OR strpos(lower(display_name),lower($1))>0 OR id=$1 OR lower(email)=lower($1)) AND ($2='' OR id>$2) ORDER BY id LIMIT 50", query, after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.ManagedAccount{}
	for rows.Next() {
		var item application.ManagedAccount
		if failure = rows.Scan(&item.ID, &item.Name, &item.Role, &item.Blocked, &item.Version); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) ManageAccount(ctx context.Context, actor, id string, input application.AccountChange) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	if _, failure = transaction.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('qmanga-account-management'))"); failure != nil {
		return failure
	}
	var actorRole string
	if failure = transaction.QueryRow(ctx, "SELECT role FROM users WHERE id=$1 AND is_verified AND NOT is_blocked FOR UPDATE", actor).Scan(&actorRole); failure != nil {
		return domain.ErrForbidden
	}
	if actorRole != "admin" && actorRole != "moderator" {
		return domain.ErrForbidden
	}
	var role string
	var blocked bool
	var version int64
	if failure = transaction.QueryRow(ctx, "SELECT role,is_blocked,management_version FROM users WHERE id=$1 FOR UPDATE", id).Scan(&role, &blocked, &version); failure != nil {
		return translate(failure)
	}
	if actor == id {
		return domain.ErrForbidden
	}
	if actorRole == "moderator" && (role != "reader" || input.Role != "reader") {
		return domain.ErrForbidden
	}
	if role == "admin" && (input.Role != "admin" || input.Blocked) {
		var count int
		if failure = transaction.QueryRow(ctx, "SELECT count(*) FROM users WHERE role='admin' AND NOT is_blocked AND is_verified").Scan(&count); failure != nil {
			return failure
		}
		if count <= 1 {
			return domain.ErrForbidden
		}
	}
	if version != input.Version {
		return domain.ErrConflict
	}
	if input.Role != "reader" && input.Role != "moderator" && input.Role != "admin" {
		return domain.ErrInvalid
	}
	if _, failure = transaction.Exec(ctx, "UPDATE users SET role=$2,is_blocked=$3,management_version=management_version+1,updated_at=now() WHERE id=$1", id, input.Role, input.Blocked); failure != nil {
		return failure
	}
	if _, failure = transaction.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", id); failure != nil {
		return failure
	}
	metadata, _ := json.Marshal(map[string]any{"reason": input.Reason, "old_role": role, "role": input.Role, "was_blocked": blocked, "blocked": input.Blocked})
	if _, failure = transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,'manage_account','user',$3,$4)", uuid.NewString(), actor, id, metadata); failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) Audit(ctx context.Context, actor, after string) ([]application.AuditEntry, error) {
	role, failure := repository.staff(ctx, actor)
	if failure != nil || role != "admin" {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,coalesce(user_id,''),action,entity_type,entity_id,coalesce(metadata->>'reason',''),created_at FROM audit_logs WHERE ($1='' OR (created_at,id)<(SELECT created_at,id FROM audit_logs WHERE id=$1)) ORDER BY created_at DESC,id DESC LIMIT 50", after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.AuditEntry{}
	for rows.Next() {
		var item application.AuditEntry
		if failure = rows.Scan(&item.ID, &item.Actor, &item.Action, &item.EntityType, &item.EntityID, &item.Reason, &item.CreatedAt); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
