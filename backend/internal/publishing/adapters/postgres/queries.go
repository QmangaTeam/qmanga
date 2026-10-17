package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

const studioColumns = `revision.id,publication.id,publication.project_id,chapter.id,title.id,title.name_kk,chapter.label,team.name,revision.status,revision.moderation_note,revision.version,revision.source_note,revision.credits,revision.rights_confirmed,coalesce(revision.rights_confirmed_by,'')`
const studioJoins = ` FROM revisions revision JOIN releases publication ON publication.id=revision.release_id JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN titles title ON title.id=chapter.title_id JOIN teams team ON team.id=publication.team_id `

func scanStudio(row pgx.Row) (application.StudioRevision, error) {
	item := application.StudioRevision{Pages: []application.Page{}}
	failure := row.Scan(&item.ID, &item.ReleaseID, &item.ProjectID, &item.ChapterID, &item.TitleID, &item.TitleName, &item.ChapterLabel, &item.TeamName, &item.State, &item.Note, &item.Version, &item.Source, &item.Credits, &item.RightsConfirmed, &item.ConfirmedBy)
	return item, failure
}
func (repository *Repository) Revision(ctx context.Context, actor, id string) (application.StudioRevision, error) {
	item, failure := scanStudio(repository.pool.QueryRow(ctx, "SELECT "+studioColumns+studioJoins+"WHERE revision.id=$1", id))
	if failure != nil {
		return item, translate(failure)
	}
	if failure = access(ctx, repository.pool, actor, item.ProjectID, "read_draft"); failure != nil {
		return application.StudioRevision{}, failure
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,revision_id,status,image_format,source_key,checksum,failure_code,upload_key,position,width,height,byte_size FROM pages WHERE revision_id=$1 ORDER BY position", id)
	if failure != nil {
		return item, failure
	}
	defer rows.Close()
	for rows.Next() {
		var page application.Page
		if failure = rows.Scan(&page.ID, &page.RevisionID, &page.Status, &page.Format, &page.SourceKey, &page.Checksum, &page.Failure, &page.UploadKey, &page.Position, &page.Width, &page.Height, &page.Bytes); failure != nil {
			return item, failure
		}
		item.Pages = append(item.Pages, page)
	}
	return item, rows.Err()
}
func (repository *Repository) ProjectChapters(ctx context.Context, actor, projectID string) ([]application.ChapterChoice, error) {
	if failure := access(ctx, repository.pool, actor, projectID, "read_draft"); failure != nil {
		return nil, failure
	}
	rows, failure := repository.pool.Query(ctx, "SELECT chapter.id,chapter.label,chapter.title_kk,chapter.order_num FROM chapters chapter JOIN projects project ON project.title_id=chapter.title_id WHERE project.id=$1 ORDER BY chapter.order_num LIMIT 1000", projectID)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.ChapterChoice{}
	for rows.Next() {
		var item application.ChapterChoice
		if failure = rows.Scan(&item.ID, &item.Label, &item.Name, &item.Order); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) revisions(ctx context.Context, condition string, args ...any) ([]application.StudioRevision, error) {
	rows, failure := repository.pool.Query(ctx, "SELECT "+studioColumns+studioJoins+"WHERE "+condition+" ORDER BY revision.created_at DESC LIMIT 100", args...)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.StudioRevision{}
	for rows.Next() {
		item, failure := scanStudio(rows)
		if failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) ProjectRevisions(ctx context.Context, actor, projectID string) ([]application.StudioRevision, error) {
	if failure := access(ctx, repository.pool, actor, projectID, "read_draft"); failure != nil {
		return nil, failure
	}
	return repository.revisions(ctx, "publication.project_id=$1", projectID)
}
func (repository *Repository) ModerationQueue(ctx context.Context, actor string) ([]application.StudioRevision, error) {
	var allowed bool
	if failure := repository.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_verified AND role IN('admin','moderator'))", actor).Scan(&allowed); failure != nil {
		return nil, failure
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return repository.revisions(ctx, "revision.status='submitted'")
}
func (repository *Repository) Manifest(ctx context.Context, actor, releaseID, previewID string) (application.Manifest, error) {
	manifest := application.Manifest{Pages: []application.ManifestPage{}}
	var state, revisionState string
	failure := repository.pool.QueryRow(ctx, `SELECT publication.id,coalesce(publication.current_revision_id,''),title.id,title.slug,title.name_kk,chapter.label,publication.project_id,team.id,team.name,title.type,title.direction,chapter.order_num,publication.status,publication.updated_at FROM releases publication JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN projects project ON project.id=publication.project_id AND project.title_id=chapter.title_id AND project.team_id=publication.team_id JOIN titles title ON title.id=chapter.title_id JOIN teams team ON team.id=publication.team_id WHERE publication.id=$1`, releaseID).Scan(&manifest.ReleaseID, &manifest.RevisionID, &manifest.TitleID, &manifest.TitleSlug, &manifest.TitleName, &manifest.ChapterLabel, &manifest.ProjectID, &manifest.TeamID, &manifest.TeamName, &manifest.Kind, &manifest.Direction, &manifest.ChapterOrder, &state, &manifest.PublishedAt)
	if failure != nil {
		return manifest, translate(failure)
	}
	if previewID != "" {
		if failure = access(ctx, repository.pool, actor, manifest.ProjectID, "read_draft"); failure != nil {
			return manifest, failure
		}
		manifest.RevisionID = previewID
	} else {
		if state == "withdrawn" {
			return manifest, domain.ErrWithdrawn
		}
		if state != "published" || manifest.RevisionID == "" {
			return manifest, domain.ErrNotFound
		}
	}
	if failure = repository.pool.QueryRow(ctx, "SELECT status,source_note,credits FROM revisions WHERE id=$1 AND release_id=$2", manifest.RevisionID, releaseID).Scan(&revisionState, &manifest.Source, &manifest.Credits); failure != nil {
		return manifest, translate(failure)
	}
	if previewID == "" && revisionState != "published" {
		return manifest, domain.ErrNotFound
	}
	rows, failure := repository.pool.Query(ctx, `SELECT page.id,page.position,page.width,page.height,variant.id,variant.tier,variant.width,variant.height,variant.byte_size,variant.offset_y,variant.mime,variant.checksum FROM pages page JOIN page_variants variant ON variant.page_id=page.id WHERE page.revision_id=$1 AND page.status='ready' AND variant.tier<>'original' ORDER BY page.position,variant.tier,variant.offset_y`, manifest.RevisionID)
	if failure != nil {
		return manifest, failure
	}
	positions := map[string]int{}
	for rows.Next() {
		var pageID string
		var position, width, height int
		var variant application.Variant
		if failure = rows.Scan(&pageID, &position, &width, &height, &variant.ID, &variant.Tier, &variant.Width, &variant.Height, &variant.Bytes, &variant.Top, &variant.Mime, &variant.Checksum); failure != nil {
			rows.Close()
			return manifest, failure
		}
		index, exists := positions[pageID]
		if !exists {
			index = len(manifest.Pages)
			positions[pageID] = index
			manifest.Pages = append(manifest.Pages, application.ManifestPage{ID: pageID, Position: position, Width: width, Height: height, Variants: map[string][]application.Variant{"eco": {}, "standard": {}, "high": {}}})
		}
		variant.PageID = pageID
		scaledHeight := max(1, int(int64(height)*int64(variant.Width)/int64(width)))
		variant.Offset = float64(variant.Top) / float64(scaledHeight)
		manifest.Pages[index].Variants[variant.Tier] = append(manifest.Pages[index].Variants[variant.Tier], variant)
	}
	failure = rows.Err()
	rows.Close()
	if failure != nil {
		return manifest, failure
	}
	if previewID == "" && len(manifest.Pages) == 0 {
		return manifest, domain.ErrNotFound
	}
	for _, direction := range []string{"previous", "next"} {
		operator, sort := "<", "DESC"
		if direction == "next" {
			operator, sort = ">", "ASC"
		}
		var neighbor string
		failure = repository.pool.QueryRow(ctx, `SELECT coalesce((SELECT publication.id FROM releases publication JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN revisions revision ON revision.id=publication.current_revision_id AND revision.release_id=publication.id WHERE publication.project_id=$1 AND publication.status='published' AND revision.status='published' AND chapter.order_num`+operator+`$2 ORDER BY chapter.order_num `+sort+` LIMIT 1),'')`, manifest.ProjectID, manifest.ChapterOrder).Scan(&neighbor)
		if failure != nil {
			return manifest, failure
		}
		if direction == "next" {
			manifest.NextID = neighbor
		} else {
			manifest.PreviousID = neighbor
		}
	}
	return manifest, nil
}
func (repository *Repository) Asset(ctx context.Context, actor, id string) (application.Asset, error) {
	var asset application.Asset
	var projectID string
	var public bool
	failure := repository.pool.QueryRow(ctx, `SELECT variant.id,variant.page_id,variant.tier,variant.object_key,variant.mime,variant.checksum,variant.width,variant.height,variant.offset_y,variant.byte_size,publication.project_id,coalesce(publication.status='published' AND publication.current_revision_id=revision.id AND revision.status='published',false),revision.updated_at FROM page_variants variant JOIN pages page ON page.id=variant.page_id JOIN revisions revision ON revision.id=page.revision_id JOIN releases publication ON publication.id=revision.release_id WHERE variant.id=$1 AND page.status='ready'`, id).Scan(&asset.ID, &asset.PageID, &asset.Tier, &asset.ObjectKey, &asset.Mime, &asset.Checksum, &asset.Width, &asset.Height, &asset.Top, &asset.Bytes, &projectID, &public, &asset.ModifiedAt)
	if failure != nil {
		return asset, translate(failure)
	}
	if !public {
		if failure = access(ctx, repository.pool, actor, projectID, "read_draft"); failure != nil {
			return application.Asset{}, domain.ErrNotFound
		}
	}
	asset.Public = public && asset.Tier != "original"
	return asset, nil
}
func (repository *Repository) Original(ctx context.Context, actor, releaseID, pageID string) (application.Variant, error) {
	manifest, failure := repository.Manifest(ctx, actor, releaseID, "")
	if failure != nil {
		return application.Variant{}, failure
	}
	var variantID string
	failure = repository.pool.QueryRow(ctx, "SELECT variant.id FROM page_variants variant JOIN pages page ON page.id=variant.page_id WHERE page.id=$1 AND page.revision_id=$2 AND variant.tier='original'", pageID, manifest.RevisionID).Scan(&variantID)
	if failure != nil {
		return application.Variant{}, translate(failure)
	}
	asset, failure := repository.Asset(ctx, actor, variantID)
	return asset.Variant, failure
}

var _ application.Repository = (*Repository)(nil)
