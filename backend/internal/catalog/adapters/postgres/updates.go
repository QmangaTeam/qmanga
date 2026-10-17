package postgres

import (
	"context"
	"qmanga/internal/catalog/application"
)

func (repository *Repository) Updates(ctx context.Context, after, team string, limit int) (application.UpdatePage, error) {
	result := application.UpdatePage{Items: []application.Update{}}
	rows, failure := repository.pool.Query(ctx, `SELECT publication.id,title.id,title.name_kk,title.slug,chapter.label,team.name,team.slug,title.is_demo,publication.updated_at,title.cover_asset_id`+publicationJoins+` JOIN titles title ON title.id=chapter.title_id WHERE `+publicationVisible+` AND ($1='' OR (publication.updated_at,publication.id)<(SELECT updated_at,id FROM releases WHERE id=$1)) AND ($2='' OR team.slug=$2) ORDER BY publication.updated_at DESC,publication.id DESC LIMIT $3`, after, team, limit+1)
	if failure != nil {
		return result, failure
	}
	defer rows.Close()
	for rows.Next() {
		var item application.Update
		if failure = rows.Scan(&item.ReleaseID, &item.TitleID, &item.TitleName, &item.TitleSlug, &item.ChapterLabel, &item.TeamName, &item.TeamSlug, &item.Demo, &item.PublishedAt, &item.CoverID); failure != nil {
			return result, failure
		}
		result.Items = append(result.Items, item)
	}
	if failure = rows.Err(); failure != nil {
		return result, failure
	}
	if len(result.Items) > limit {
		result.Items = result.Items[:limit]
		result.Next = result.Items[limit-1].ReleaseID
	}
	return result, nil
}
