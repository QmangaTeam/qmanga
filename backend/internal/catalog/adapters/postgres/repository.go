package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const publicationJoins = ` FROM chapters chapter JOIN releases publication ON publication.chapter_id=chapter.id
 JOIN revisions revision ON revision.id=publication.current_revision_id AND revision.release_id=publication.id
 JOIN projects project ON project.id=publication.project_id AND project.title_id=chapter.title_id AND project.team_id=publication.team_id
 JOIN teams team ON team.id=publication.team_id `
const publicationVisible = `publication.status='published' AND publication.language='kk' AND revision.status='published'`
const titleVisible = `EXISTS(SELECT 1` + publicationJoins + `WHERE chapter.title_id=title.id AND ` + publicationVisible + `)`
const summaryColumns = `title.id,title.slug,title.name_kk,title.type,title.status_original,title.status_translation,title.cover_asset_id,title.is_demo,title.updated_at,
 (SELECT count(DISTINCT chapter.id)` + publicationJoins + `WHERE chapter.title_id=title.id AND ` + publicationVisible + `),
 (SELECT count(*) FROM ratings rating WHERE rating.title_id=title.id),
 coalesce((SELECT avg(rating.score)::float8 FROM ratings rating WHERE rating.title_id=title.id),0)`

func translate(failure error) error {
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var databaseError *pgconn.PgError
	if errors.As(failure, &databaseError) {
		switch databaseError.Code {
		case "23505":
			return domain.ErrConflict
		case "23503", "23514", "22001":
			return domain.ErrInvalid
		}
	}
	return failure
}

func scanSummary(row pgx.Row) (application.Summary, error) {
	var item application.Summary
	failure := row.Scan(&item.ID, &item.Slug, &item.Name, &item.Kind, &item.OriginalStatus, &item.TranslationStatus, &item.CoverID, &item.Demo, &item.UpdatedAt, &item.ChapterCount, &item.RatingCount, &item.Score)
	return item, failure
}

func (repository *Repository) List(ctx context.Context, filter application.Filter) (application.Page, error) {
	arguments := make([]any, 0)
	conditions := []string{titleVisible}
	bind := func(value any) string {
		arguments = append(arguments, value)
		return fmt.Sprintf("$%d", len(arguments))
	}
	if filter.Query != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
		conditions = append(conditions, "title.search_text LIKE "+bind("%"+escaped+"%"))
	}
	for _, constraint := range []struct{ column, value string }{{"title.type", filter.Kind}, {"title.status_original", filter.OriginalStatus}, {"title.status_translation", filter.TranslationStatus}} {
		if constraint.value != "" {
			conditions = append(conditions, constraint.column+"="+bind(constraint.value))
		}
	}
	if filter.Genre != "" {
		conditions = append(conditions, `EXISTS(SELECT 1 FROM title_genres mapping JOIN genres genre ON genre.id=mapping.genre_id WHERE mapping.title_id=title.id AND genre.slug=`+bind(filter.Genre)+`)`)
	}
	orderBy := "title.updated_at DESC,title.id DESC"
	if filter.Sort == "name" {
		orderBy = `title.name_kk COLLATE "C",title.id`
	}
	if filter.After != nil {
		if filter.Sort == "name" {
			conditions = append(conditions, `(title.name_kk COLLATE "C",title.id) > (`+bind(filter.After.Name)+` COLLATE "C",`+bind(filter.After.ID)+`)`)
		} else {
			conditions = append(conditions, `(title.updated_at,title.id) < (`+bind(filter.After.UpdatedAt)+`,`+bind(filter.After.ID)+`)`)
		}
	}
	query := "SELECT " + summaryColumns + " FROM titles title WHERE " + strings.Join(conditions, " AND ") + " ORDER BY " + orderBy + " LIMIT " + bind(filter.Limit+1)
	rows, failure := repository.pool.Query(ctx, query, arguments...)
	if failure != nil {
		return application.Page{}, failure
	}
	defer rows.Close()
	page := application.Page{Items: []application.Summary{}}
	for rows.Next() {
		item, failure := scanSummary(rows)
		if failure != nil {
			return page, failure
		}
		page.Items = append(page.Items, item)
	}
	if failure = rows.Err(); failure != nil {
		return page, failure
	}
	if len(page.Items) > filter.Limit {
		page.Items = page.Items[:filter.Limit]
		last := page.Items[len(page.Items)-1]
		page.Next = &application.Cursor{UpdatedAt: last.UpdatedAt, Name: last.Name, ID: last.ID}
	}
	return page, nil
}

