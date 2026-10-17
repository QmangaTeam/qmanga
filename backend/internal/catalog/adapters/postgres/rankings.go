package postgres

import (
	"context"
	"qmanga/internal/catalog/application"
)

func (repository *Repository) Rankings(ctx context.Context, kind string, limit int) ([]application.Summary, error) {
	orderBy := "coalesce(readers.total,0) DESC,score DESC,rating_count DESC,title.id DESC"
	if kind != "" {
		orderBy = "score DESC,rating_count DESC,title.id DESC"
	}
	rows, failure := repository.pool.Query(ctx, `WITH reader_counts AS (
 SELECT progress.title_id,count(*) AS total FROM reading_progress progress
 JOIN users reader ON reader.id=progress.user_id AND NOT reader.is_blocked GROUP BY progress.title_id)
 SELECT `+summaryColumns+` FROM titles title
 LEFT JOIN reader_counts readers ON readers.title_id=title.id
 WHERE `+hasChapters+` AND ($1='' OR title.type=$1) ORDER BY `+orderBy+` LIMIT $2`, kind, limit)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Summary{}
	for rows.Next() {
		item, failure := scanSummary(rows)
		if failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
