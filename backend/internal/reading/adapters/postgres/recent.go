package postgres

import (
	"context"
	"qmanga/internal/reading/application"
)

const referenceColumns = `title.id,title.name_kk,title.slug,coalesce(title.cover_asset_id,''),publication.id,coalesce(publication.current_revision_id,''),coalesce(chapter.label,''),coalesce((SELECT name FROM teams WHERE id=publication.team_id),'')`

func (repository *Repository) Recent(ctx context.Context, actor string, limit int) ([]application.RecentEntry, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT `+referenceColumns+`,progress.release_id,coalesce(progress.revision_id,''),coalesce(progress.page_id,''),progress.fraction,progress.furthest_chapter_order,progress.version,progress.device_id,progress.sequence,progress.updated_at,progress.page_position,
 CASE WHEN publication.status<>'published' OR publication.language<>'kk' OR current_revision.status IS DISTINCT FROM 'published' THEN 'withdrawn'
 WHEN publication.current_revision_id IS DISTINCT FROM progress.revision_id OR NOT EXISTS(SELECT 1 FROM pages page WHERE page.id=progress.page_id AND page.revision_id=progress.revision_id) THEN 'updated' ELSE 'available' END
 FROM reading_progress progress JOIN users account ON account.id=progress.user_id AND NOT account.is_blocked JOIN titles title ON title.id=progress.title_id
 JOIN releases publication ON publication.id=progress.release_id JOIN chapters chapter ON chapter.id=publication.chapter_id AND chapter.title_id=title.id
 LEFT JOIN revisions current_revision ON current_revision.id=publication.current_revision_id AND current_revision.release_id=publication.id
 WHERE progress.user_id=$1 ORDER BY progress.updated_at DESC,progress.title_id DESC LIMIT $2`, actor, limit)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.RecentEntry{}
	for rows.Next() {
		var item application.RecentEntry
		if failure = rows.Scan(&item.TitleID, &item.Name, &item.Slug, &item.CoverID, &item.ReleaseID, &item.RevisionID, &item.ChapterLabel, &item.TeamName, &item.Progress.ReleaseID, &item.Progress.RevisionID, &item.Progress.PageID, &item.Progress.Fraction, &item.Progress.FurthestOrder, &item.Progress.Version, &item.Progress.DeviceID, &item.Progress.Sequence, &item.UpdatedAt, &item.PagePosition, &item.Availability); failure != nil {
			return nil, failure
		}
		item.Progress.TitleID = item.TitleID
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *Repository) References(ctx context.Context, ids []string) ([]application.Reference, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT `+referenceColumns+` FROM releases publication JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN titles title ON title.id=chapter.title_id JOIN revisions revision ON revision.id=publication.current_revision_id AND revision.release_id=publication.id AND revision.status='published' WHERE publication.id=ANY($1::varchar[]) AND publication.status='published' AND publication.language='kk'`, ids)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Reference{}
	for rows.Next() {
		var item application.Reference
		if failure = rows.Scan(&item.TitleID, &item.Name, &item.Slug, &item.CoverID, &item.ReleaseID, &item.RevisionID, &item.ChapterLabel, &item.TeamName); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