func (repository *Repository) Detail(ctx context.Context, slug string) (application.Detail, error) {
	return repository.detail(ctx, "title.slug=$1 AND "+titleVisible, slug, true)
}
func (repository *Repository) AdminDetail(ctx context.Context, actorID, id string) (application.Detail, error) {
	allowed, failure := repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return application.Detail{}, failure
	}
	if !allowed {
		return application.Detail{}, domain.ErrForbidden
	}
	return repository.detail(ctx, "title.id=$1", id, false)
}
func (repository *Repository) detail(ctx context.Context, condition, id string, public bool) (application.Detail, error) {
	detail := application.Detail{Genres: []application.Genre{}, Projects: []application.Project{}}
	title := &detail.Title
	failure := repository.pool.QueryRow(ctx, `SELECT title.id,title.slug,title.name_kk,title.name_original,title.aliases,title.type,title.status_original,title.status_translation,title.description_kk,title.direction,title.is_demo,title.cover_asset_id,
 coalesce((SELECT avg(score)::float8 FROM ratings WHERE title_id=title.id),0),(SELECT count(*) FROM ratings WHERE title_id=title.id) FROM titles title WHERE `+condition, id).Scan(&title.ID, &title.Slug, &title.Name, &title.OriginalName, &title.Aliases, &title.Kind, &title.OriginalStatus, &title.TranslationStatus, &title.Description, &title.Direction, &title.Demo, &detail.CoverID, &detail.Score, &detail.RatingCount)
	if failure != nil {
		return detail, translate(failure)
	}
	rows, failure := repository.pool.Query(ctx, `SELECT genre.id,genre.name_kk,genre.slug FROM genres genre JOIN title_genres mapping ON mapping.genre_id=genre.id WHERE mapping.title_id=$1 ORDER BY genre.name_kk`, title.ID)
	if failure != nil {
		return detail, failure
	}
	for rows.Next() {
		var genre application.Genre
		if failure = rows.Scan(&genre.ID, &genre.Name, &genre.Slug); failure != nil {
			rows.Close()
			return detail, failure
		}
		detail.Genres = append(detail.Genres, genre)
		title.GenreIDs = append(title.GenreIDs, genre.ID)
	}
	failure = rows.Err()
	rows.Close()
	if failure != nil {
		return detail, failure
	}
	projectCondition := `outer_project.title_id=$1`
	if public {
		projectCondition += ` AND EXISTS(SELECT 1` + publicationJoins + ` WHERE publication.project_id=outer_project.id AND ` + publicationVisible + `)`
	}
	detail.Projects, failure = repository.projects(ctx, projectCondition, title.ID)
	return detail, failure
}

