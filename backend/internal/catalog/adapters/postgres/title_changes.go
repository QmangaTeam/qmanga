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

type changeQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func changeSnapshot(ctx context.Context, q changeQuery, id string) (domain.Metadata, domain.ChangeSnapshot, error) {
	var base domain.Metadata
	var snapshot domain.ChangeSnapshot
	var rating int
	var categories []string
	err := q.QueryRow(ctx, `SELECT slug,name_kk,name_original,aliases,authors,type,status_original,status_translation,description_kk,direction,age_rating,publication,is_demo,is_featured,is_listed,coalesce(cover_asset_id,''),banner_asset_id,ARRAY(SELECT genre_id FROM title_genres WHERE title_id=title.id ORDER BY genre_id),ARRAY(SELECT category_id FROM title_categories WHERE title_id=title.id ORDER BY category_id) FROM titles title WHERE id=$1`, id).Scan(&base.Slug, &base.Name, &base.OriginalName, &base.Aliases, &base.Authors, &base.Kind, &base.OriginalStatus, &base.TranslationStatus, &base.Description, &base.Direction, &rating, &base.Publication, &base.Demo, &base.Featured, &base.Listed, &snapshot.CoverID, &snapshot.BannerID, &base.GenreIDs, &categories)
	base.AgeRating = &rating
	snapshot.Data = domain.ProposedTitleData{Name: base.Name, OriginalName: base.OriginalName, Aliases: base.Aliases, Authors: base.Authors, Kind: base.Kind, OriginalStatus: base.OriginalStatus, TranslationStatus: base.TranslationStatus, Description: base.Description, Direction: base.Direction, AgeRating: rating, Publication: base.Publication, GenreIDs: base.GenreIDs, CategoryIDs: categories}
	return base, snapshot, translate(err)
}
func changeStaff(ctx context.Context, q changeQuery, actor string) (bool, error) {
	var allowed bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND role IN('admin','moderator') AND is_verified AND NOT is_blocked)`, actor).Scan(&allowed)
	return allowed, err
}
func changeAccess(ctx context.Context, tx pgx.Tx, actor, project string) (application.TitleChangeContext, error) {
	var item application.TitleChangeContext
	var team, status string
	err := tx.QueryRow(ctx, `SELECT project.id,project.title_id,project.team_id,project.status,team.name FROM projects project JOIN teams team ON team.id=project.team_id WHERE project.id=$1`, project).Scan(&item.ProjectID, &item.TitleID, &team, &status, &item.TeamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, domain.ErrForbidden
	}
	if err != nil {
		return item, err
	}
	if status != "active" {
		return item, domain.ErrForbidden
	}
	if err = studioTeam(ctx, tx, actor, team, true); err != nil {
		return item, err
	}
	item.Base, item.Snapshot, err = changeSnapshot(ctx, tx, item.TitleID)
	item.TitleName = item.Base.Name
	item.Slug = item.Base.Slug
	return item, err
}
func (r *Repository) ChangeContext(ctx context.Context, actor, project string) (application.TitleChangeContext, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return application.TitleChangeContext{}, err
	}
	defer tx.Rollback(ctx)
	item, err := changeAccess(ctx, tx, actor, project)
	if err != nil {
		return item, err
	}
	return item, tx.Commit(ctx)
}
func (r *Repository) ChangeExists(ctx context.Context, id string) (bool, error) {
	var found bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM title_change_requests WHERE id=$1)`, id).Scan(&found)
	return found, err
}
func (r *Repository) CreateChange(ctx context.Context, actor, project, id string, input domain.TitleChangeInput, assets []application.TitleChangeAsset) (application.TitleChange, error) {
	var empty application.TitleChange
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(ctx)
	current, err := changeAccess(ctx, tx, actor, project)
	if err != nil {
		return empty, err
	}
	var locked, team string
	if err = tx.QueryRow(ctx, `SELECT title.id,project.team_id FROM titles title JOIN projects project ON project.title_id=title.id WHERE project.id=$1 FOR UPDATE OF title`, project).Scan(&locked, &team); err != nil {
		return empty, err
	}
	current.Base, current.Snapshot, err = changeSnapshot(ctx, tx, current.TitleID)
	if err != nil {
		return empty, err
	}
	if input.BaseVersion != current.Snapshot.Version() {
		return empty, domain.ErrConflict
	}
	evidence, cover, banner := 0, 0, 0
	coverID, bannerID := "", ""
	for _, asset := range assets {
		switch asset.Purpose {
		case "evidence":
			evidence++
		case "cover":
			cover++
			coverID = asset.ID
		case "banner":
			banner++
			bannerID = asset.ID
		default:
			return empty, domain.ErrInvalid
		}
	}
	if cover > 1 || banner > 1 {
		return empty, domain.ErrInvalid
	}
	// A chapter page is copied into a processed cover asset before persistence.
	validation := input
	validation.CoverPageID = ""
	if err = validation.Validate(evidence, cover == 1, banner == 1); err != nil {
		return empty, err
	}
	input.CoverAction = validation.CoverAction
	input.BannerAction = validation.BannerAction
	input.Reason = validation.Reason
	input.Links = validation.Links
	if _, err = input.Data.Apply(current.TitleID, current.Base); err != nil {
		return empty, err
	}
	for table, ids := range map[string][]string{"genres": input.Data.GenreIDs, "categories": input.Data.CategoryIDs} {
		var count int
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE id=ANY($1::text[])", ids).Scan(&count); err != nil {
			return empty, err
		}
		if count != len(ids) {
			return empty, domain.ErrInvalid
		}
	}
	var pending, today int
	if err = tx.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status='pending'),count(*) FILTER(WHERE created_at>now()-interval '1 day') FROM title_change_requests WHERE team_id=$1`, team).Scan(&pending, &today); err != nil {
		return empty, err
	}
	if pending >= 5 || today >= 20 {
		return empty, domain.ErrRequestLimit
	}
	before, _ := json.Marshal(current.Snapshot)
	proposed, _ := json.Marshal(input.Data)
	if input.Links == nil {
		input.Links = []string{}
	}
	links, _ := json.Marshal(input.Links)
	_, err = tx.Exec(ctx, `INSERT INTO title_change_requests(id,title_id,project_id,team_id,requested_by,before_data,proposed_data,base_version,reason,links,cover_action,banner_action,cover_page_id,cover_asset_id,banner_asset_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, id, current.TitleID, project, team, actor, before, proposed, input.BaseVersion, input.Reason, links, input.CoverAction, input.BannerAction, input.CoverPageID, coverID, bannerID)
	if err != nil {
		return empty, translate(err)
	}
	for _, a := range assets {
		_, err = tx.Exec(ctx, `INSERT INTO title_change_assets(id,request_id,purpose,object_key,mime,checksum,width,height,byte_size,position)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, a.ID, id, a.Purpose, a.ObjectKey, a.Mime, a.Checksum, a.Width, a.Height, a.Bytes, a.Position)
		if err != nil {
			return empty, translate(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return empty, err
	}
	detail, err := r.ChangeDetail(ctx, actor, id)
	return detail.Request, err
}

const changeColumns = `request.id,coalesce(request.project_id,''),request.title_id,title.name_kk,coalesce(team.name,''),coalesce(request.requested_by,''),coalesce(author.display_name,''),request.status,request.reason,request.resolution,coalesce(request.reviewed_by,''),coalesce(reviewer.display_name,''),request.base_version,request.cover_action,request.banner_action,request.cover_asset_id,request.banner_asset_id,request.cover_page_id,request.links,request.before_data,request.proposed_data,request.created_at,request.reviewed_at`
const changeJoins = ` FROM title_change_requests request JOIN titles title ON title.id=request.title_id LEFT JOIN teams team ON team.id=request.team_id LEFT JOIN users author ON author.id=request.requested_by LEFT JOIN users reviewer ON reviewer.id=request.reviewed_by `

func scanChange(row interface{ Scan(...any) error }) (application.TitleChange, error) {
	var x application.TitleChange
	err := row.Scan(&x.ID, &x.ProjectID, &x.TitleID, &x.TitleName, &x.TeamName, &x.RequestedBy, &x.RequestedName, &x.Status, &x.Reason, &x.Resolution, &x.ReviewedBy, &x.ReviewedName, &x.BaseVersion, &x.CoverAction, &x.BannerAction, &x.CoverAssetID, &x.BannerAssetID, &x.CoverPageID, &x.Links, &x.Before, &x.Proposed, &x.CreatedAt, &x.ReviewedAt)
	return x, translate(err)
}
func (r *Repository) ChangeDetail(ctx context.Context, actor, id string) (application.TitleChangeDetail, error) {
	var result application.TitleChangeDetail
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	result.CanReview, err = changeStaff(ctx, tx, actor)
	if err != nil {
		return result, err
	}
	result.Request, err = scanChange(tx.QueryRow(ctx, "SELECT "+changeColumns+changeJoins+" WHERE request.id=$1", id))
	if err != nil {
		return result, err
	}
	if !result.CanReview {
		var allowed bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects project JOIN teams team ON team.id=project.team_id AND team.is_approved JOIN team_memberships membership ON membership.team_id=team.id JOIN users account ON account.id=membership.user_id AND account.is_verified AND NOT account.is_blocked WHERE project.id=$1 AND account.id=$2)`, result.Request.ProjectID, actor).Scan(&allowed)
		if err != nil {
			return result, err
		}
		if !allowed {
			return result, domain.ErrForbidden
		}
	}
	_, result.Current, err = changeSnapshot(ctx, tx, result.Request.TitleID)
	if err != nil {
		return result, err
	}
	result.Request.Assets = []application.TitleChangeAsset{}
	rows, err := tx.Query(ctx, `SELECT id,purpose,object_key,mime,checksum,width,height,byte_size,position FROM title_change_assets WHERE request_id=$1 ORDER BY position`, id)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var a application.TitleChangeAsset
		if err = rows.Scan(&a.ID, &a.Purpose, &a.ObjectKey, &a.Mime, &a.Checksum, &a.Width, &a.Height, &a.Bytes, &a.Position); err != nil {
			rows.Close()
			return result, err
		}
		result.Request.Assets = append(result.Request.Assets, a)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}
