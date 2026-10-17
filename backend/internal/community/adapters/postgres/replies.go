package postgres

import (
	"context"
	"qmanga/internal/community/application"
	"qmanga/internal/community/domain"
)

func (repository *Repository) Replies(ctx context.Context, parent, after string) (application.Discussion, error) {
	result := application.Discussion{Comments: []application.Comment{}}
	var title string
	if failure := repository.pool.QueryRow(ctx, `SELECT title_id FROM comments WHERE id=$1 AND parent_id IS NULL`, parent).Scan(&title); failure != nil {
		return result, domain.ErrNotFound
	}
	var valid bool
	if failure := repository.pool.QueryRow(ctx, `SELECT `+visibleTitle, title).Scan(&valid); failure != nil {
		return result, failure
	}
	if !valid {
		return result, domain.ErrNotFound
	}
	rows, failure := repository.pool.Query(ctx, `SELECT id,user_id,title_id,COALESCE(release_id,''),parent_id,CASE WHEN is_hidden OR deleted_at IS NOT NULL THEN '' ELSE body END,is_spoiler,is_hidden OR deleted_at IS NOT NULL,created_at FROM comments WHERE parent_id=$1 AND ($2::varchar='' OR (created_at,id)>(SELECT created_at,id FROM comments WHERE id=$2::varchar AND parent_id=$1))ORDER BY created_at,id LIMIT 20`, parent, after)
	if failure != nil {
		return result, failure
	}
	defer rows.Close()
	for rows.Next() {
		var item application.Comment
		if failure = rows.Scan(&item.ID, &item.UserID, &item.TitleID, &item.ReleaseID, &item.ParentID, &item.Body, &item.Spoiler, &item.Hidden, &item.CreatedAt); failure != nil {
			return result, failure
		}
		result.Comments = append(result.Comments, item)
	}
	if len(result.Comments) == 20 {
		result.Next = result.Comments[len(result.Comments)-1].ID
	}
	return result, rows.Err()
}