func (repository *Repository) UpdateTitle(ctx context.Context, actorID string, title domain.Title) error {
	return repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		search := strings.ToLower(title.Name + " " + title.OriginalName + " " + strings.Join(title.Aliases, " "))
		result, failure := transaction.Exec(ctx, `UPDATE titles SET slug=$2,name_kk=$3,name_original=$4,aliases=$5,type=$6,status_original=$7,status_translation=$8,description_kk=$9,direction=$10,is_demo=$11,search_text=$12,updated_at=now() WHERE id=$1`, title.ID, title.Slug, title.Name, title.OriginalName, title.Aliases, title.Kind, title.OriginalStatus, title.TranslationStatus, title.Description, title.Direction, title.Demo, search)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		if _, failure = transaction.Exec(ctx, "DELETE FROM title_genres WHERE title_id=$1", title.ID); failure != nil {
			return failure
		}
		for _, genreID := range title.GenreIDs {
			if _, failure = transaction.Exec(ctx, "INSERT INTO title_genres(title_id,genre_id) VALUES($1,$2)", title.ID, genreID); failure != nil {
				return failure
			}
		}
		_, failure = transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id) VALUES($1,$2,'update','title',$3)", uuid.NewString(), actorID, title.ID)
		return failure
	})
}
func (repository *Repository) ApprovedTeams(ctx context.Context, actorID string) ([]application.Team, error) {
	allowed, failure := repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return nil, failure
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,name,slug,description FROM teams WHERE is_approved ORDER BY name LIMIT 100")
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Team{}
	for rows.Next() {
		var item application.Team
		if failure = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) Chapters(ctx context.Context, slug string, after, limit int) (application.ChapterPage, error) {
	page := application.ChapterPage{Items: []application.Chapter{}}
	var titleID string
	failure := repository.pool.QueryRow(ctx, `SELECT title.id FROM titles title WHERE title.slug=$1 AND `+titleVisible, slug).Scan(&titleID)
	if failure != nil {
		return page, translate(failure)
	}
	rows, failure := repository.pool.Query(ctx, `SELECT outer_chapter.id,outer_chapter.label,outer_chapter.volume,outer_chapter.title_kk,outer_chapter.order_num FROM chapters outer_chapter WHERE outer_chapter.title_id=$1 AND outer_chapter.order_num>$2 AND EXISTS(SELECT 1`+publicationJoins+`WHERE chapter.id=outer_chapter.id AND `+publicationVisible+`) ORDER BY outer_chapter.order_num LIMIT $3`, titleID, after, limit+1)
	if failure != nil {
		return page, failure
	}
	for rows.Next() {
		item := application.Chapter{Releases: []application.Release{}}
		if failure = rows.Scan(&item.ID, &item.Label, &item.Volume, &item.Name, &item.Order); failure != nil {
			rows.Close()
			return page, failure
		}
		page.Items = append(page.Items, item)
	}
	failure = rows.Err()
	rows.Close()
	if failure != nil {
		return page, failure
	}
	if len(page.Items) > limit {
		page.Items = page.Items[:limit]
		order := page.Items[len(page.Items)-1].Order
		page.NextOrder = &order
	}
	chapterIDs := make([]string, 0, len(page.Items))
	positions := make(map[string]int)
	for position, item := range page.Items {
		chapterIDs = append(chapterIDs, item.ID)
		positions[item.ID] = position
	}
	rows, failure = repository.pool.Query(ctx, `SELECT chapter.id,publication.id,publication.project_id,team.id,team.name,team.slug,revision.updated_at`+publicationJoins+`WHERE chapter.id=ANY($1::text[]) AND `+publicationVisible+` ORDER BY team.name,publication.id`, chapterIDs)
	if failure != nil {
		return page, failure
	}
	defer rows.Close()
	for rows.Next() {
		var chapterID string
		var item application.Release
		if failure = rows.Scan(&chapterID, &item.ID, &item.ProjectID, &item.TeamID, &item.TeamName, &item.TeamSlug, &item.PublishedAt); failure != nil {
			return page, failure
		}
		position := positions[chapterID]
		page.Items[position].Releases = append(page.Items[position].Releases, item)
	}
	return page, rows.Err()
}

func (repository *Repository) Genres(ctx context.Context) ([]application.Genre, error) {
	rows, failure := repository.pool.Query(ctx, "SELECT id,name_kk,slug FROM genres ORDER BY name_kk")
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Genre{}
	for rows.Next() {
		var item application.Genre
		if failure = rows.Scan(&item.ID, &item.Name, &item.Slug); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) Team(ctx context.Context, slug string) (application.Team, error) {
	var team application.Team
	failure := repository.pool.QueryRow(ctx, "SELECT id,slug,name,description FROM teams WHERE slug=$1 AND is_approved", slug).Scan(&team.ID, &team.Slug, &team.Name, &team.Description)
	if failure != nil {
		return team, translate(failure)
	}
	team.Projects, failure = repository.projects(ctx, `outer_project.team_id=$1 AND EXISTS(SELECT 1`+publicationJoins+` WHERE publication.project_id=outer_project.id AND `+publicationVisible+`)`, team.ID)
	return team, failure
}

func (repository *Repository) projects(ctx context.Context, condition string, arguments ...any) ([]application.Project, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT outer_project.id,outer_project.title_id,outer_project.team_id,title.name_kk,title.slug,team.name,team.slug,outer_project.status FROM projects outer_project JOIN titles title ON title.id=outer_project.title_id JOIN teams team ON team.id=outer_project.team_id WHERE `+condition+` ORDER BY title.name_kk,team.name LIMIT 100`, arguments...)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Project{}
	for rows.Next() {
		var item application.Project
		if failure = rows.Scan(&item.ID, &item.TitleID, &item.TeamID, &item.TitleName, &item.TitleSlug, &item.TeamName, &item.TeamSlug, &item.Status); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) StudioProjects(ctx context.Context, actorID string) ([]application.Project, error) {
	return repository.projects(ctx, `EXISTS(SELECT 1 FROM team_memberships membership JOIN users account ON account.id=membership.user_id WHERE membership.team_id=outer_project.team_id AND account.id=$1 AND account.is_verified)`, actorID)
}

func (repository *Repository) IsAdmin(ctx context.Context, actorID string) (bool, error) {
	var allowed bool
	failure := repository.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND role='admin' AND is_verified)", actorID).Scan(&allowed)
	return allowed, failure
}

func (repository *Repository) withAdmin(ctx context.Context, actorID string, mutation func(pgx.Tx) error) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var allowed bool
	failure = transaction.QueryRow(ctx, "SELECT role='admin' AND is_verified FROM users WHERE id=$1 FOR SHARE", actorID).Scan(&allowed)
	if errors.Is(failure, pgx.ErrNoRows) || (failure == nil && !allowed) {
		return domain.ErrForbidden
	}
	if failure != nil {
		return failure
	}
	if failure = mutation(transaction); failure != nil {
		return translate(failure)
	}
	return transaction.Commit(ctx)
}

func audit(ctx context.Context, transaction pgx.Tx, actorID, kind, entityID string) error {
	_, failure := transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id) VALUES($1,$2,'create',$3,$4)", uuid.NewString(), actorID, kind, entityID)
	return failure
}

func (repository *Repository) CreateTitle(ctx context.Context, actorID string, title domain.Title) error {
	return repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		search := strings.ToLower(title.Name + " " + title.OriginalName + " " + strings.Join(title.Aliases, " "))
		_, failure := transaction.Exec(ctx, `INSERT INTO titles(id,slug,name_kk,name_original,aliases,type,status_original,status_translation,description_kk,direction,is_demo,search_text) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, title.ID, title.Slug, title.Name, title.OriginalName, title.Aliases, title.Kind, title.OriginalStatus, title.TranslationStatus, title.Description, title.Direction, title.Demo, search)
		if failure != nil {
			return failure
		}
		for _, genreID := range title.GenreIDs {
			if _, failure = transaction.Exec(ctx, "INSERT INTO title_genres(title_id,genre_id) VALUES($1,$2)", title.ID, genreID); failure != nil {
				return failure
			}
		}
		return audit(ctx, transaction, actorID, "title", title.ID)
	})
}

