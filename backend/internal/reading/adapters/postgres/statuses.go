package postgres

import "context"

func (repository *Repository) Statuses(ctx context.Context, actor string, ids []string) (map[string]string, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT title_id,status FROM library_entries WHERE user_id=$1 AND title_id=ANY($2::varchar[]) AND NOT is_removed`, actor, ids)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var id, status string
		if failure = rows.Scan(&id, &status); failure != nil {
			return nil, failure
		}
		result[id] = status
	}
	return result, rows.Err()
}