func (r *Repository) ProjectChanges(ctx context.Context, actor, project, after string) ([]application.TitleChange, error) {
	// Listing and proposal creation share the editor ACL; evidence detail allows current viewers as well.
	if _, err := r.ChangeContext(ctx, actor, project); err != nil {
		return nil, err
	}
	return r.listChanges(ctx, project, after, false)
}
func (r *Repository) ChangeQueue(ctx context.Context, actor, after string) ([]application.TitleChange, error) {
	allowed, err := changeStaff(ctx, r.pool, actor)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return r.listChanges(ctx, "", after, true)
}
func (r *Repository) listChanges(ctx context.Context, project, after string, queue bool) ([]application.TitleChange, error) {
	condition := "request.project_id=$1"
	if queue {
		condition = "request.status='pending' AND $1::text=''"
	}
	rows, err := r.pool.Query(ctx, "SELECT "+changeColumns+changeJoins+" WHERE "+condition+" AND ($2::text='' OR (request.created_at,request.id)<(SELECT created_at,id FROM title_change_requests WHERE id=$2)) ORDER BY request.created_at DESC,request.id DESC LIMIT 50", project, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []application.TitleChange{}
	for rows.Next() {
		item, err := scanChange(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) ReviewChange(ctx context.Context, actor, id, decision, reason string) error {
	if decision != "approve" && decision != "reject" || len(reason) > 8000 || decision == "reject" && len(strings.TrimSpace(reason)) < 3 {
		return domain.ErrInvalid
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT role IN('admin','moderator') AND is_verified AND NOT is_blocked FROM users WHERE id=$1 FOR SHARE`, actor).Scan(&allowed)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && !allowed {
		return domain.ErrForbidden
	}
	if err != nil {
		return err
	}
	item, err := scanChange(tx.QueryRow(ctx, "SELECT "+changeColumns+changeJoins+" WHERE request.id=$1 FOR UPDATE OF request", id))
	if err != nil {
		return err
	}
	if item.Status != "pending" {
		return domain.ErrConflict
	}
	var locked string
	if err = tx.QueryRow(ctx, `SELECT id FROM titles WHERE id=$1 FOR UPDATE`, item.TitleID).Scan(&locked); err != nil {
		return err
	}
	base, snapshot, err := changeSnapshot(ctx, tx, item.TitleID)
	if err != nil {
		return err
	}
	status := "rejected"
	if decision == "approve" {
		if snapshot.Version() != item.BaseVersion {
			return domain.ErrConflict
		}
		title, err := item.Proposed.Apply(item.TitleID, base)
		if err != nil {
			return err
		}
		// Keep approved cover metadata readable by the previous API during an application rollback.
		// Pending artwork and evidence are never copied into this legacy public-cover read model.
		if item.CoverAction == "replace" {
			result, failure := tx.Exec(ctx, `INSERT INTO title_cover_assets(id,title_id,object_key,mime,checksum,width,height,byte_size,selected_by) SELECT id,$2,object_key,mime,checksum,width,height,byte_size,$3 FROM title_change_assets WHERE id=$1 AND request_id=$4 AND purpose='cover'`, item.CoverAssetID, item.TitleID, actor, item.ID)
			if failure != nil {
				return translate(failure)
			}
			if result.RowsAffected() != 1 {
				return domain.ErrInvalid
			}
		}
		cover, banner := snapshot.CoverID, snapshot.BannerID
		if item.CoverAction == "replace" {
			cover = item.CoverAssetID
		}
		if item.CoverAction == "remove" {
			cover = ""
		}
		if item.BannerAction == "replace" {
			banner = item.BannerAssetID
		}
		if item.BannerAction == "remove" {
			banner = ""
		}
		_, err = tx.Exec(ctx, `UPDATE titles SET name_kk=$2,name_original=$3,aliases=$4,authors=$5,type=$6,status_original=$7,status_translation=$8,description_kk=$9,direction=$10,age_rating=$11,publication=$12,search_text=$13,cover_asset_id=$14,banner_asset_id=$15,cover_managed=CASE WHEN $16 THEN true ELSE cover_managed END,updated_at=now() WHERE id=$1`, title.ID, title.Name, title.OriginalName, title.Aliases, title.Authors, title.Kind, title.OriginalStatus, title.TranslationStatus, title.Description, title.Direction, title.AgeRating, title.Publication, strings.ToLower(title.Name+" "+title.OriginalName+" "+strings.Join(title.Aliases, " ")), cover, banner, item.CoverAction != "keep")
		if err != nil {
			return translate(err)
		}
		for table, ids := range map[string][]string{"title_genres": item.Proposed.GenreIDs, "title_categories": item.Proposed.CategoryIDs} {
			column := "genre_id"
			if table == "title_categories" {
				column = "category_id"
			}
			if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE title_id=$1", title.ID); err != nil {
				return err
			}
			for _, value := range ids {
				if _, err = tx.Exec(ctx, "INSERT INTO "+table+"(title_id,"+column+") VALUES($1,$2)", title.ID, value); err != nil {
					return translate(err)
				}
			}
		}
		status = "approved"
	}
	if _, err = tx.Exec(ctx, `UPDATE title_change_requests SET status=$2,resolution=$3,reviewed_by=$4,reviewed_at=now() WHERE id=$1`, id, status, reason, actor); err != nil {
		return err
	}
	metadata, _ := json.Marshal(map[string]string{"request_id": id, "reason": reason, "status": status})
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata)VALUES($1,$2,$3,'title',$4,$5)`, uuid.NewString(), actor, "title_change_"+status, item.TitleID, metadata); err != nil {
		return err
	}
	data, _ := json.Marshal(map[string]string{"project_id": item.ProjectID, "request_id": id})
	if item.RequestedBy != "" {
		if _, err = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,type,title_kk,body_kk,data)VALUES($1,$2,$3,$4,$5,$6)`, uuid.NewString(), item.RequestedBy, "title_change_"+status, item.TitleName, strings.TrimSpace(item.TitleName+"\n"+reason), data); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *Repository) ChangeCoverPage(ctx context.Context, actor, project, page string) (application.TitleChangeAsset, error) {
	var a application.TitleChangeAsset
	if _, err := r.ChangeContext(ctx, actor, project); err != nil {
		return a, err
	}
	err := r.pool.QueryRow(ctx, `SELECT variant.id,variant.object_key,variant.mime,variant.checksum,variant.width,variant.height,variant.byte_size FROM page_variants variant JOIN pages page ON page.id=variant.page_id JOIN revisions revision ON revision.id=page.revision_id JOIN releases publication ON publication.id=revision.release_id WHERE page.id=$1 AND publication.project_id=$2 AND page.status='ready' AND variant.tier='eco' ORDER BY variant.offset_y LIMIT 1`, page, project).Scan(&a.ID, &a.ObjectKey, &a.Mime, &a.Checksum, &a.Width, &a.Height, &a.Bytes)
	return a, translate(err)
}

var _ application.TitleChangeRepository = (*Repository)(nil)
