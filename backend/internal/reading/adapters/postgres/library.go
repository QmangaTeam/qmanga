package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/reading/application"
	"strings"
)

func (repository *Repository) Library(ctx context.Context, actor string, filter application.LibraryFilter) (application.LibraryPage, error) {
	page := application.LibraryPage{Items: []application.Entry{}}
	transaction, failure := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if failure != nil {
		return page, failure
	}
	defer transaction.Rollback(ctx)
	const conditions = `library.user_id=$1 AND NOT library.is_removed AND ($2='' OR library.status=$2) AND (NOT $3 OR library.is_favorite) AND ($4='' OR lower(title.name_kk||' '||title.name_original||' '||coalesce(array_to_string(title.aliases,' '),'')) LIKE $5)`
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
	arguments := []any{actor, filter.Status, filter.Favorite, filter.Query, "%" + escaped + "%"}
	failure = transaction.QueryRow(ctx, `SELECT count(*) FROM library_entries library JOIN titles title ON title.id=library.title_id WHERE `+conditions, arguments...).Scan(&page.Total)
	if failure != nil {
		return page, failure
	}
	arguments = append(arguments, filter.After)
	rows, failure := transaction.Query(ctx, `SELECT `+libraryColumns+libraryJoins+` WHERE `+conditions+` AND ($6='' OR library.title_id<$6) ORDER BY library.title_id DESC LIMIT 101`, arguments...)
	if failure != nil {
		return page, failure
	}
	defer rows.Close()
	for rows.Next() {
		item, failure := scanEntry(rows)
		if failure != nil {
			return page, failure
		}
		page.Items = append(page.Items, item)
	}
	if failure = rows.Err(); failure != nil {
		return page, failure
	}
	if len(page.Items) > 100 {
		page.Items = page.Items[:100]
		page.Next = page.Items[99].TitleID
	}
	return page, transaction.Commit(ctx)
}

func (repository *Repository) Suggestion(ctx context.Context, actor, title string) (bool, error) {
	var eligible bool
	failure := repository.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM (SELECT chapter_id FROM reading_completions WHERE user_id=$1 AND title_id=$2 LIMIT 3) completed)=3 AND NOT EXISTS(SELECT 1 FROM library_entries WHERE user_id=$1 AND title_id=$2 AND NOT is_removed)`, actor, title).Scan(&eligible)
	return eligible, failure
}
