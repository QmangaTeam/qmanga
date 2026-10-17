package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/reading/application"
	"qmanga/internal/reading/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }

const progressColumns = `title_id,release_id,COALESCE(revision_id,''),COALESCE(page_id,''),fraction,furthest_chapter_order,version,device_id,sequence`

func scan(row pgx.Row) (domain.Progress, error) {
	var item domain.Progress
	failure := row.Scan(&item.TitleID, &item.ReleaseID, &item.RevisionID, &item.PageID, &item.Fraction, &item.FurthestOrder, &item.Version, &item.DeviceID, &item.Sequence)
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.Progress{}, nil
	}
	return item, failure
}
func (repository *Repository) Progress(ctx context.Context, actor, title string) (domain.Progress, error) {
	return scan(repository.pool.QueryRow(ctx, `SELECT `+progressColumns+` FROM reading_progress WHERE user_id=$1 AND title_id=$2`, actor, title))
}
func (repository *Repository) List(ctx context.Context, actor, after string) ([]application.Entry, error) {
	return repository.list(ctx, actor, after, "")
}
func (repository *Repository) Selection(ctx context.Context, actor, title string) (*application.Entry, error) {
	items, failure := repository.list(ctx, actor, "", title)
	if failure != nil || len(items) == 0 {
		return nil, failure
	}
	return &items[0], nil
}
func (repository *Repository) list(ctx context.Context, actor, after, title string) ([]application.Entry, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT library.title_id,title.name_kk,title.slug,library.status,COALESCE(library.preferred_team_id,''),library.is_favorite,library.notify_new_releases,library.updated_at,(SELECT COUNT(DISTINCT chapter.id) FROM chapters chapter JOIN releases release ON release.chapter_id=chapter.id JOIN revisions revision ON revision.id=release.current_revision_id AND revision.status='published' WHERE chapter.title_id=title.id AND release.status='published' AND chapter.order_num>COALESCE(progress.furthest_chapter_order,0) AND (library.preferred_team_id IS NULL OR release.team_id=library.preferred_team_id)) FROM library_entries library JOIN titles title ON title.id=library.title_id LEFT JOIN reading_progress progress ON progress.user_id=library.user_id AND progress.title_id=library.title_id WHERE library.user_id=$1 AND ($2='' OR library.title_id<$2) AND ($3='' OR library.title_id=$3) ORDER BY library.title_id DESC LIMIT 100`, actor, after, title)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Entry{}
	for rows.Next() {
		var item application.Entry
		if failure = rows.Scan(&item.TitleID, &item.Name, &item.Slug, &item.Status, &item.PreferredTeamID, &item.Favorite, &item.Notify, &item.UpdatedAt, &item.Unseen); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	if failure = rows.Err(); failure != nil {
		return nil, failure
	}
	rows.Close()
	for index := range items {
		items[index].Progress, failure = repository.Progress(ctx, actor, items[index].TitleID)
		if failure != nil {
			return nil, failure
		}
	}
	return items, nil
}
func (repository *Repository) Save(ctx context.Context, actor string, input application.Selection) error {
	result, failure := repository.pool.Exec(ctx, `INSERT INTO library_entries(id,user_id,title_id,status,is_favorite,preferred_team_id,notify_new_releases) SELECT $1::varchar,$2::varchar,title.id,$4::varchar,$5::boolean,NULLIF($6::varchar,''),$7::boolean FROM titles title WHERE title.id=$3 AND EXISTS(SELECT 1 FROM users WHERE id=$2::varchar) AND ($6::varchar='' OR EXISTS(SELECT 1 FROM projects WHERE title_id=title.id AND team_id=$6::varchar)) ON CONFLICT(user_id,title_id)DO UPDATE SET status=EXCLUDED.status,is_favorite=EXCLUDED.is_favorite,preferred_team_id=EXCLUDED.preferred_team_id,notify_new_releases=EXCLUDED.notify_new_releases,updated_at=now()`, uuid.NewString(), actor, input.TitleID, input.Status, input.Favorite, input.PreferredTeamID, input.Notify)
	if failure == nil && result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return failure
}
func (repository *Repository) Remove(ctx context.Context, actor, title string) error {
	_, failure := repository.pool.Exec(ctx, `DELETE FROM library_entries WHERE user_id=$1 AND title_id=$2`, actor, title)
	return failure
}
func (repository *Repository) Advance(ctx context.Context, actor string, event domain.Event) (application.Result, error) {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return application.Result{}, failure
	}
	defer transaction.Rollback(ctx)
	var locked string
	if failure = transaction.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, actor).Scan(&locked); failure != nil {
		return application.Result{}, domain.ErrForbidden
	}
	current, failure := scan(transaction.QueryRow(ctx, `SELECT `+progressColumns+` FROM reading_progress WHERE user_id=$1 AND title_id=$2 FOR UPDATE`, actor, event.TitleID))
	if failure != nil {
		return application.Result{}, failure
	}
	encoded, _ := json.Marshal(event)
	digest := sha256.Sum256(encoded)
	requestHash := hex.EncodeToString(digest[:])
	var storedHash, eventTitle string
	var acceptedVersion int64
	failure = transaction.QueryRow(ctx, `SELECT title_id,COALESCE(request_hash,''),accepted_version FROM reading_events WHERE user_id=$1 AND event_id=$2`, actor, event.ID).Scan(&eventTitle, &storedHash, &acceptedVersion)
	if failure != nil && !errors.Is(failure, pgx.ErrNoRows) {
		return application.Result{}, failure
	}
	if failure == nil {
		if eventTitle != event.TitleID || storedHash != "" && storedHash != requestHash {
			return application.Result{}, domain.ErrInvalid
		}
		outcome := "duplicate"
		if acceptedVersion != current.Version || storedHash == "" {
			outcome = "conflict"
		}
		return application.Result{Progress: current, Outcome: outcome}, transaction.Commit(ctx)
	}
	var sequence int64
	failure = transaction.QueryRow(ctx, `SELECT last_sequence FROM reading_devices WHERE user_id=$1 AND title_id=$2 AND device_id=$3`, actor, event.TitleID, event.DeviceID).Scan(&sequence)
	if failure != nil && !errors.Is(failure, pgx.ErrNoRows) {
		return application.Result{}, failure
	}
	if sequence >= event.Sequence {
		return application.Result{Progress: current, Outcome: "stale"}, transaction.Commit(ctx)
	}
	var chapterOrder, pagePosition, lastPosition int
	failure = transaction.QueryRow(ctx, `SELECT chapter.order_num,page.position,(SELECT MAX(position) FROM pages WHERE revision_id=revision.id) FROM releases release JOIN chapters chapter ON chapter.id=release.chapter_id JOIN revisions revision ON revision.id=release.current_revision_id JOIN pages page ON page.revision_id=revision.id WHERE release.id=$1 AND chapter.title_id=$2 AND revision.id=$3 AND page.id=$4 AND release.status='published' AND revision.status='published'`, event.ReleaseID, event.TitleID, event.RevisionID, event.PageID).Scan(&chapterOrder, &pagePosition, &lastPosition)
	if errors.Is(failure, pgx.ErrNoRows) {
		return application.Result{}, domain.ErrNotFound
	}
	if failure != nil {
		return application.Result{}, failure
	}
	next, failure := domain.Advance(current, event, chapterOrder, pagePosition == lastPosition)
	if errors.Is(failure, domain.ErrConflict) {
		return application.Result{Progress: current, Outcome: "conflict"}, transaction.Commit(ctx)
	}
	if failure != nil {
		return application.Result{}, failure
	}
	_, failure = transaction.Exec(ctx, `INSERT INTO reading_progress(id,user_id,title_id,release_id,revision_id,page_id,page_position,fraction,furthest_chapter_order,version,device_id,sequence) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(user_id,title_id)DO UPDATE SET release_id=EXCLUDED.release_id,revision_id=EXCLUDED.revision_id,page_id=EXCLUDED.page_id,page_position=EXCLUDED.page_position,fraction=EXCLUDED.fraction,furthest_chapter_order=EXCLUDED.furthest_chapter_order,version=EXCLUDED.version,device_id=EXCLUDED.device_id,sequence=EXCLUDED.sequence,updated_at=now()`, uuid.NewString(), actor, event.TitleID, next.ReleaseID, next.RevisionID, next.PageID, pagePosition, next.Fraction, next.FurthestOrder, next.Version, next.DeviceID, next.Sequence)
	if failure != nil {
		return application.Result{}, failure
	}
	if _, failure = transaction.Exec(ctx, `INSERT INTO reading_devices(user_id,title_id,device_id,last_sequence)VALUES($1,$2,$3,$4)ON CONFLICT(user_id,title_id,device_id)DO UPDATE SET last_sequence=EXCLUDED.last_sequence`, actor, event.TitleID, event.DeviceID, event.Sequence); failure != nil {
		return application.Result{}, failure
	}
	if _, failure = transaction.Exec(ctx, `INSERT INTO reading_events(user_id,event_id,title_id,request_hash,accepted_version)VALUES($1,$2,$3,$4,$5)`, actor, event.ID, event.TitleID, requestHash, next.Version); failure != nil {
		return application.Result{}, failure
	}
	return application.Result{Progress: next, Outcome: "accepted"}, transaction.Commit(ctx)
}
