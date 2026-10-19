package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	"strings"
)

const requestColumns = `request.id,request.team_id,COALESCE(request.title_id,''),request.name_kk,request.name_original,request.title_slug,request.kind,request.description_kk,request.source_url,request.note,request.status,request.resolution,COALESCE(request.project_id,''),team.name,COALESCE(title.name_kk,request.name_kk),request.created_at,request.updated_at,coalesce(title.cover_asset_id,'')`
const requestJoins = ` FROM project_requests request JOIN teams team ON team.id=request.team_id LEFT JOIN titles title ON title.id=request.title_id `

func scanRequest(row interface{ Scan(...any) error }) (domain.ProjectRequest, error) {
	var item domain.ProjectRequest
	failure := row.Scan(&item.ID, &item.TeamID, &item.TitleID, &item.Name, &item.OriginalName, &item.Slug, &item.Kind, &item.Description, &item.SourceURL, &item.Note, &item.Status, &item.Resolution, &item.ProjectID, &item.TeamName, &item.TitleName, &item.CreatedAt, &item.UpdatedAt, &item.CoverID)
	return item, translate(failure)
}
func studioTeam(ctx context.Context, transaction pgx.Tx, actor, teamID string, approved bool) error {
	var role string
	var active bool
	failure := transaction.QueryRow(ctx, `SELECT membership.role,team.is_approved FROM teams team JOIN team_memberships membership ON membership.team_id=team.id AND membership.user_id=$1 JOIN users account ON account.id=membership.user_id AND account.is_verified AND NOT account.is_blocked WHERE team.id=$2 FOR UPDATE OF team FOR SHARE OF membership,account`, actor, teamID).Scan(&role, &active)
	if errors.Is(failure, pgx.ErrNoRows) || (failure == nil && (role != "owner" && role != "editor" || approved && !active)) {
		return domain.ErrForbidden
	}
	return failure
}
func (repository *Repository) CreateProjectRequest(ctx context.Context, actor string, request domain.ProjectRequest) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	if failure = studioTeam(ctx, transaction, actor, request.TeamID, true); failure != nil {
		return failure
	}
	if request.TitleID != "" {
		var visible bool
		if failure = transaction.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM titles title WHERE title.id=$1 AND "+titleVisible+")", request.TitleID).Scan(&visible); failure != nil {
			return failure
		}
		if !visible {
			return domain.ErrNotFound
		}
		var exists bool
		if failure = transaction.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE team_id=$1 AND title_id=$2)", request.TeamID, request.TitleID).Scan(&exists); failure != nil {
			return failure
		}
		if exists {
			return domain.ErrConflict
		}
	}
	var pending, today int
	if failure = transaction.QueryRow(ctx, "SELECT count(*) FILTER(WHERE status='pending'),count(*) FILTER(WHERE created_at>now()-interval '1 day') FROM project_requests WHERE team_id=$1", request.TeamID).Scan(&pending, &today); failure != nil {
		return failure
	}
	if pending >= 5 || today >= 20 {
		return domain.ErrRequestLimit
	}
	target := "title:" + request.TitleID
	if request.TitleID == "" {
		target = "new:" + strings.ToLower(strings.TrimSpace(request.Name))
	}
	_, failure = transaction.Exec(ctx, `INSERT INTO project_requests(id,team_id,requested_by,title_id,target_key,name_kk,name_original,title_slug,kind,description_kk,source_url,note) VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,$12)`, request.ID, request.TeamID, actor, request.TitleID, target, request.Name, request.OriginalName, request.Slug, request.Kind, request.Description, request.SourceURL, request.Note)
	if failure != nil {
		return translate(failure)
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) ProjectRequests(ctx context.Context, actor string, admin bool, after string) ([]domain.ProjectRequest, error) {
	condition := `EXISTS(SELECT 1 FROM team_memberships membership JOIN users account ON account.id=membership.user_id AND account.is_verified AND NOT account.is_blocked WHERE membership.team_id=request.team_id AND account.id=$1)`
	if admin {
		allowed, failure := repository.IsAdmin(ctx, actor)
		if failure != nil {
			return nil, failure
		}
		if !allowed {
			return nil, domain.ErrForbidden
		}
		condition = "request.status='pending' AND $1::text<>''"
	}
	rows, failure := repository.pool.Query(ctx, "SELECT "+requestColumns+requestJoins+" WHERE "+condition+" AND ($2::text='' OR (request.created_at,request.id)<(SELECT created_at,id FROM project_requests WHERE id=$2)) ORDER BY request.created_at DESC,request.id DESC LIMIT 50", actor, after)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []domain.ProjectRequest{}
	for rows.Next() {
		item, failure := scanRequest(rows)
		if failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) ReviewProjectRequest(ctx context.Context, actor, id, decision, reason, titleID string) (domain.ProjectRequest, error) {
	var result domain.ProjectRequest
	failure := repository.withAdmin(ctx, actor, func(transaction pgx.Tx) error {
		var teamID string
		if failure := transaction.QueryRow(ctx, "SELECT team_id FROM project_requests WHERE id=$1", id).Scan(&teamID); failure != nil {
			return failure
		}
		var approved bool
		if failure := transaction.QueryRow(ctx, "SELECT is_approved FROM teams WHERE id=$1 FOR UPDATE", teamID).Scan(&approved); failure != nil {
			return failure
		}
		item, failure := scanRequest(transaction.QueryRow(ctx, "SELECT "+requestColumns+requestJoins+" WHERE request.id=$1 FOR UPDATE OF request", id))
		if failure != nil {
			return failure
		}
		expected := "approved"
		if decision == "reject" {
			expected = "rejected"
		}
		if item.Status != "pending" {
			if item.Status == expected && item.Resolution == reason {
				result = item
				return nil
			}
			return domain.ErrConflict
		}
		if !approved {
			return domain.ErrForbidden
		}
		projectID := ""
		if decision == "approve" {
			if titleID == "" {
				titleID = item.TitleID
			}
			if titleID == "" {
				titleID = uuid.NewString()
				_, failure = transaction.Exec(ctx, `INSERT INTO titles(id,slug,name_kk,name_original,type,description_kk,search_text,direction) VALUES($1,$2,$3,$4,$5,$6,$7,'ltr')`, titleID, item.Slug, item.Name, item.OriginalName, item.Kind, item.Description, strings.ToLower(item.Name+" "+item.OriginalName))
				if failure != nil {
					return failure
				}
				if failure = audit(ctx, transaction, actor, "title", titleID); failure != nil {
					return failure
				}
			}
			var titleExists bool
			if failure = transaction.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM titles WHERE id=$1)", titleID).Scan(&titleExists); failure != nil {
				return failure
			}
			if !titleExists {
				return domain.ErrNotFound
			}
			projectID = uuid.NewString()
			if failure = transaction.QueryRow(ctx, `INSERT INTO projects(id,title_id,team_id) VALUES($1,$2,$3) ON CONFLICT(title_id,team_id) DO UPDATE SET team_id=excluded.team_id RETURNING id`, projectID, titleID, item.TeamID).Scan(&projectID); failure != nil {
				return failure
			}
		}
		_, failure = transaction.Exec(ctx, "UPDATE project_requests SET status=$2,resolution=$3,reviewed_by=$4,project_id=NULLIF($5,''),title_id=COALESCE(NULLIF($6,''),title_id),updated_at=now() WHERE id=$1", id, expected, reason, actor, projectID, titleID)
		if failure != nil {
			return failure
		}
		metadata, _ := json.Marshal(map[string]string{"reason": reason, "project_id": projectID})
		if _, failure = transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,$3,'project_request',$4,$5)", uuid.NewString(), actor, decision+"_project_request", id, metadata); failure != nil {
			return failure
		}
		notificationData, _ := json.Marshal(map[string]string{"project_id": projectID, "request_id": id})
		if _, failure = transaction.Exec(ctx, `INSERT INTO notifications(id,user_id,type,title_kk,body_kk,data) SELECT gen_random_uuid()::varchar,membership.user_id,$2,$3,$4,$5 FROM team_memberships membership WHERE membership.team_id=$1 AND membership.role IN('owner','editor')`, item.TeamID, "project_request_"+expected, item.TitleName, strings.TrimSpace(item.TitleName+"\n"+reason), notificationData); failure != nil {
			return failure
		}
		result, failure = scanRequest(transaction.QueryRow(ctx, "SELECT "+requestColumns+requestJoins+" WHERE request.id=$1", id))
		return failure
	})
	return result, failure
}
func (repository *Repository) CancelProjectRequest(ctx context.Context, actor, id string) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var teamID string
	if failure = transaction.QueryRow(ctx, "SELECT team_id FROM project_requests WHERE id=$1", id).Scan(&teamID); failure != nil {
		return translate(failure)
	}
	if failure = studioTeam(ctx, transaction, actor, teamID, false); failure != nil {
		return failure
	}
	result, failure := transaction.Exec(ctx, "UPDATE project_requests SET status='cancelled',updated_at=now() WHERE id=$1 AND status='pending'", id)
	if failure != nil {
		return failure
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) StudioProject(ctx context.Context, actor, id string) (application.StudioProject, error) {
	var item application.StudioProject
	failure := repository.pool.QueryRow(ctx, `SELECT project.id,project.title_id,project.team_id,title.name_kk,title.slug,team.name,team.slug,project.status,membership.role,team.is_approved,coalesce(title.cover_asset_id,'') FROM projects project JOIN titles title ON title.id=project.title_id JOIN teams team ON team.id=project.team_id JOIN team_memberships membership ON membership.team_id=team.id AND membership.user_id=$1 JOIN users account ON account.id=membership.user_id AND account.is_verified AND NOT account.is_blocked WHERE project.id=$2`, actor, id).Scan(&item.Project.ID, &item.Project.TitleID, &item.Project.TeamID, &item.Project.TitleName, &item.Project.TitleSlug, &item.Project.TeamName, &item.Project.TeamSlug, &item.Project.Status, &item.Role, &item.Approved, &item.CoverID)
	if errors.Is(failure, pgx.ErrNoRows) {
		return item, domain.ErrForbidden
	}
	return item, failure
}
func (repository *Repository) AdminProject(ctx context.Context, actor, id string) (application.Project, error) {
	var item application.Project
	failure := repository.withAdmin(ctx, actor, func(transaction pgx.Tx) error {
		return transaction.QueryRow(ctx, `SELECT project.id,project.title_id,project.team_id,title.name_kk,title.slug,team.name,team.slug,project.status
		 FROM projects project JOIN titles title ON title.id=project.title_id JOIN teams team ON team.id=project.team_id WHERE project.id=$1`, id).
			Scan(&item.ID, &item.TitleID, &item.TeamID, &item.TitleName, &item.TitleSlug, &item.TeamName, &item.TeamSlug, &item.Status)
	})
	return item, failure
}
func (repository *Repository) CreateStudioChapter(ctx context.Context, actor, project, label, volume, name string, order int) (string, error) {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return "", failure
	}
	defer transaction.Rollback(ctx)
	var teamID, titleID, status string
	if failure = transaction.QueryRow(ctx, "SELECT team_id,title_id,status FROM projects WHERE id=$1", project).Scan(&teamID, &titleID, &status); failure != nil {
		return "", domain.ErrForbidden
	}
	if failure = studioTeam(ctx, transaction, actor, teamID, true); failure != nil {
		return "", failure
	}
	if status != "active" {
		return "", domain.ErrForbidden
	}
	var locked string
	if failure = transaction.QueryRow(ctx, "SELECT id FROM titles WHERE id=$1 FOR UPDATE", titleID).Scan(&locked); failure != nil {
		return "", failure
	}
	var id, existingName string
	var existingOrder int
	failure = transaction.QueryRow(ctx, "SELECT id,title_kk,order_num FROM chapters WHERE title_id=$1 AND label=$2 AND volume=$3 ORDER BY order_num LIMIT 1", titleID, label, volume).Scan(&id, &existingName, &existingOrder)
	if failure == nil {
		if name != "" && name != existingName || order != 0 && order != existingOrder {
			return "", domain.ErrConflict
		}
		return id, transaction.Commit(ctx)
	}
	if !errors.Is(failure, pgx.ErrNoRows) {
		return "", failure
	}
	if order == 0 {
		if failure = transaction.QueryRow(ctx, "SELECT COALESCE(max(order_num),0)+1 FROM chapters WHERE title_id=$1", titleID).Scan(&order); failure != nil {
			return "", failure
		}
	}
	if failure = domain.ValidateChapter(label, order); failure != nil {
		return "", failure
	}
	id = uuid.NewString()
	if _, failure = transaction.Exec(ctx, "INSERT INTO chapters(id,title_id,label,volume,title_kk,order_num)VALUES($1,$2,$3,$4,$5,$6)", id, titleID, label, volume, name, order); failure != nil {
		return "", translate(failure)
	}
	if failure = audit(ctx, transaction, actor, "chapter", id); failure != nil {
		return "", failure
	}
	return id, transaction.Commit(ctx)
}