func (repository *Repository) CreateGenre(ctx context.Context, actorID string, genre application.Genre) error {
	return repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		if _, failure := transaction.Exec(ctx, "INSERT INTO genres(id,name_kk,slug) VALUES($1,$2,$3)", genre.ID, genre.Name, genre.Slug); failure != nil {
			return failure
		}
		return audit(ctx, transaction, actorID, "genre", genre.ID)
	})
}

func (repository *Repository) CreateChapter(ctx context.Context, actorID, titleID, label, volume, name string, order int) (string, error) {
	id := uuid.NewString()
	failure := repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		if _, failure := transaction.Exec(ctx, "INSERT INTO chapters(id,title_id,label,order_num,volume,title_kk) VALUES($1,$2,$3,$4,$5,$6)", id, titleID, label, order, volume, name); failure != nil {
			return failure
		}
		return audit(ctx, transaction, actorID, "chapter", id)
	})
	return id, failure
}

func (repository *Repository) CreateProject(ctx context.Context, actorID, id, titleID, teamID string) error {
	return repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		var approved bool
		failure := transaction.QueryRow(ctx, "SELECT is_approved FROM teams WHERE id=$1 FOR SHARE", teamID).Scan(&approved)
		if errors.Is(failure, pgx.ErrNoRows) || (failure == nil && !approved) {
			return domain.ErrForbidden
		}
		if failure != nil {
			return failure
		}
		if _, failure = transaction.Exec(ctx, "INSERT INTO projects(id,title_id,team_id) VALUES($1,$2,$3)", id, titleID, teamID); failure != nil {
			return failure
		}
		return audit(ctx, transaction, actorID, "project", id)
	})
}

func (repository *Repository) AdminTitles(ctx context.Context, actorID string) ([]application.Summary, error) {
	allowed, failure := repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return nil, failure
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT "+summaryColumns+" FROM titles title ORDER BY title.updated_at DESC,title.id DESC LIMIT 100")
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
