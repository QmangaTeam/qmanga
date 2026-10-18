package postgres

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/url"
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
	rows, failure := repository.pool.Query(ctx, `SELECT audit.id,coalesce(audit.user_id,''),audit.action,audit.entity_type,audit.entity_id,coalesce(audit.metadata->>'reason',''),audit.created_at,
 coalesce(actor.display_name,''),coalesce(target_user.display_name,title.name_kk,team.name,''),audit.metadata,
 coalesce(title.slug,''),coalesce(team.slug,''),coalesce(release.id,''),coalesce(chapter.label,''),coalesce(comment.id,''),coalesce(report.id,''),coalesce(revision.id,release.current_revision_id,'')
 FROM audit_logs audit LEFT JOIN users actor ON actor.id=audit.user_id
 LEFT JOIN users target_user ON audit.entity_type='user' AND target_user.id=audit.entity_id
 LEFT JOIN revisions revision ON audit.entity_type IN('release','revision') AND revision.id=audit.entity_id
 LEFT JOIN releases release ON audit.entity_type IN('release','revision') AND release.id=coalesce(revision.release_id,audit.entity_id)
 LEFT JOIN comments comment ON audit.entity_type='community' AND comment.id=audit.entity_id
 LEFT JOIN reports report ON audit.entity_type='community' AND report.id=audit.entity_id
 LEFT JOIN projects project ON audit.entity_type='project' AND project.id=audit.entity_id
 LEFT JOIN chapters chapter ON chapter.id=CASE WHEN audit.entity_type='chapter' THEN audit.entity_id ELSE release.chapter_id END
 LEFT JOIN teams team ON team.id=CASE WHEN audit.entity_type='team' THEN audit.entity_id ELSE coalesce(release.team_id,project.team_id) END
 LEFT JOIN titles title ON title.id=CASE WHEN audit.entity_type='title' THEN audit.entity_id ELSE coalesce(chapter.title_id,project.title_id,comment.title_id,report.title_id) END
 WHERE ($1='' OR (audit.created_at,audit.id)<(SELECT created_at,id FROM audit_logs WHERE id=$1)) ORDER BY audit.created_at DESC,audit.id DESC LIMIT 50`, after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.AuditEntry{}
	for rows.Next() {
		var item application.AuditEntry
		var metadata []byte
		var titleSlug, teamSlug, releaseID, chapterLabel, commentID, reportID, revisionID string
		if failure = rows.Scan(&item.ID, &item.Actor, &item.Action, &item.EntityType, &item.EntityID, &item.Reason, &item.CreatedAt, &item.ActorName, &item.TargetName, &metadata, &titleSlug, &teamSlug, &releaseID, &chapterLabel, &commentID, &reportID, &revisionID); failure != nil {
			return nil, failure
		}
		var details map[string]any
		if failure = json.Unmarshal(metadata, &details); failure != nil {
			return nil, failure
		}
		item.Before, item.After = auditStates(details)
		switch {
		case item.EntityType == "user":
			item.TargetPath = "/moderation/accounts?q=" + url.QueryEscape(item.EntityID)
		case item.EntityType == "team" && teamSlug != "":
			item.TargetPath = "/teams/" + url.PathEscape(teamSlug)
		case releaseID != "":
			item.TargetPath = "/read/" + url.PathEscape(releaseID)
			if revisionID != "" {
				item.TargetPath += "?preview=" + url.QueryEscape(revisionID)
			}
			if chapterLabel != "" {
				item.TargetName += " · " + chapterLabel
			}
		case titleSlug != "":
			item.TargetPath = "/titles/" + url.PathEscape(titleSlug)
			if commentID != "" {
				item.TargetPath += "?tab=comments"
			}
		case item.EntityType == "project_request":
			item.TargetPath = "/admin/requests"
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Audit is deliberately a safe projection, never an arbitrary metadata dump.
func auditStates(metadata map[string]any) (map[string]any, map[string]any) {
	before, after := map[string]any{}, map[string]any{}
	allowed := map[string]bool{"role": true, "blocked": true, "hidden": true, "status": true, "revision_status": true, "release_status": true, "cover_asset_id": true}
	for key, destination := range map[string]map[string]any{"before": before, "after": after} {
		if state, ok := metadata[key].(map[string]any); ok {
			for field, value := range state {
				if allowed[field] {
					switch value.(type) {
					case string, bool, float64:
						destination[field] = value
					}
				}
			}
		}
	}
	if value, ok := metadata["old_role"].(string); ok {
		before["role"] = value
	}
	if value, ok := metadata["was_blocked"].(bool); ok {
		before["blocked"] = value
	}
	if value, ok := metadata["role"].(string); ok {
		after["role"] = value
	}
	if value, ok := metadata["blocked"].(bool); ok {
		after["blocked"] = value
	}

	return before, after
}
