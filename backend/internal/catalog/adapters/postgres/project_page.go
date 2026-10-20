package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/catalog/application"
	"strconv"
	"strings"
)

func (repository *Repository) StudioProjectPage(ctx context.Context, actor string, filter application.ProjectFilter) (application.ProjectPage, error) {
	page := application.ProjectPage{Items: []application.Project{}}
	transaction, failure := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if failure != nil {
		return page, failure
	}
	defer transaction.Rollback(ctx)
	const joins = ` FROM projects project JOIN titles title ON title.id=project.title_id JOIN teams team ON team.id=project.team_id `
	const condition = `EXISTS(SELECT 1 FROM team_memberships membership JOIN users account ON account.id=membership.user_id WHERE membership.team_id=project.team_id AND account.id=$1 AND account.is_verified AND NOT account.is_blocked) AND ($2='' OR project.team_id=$2) AND ($3='' OR lower(title.name_kk||' '||team.name) LIKE $4)`
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
	arguments := []any{actor, filter.TeamID, filter.Query, "%" + escaped + "%"}
	if failure = transaction.QueryRow(ctx, `SELECT count(*)`+joins+` WHERE `+condition, arguments...).Scan(&page.Total); failure != nil {
		return page, failure
	}
	rows, failure := transaction.Query(ctx, `SELECT project.id,project.title_id,project.team_id,title.name_kk,title.slug,team.name,team.slug,project.status,coalesce(title.cover_asset_id,''),(SELECT count(DISTINCT chapter_id) FROM releases WHERE project_id=project.id),(SELECT count(*) FROM ratings rating JOIN users rating_user ON rating_user.id=rating.user_id AND NOT rating_user.is_blocked WHERE rating.title_id=title.id),coalesce((SELECT avg(rating.score)::float8 FROM ratings rating JOIN users rating_user ON rating_user.id=rating.user_id AND NOT rating_user.is_blocked WHERE rating.title_id=title.id),0),coalesce(work.id,''),coalesce(work.status,'')`+joins+` LEFT JOIN LATERAL (SELECT revision.id,revision.status FROM revisions revision JOIN releases release ON release.id=revision.release_id WHERE release.project_id=project.id AND NOT EXISTS(SELECT 1 FROM revisions newer WHERE newer.release_id=revision.release_id AND (newer.revision_num>revision.revision_num OR (newer.revision_num=revision.revision_num AND (newer.created_at,newer.id)>(revision.created_at,revision.id)))) AND revision.status IN ('draft','processing','ready','submitted','changes_requested') ORDER BY CASE revision.status WHEN 'changes_requested' THEN 0 WHEN 'ready' THEN 1 WHEN 'draft' THEN 2 WHEN 'processing' THEN 3 ELSE 4 END,revision.updated_at DESC,revision.id LIMIT 1) work ON true `+` WHERE `+condition+` ORDER BY title.name_kk,team.name,project.id LIMIT 101 OFFSET $5`, append(arguments, filter.Offset)...)
	if failure != nil {
		return page, failure
	}
	defer rows.Close()
	for rows.Next() {
		var item application.Project
		if failure = rows.Scan(&item.ID, &item.TitleID, &item.TeamID, &item.TitleName, &item.TitleSlug, &item.TeamName, &item.TeamSlug, &item.Status, &item.CoverID, &item.ChapterCount, &item.RatingCount, &item.Score, &item.NextRevisionID, &item.NextRevisionState); failure != nil {
			return page, failure
		}
		page.Items = append(page.Items, item)
	}
	if failure = rows.Err(); failure != nil {
		return page, failure
	}
	if len(page.Items) > 100 {
		page.Items = page.Items[:100]
		page.Next = strconv.Itoa(filter.Offset + 100)
	}
	return page, transaction.Commit(ctx)
}
