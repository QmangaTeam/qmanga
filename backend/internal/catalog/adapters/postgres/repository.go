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
const hasChapters = `EXISTS(SELECT 1` + publicationJoins + `WHERE chapter.title_id=title.id AND ` + publicationVisible + `)`
const titleChapterCount = `(SELECT count(DISTINCT chapter.id)` + publicationJoins + `WHERE chapter.title_id=title.id AND ` + publicationVisible + `)`
const publicationYear = `COALESCE(CASE WHEN title.publication->>'year' ~ '^[0-9]{1,4}$' THEN (title.publication->>'year')::int ELSE 0 END,0)`
const titleVisible = "(title.is_listed OR " + hasChapters + ")"
const summaryColumns = `title.id,title.slug,title.name_kk,title.type,title.status_original,title.status_translation,title.cover_asset_id,title.is_demo,title.created_at,title.updated_at,
` + titleChapterCount + `,
 (SELECT count(*) FROM ratings rating JOIN users rating_user ON rating_user.id=rating.user_id AND NOT rating_user.is_blocked WHERE rating.title_id=title.id) AS rating_count,
 coalesce((SELECT avg(rating.score)::float8 FROM ratings rating JOIN users rating_user ON rating_user.id=rating.user_id AND NOT rating_user.is_blocked WHERE rating.title_id=title.id),0) AS score,title.age_rating`

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
	failure := row.Scan(&item.ID, &item.Slug, &item.Name, &item.Kind, &item.OriginalStatus, &item.TranslationStatus, &item.CoverID, &item.Demo, &item.CreatedAt, &item.UpdatedAt, &item.ChapterCount, &item.RatingCount, &item.Score, &item.AgeRating)
	return item, failure
}

func (repository *Repository) List(ctx context.Context, filter application.Filter) (application.Page, error) {
	arguments := make([]any, 0)
	conditions := []string{titleVisible}
	if filter.HasChapters {
		conditions = append(conditions, hasChapters)
	}
	if filter.Featured {
		conditions = append(conditions, "title.is_featured")
	}
	bind := func(value any) string {
		arguments = append(arguments, value)
		return fmt.Sprintf("$%d", len(arguments))
	}
	if len(filter.AgeRatings) > 0 {
		ratings := make([]int32, len(filter.AgeRatings))
		for index, rating := range filter.AgeRatings {
			ratings[index] = int32(rating)
		}
		conditions = append(conditions, "title.age_rating = ANY("+bind(ratings)+"::int[])")
	}
	if filter.Query != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
		pattern := "%" + escaped + "%"
		conditions = append(conditions, "(title.search_text LIKE "+bind(pattern)+" OR EXISTS(SELECT 1 FROM title_genres mapping JOIN genres genre ON genre.id=mapping.genre_id WHERE mapping.title_id=title.id AND (lower(genre.name_kk) LIKE "+bind(pattern)+" OR lower(genre.name_ru) LIKE "+bind(pattern)+")))")
	}
	for _, constraint := range []struct {
		column string
		values []string
	}{{"title.type", filter.Kinds}, {"title.status_original", filter.OriginalStatuses}, {"title.status_translation", filter.TranslationStatuses}} {
		if len(constraint.values) > 0 {
			conditions = append(conditions, constraint.column+" = ANY("+bind(constraint.values)+"::text[])")
		}
	}
	if len(filter.Formats) > 0 {
		conditions = append(conditions, "title.publication->'formats' ?| "+bind(filter.Formats)+"::text[]")
	}
	if len(filter.Sources) > 0 {
		conditions = append(conditions, "title.publication->>'source' = ANY("+bind(filter.Sources)+"::text[])")
	}
	if len(filter.Countries) > 0 {
		conditions = append(conditions, "title.publication->>'country' = ANY("+bind(filter.Countries)+"::text[])")
	}
	if filter.ChaptersMin > 0 {
		conditions = append(conditions, titleChapterCount+" >= "+bind(filter.ChaptersMin))
	}
	if filter.ChaptersMax > 0 {
		conditions = append(conditions, titleChapterCount+" <= "+bind(filter.ChaptersMax))
	}
	if filter.YearMin > 0 || filter.YearMax > 0 {
		conditions = append(conditions, publicationYear+" > 0")
	}
	if filter.YearMin > 0 {
		conditions = append(conditions, publicationYear+" >= "+bind(filter.YearMin))
	}
	if filter.YearMax > 0 {
		conditions = append(conditions, publicationYear+" <= "+bind(filter.YearMax))
	}
	for _, slug := range filter.Genres {
		conditions = append(conditions, `EXISTS(SELECT 1 FROM title_genres mapping JOIN genres genre ON genre.id=mapping.genre_id WHERE mapping.title_id=title.id AND genre.slug=`+bind(slug)+`)`)
	}
	if filter.Category != "" {
		conditions = append(conditions, `EXISTS(SELECT 1 FROM title_categories mapping JOIN categories category ON category.id=mapping.category_id WHERE mapping.title_id=title.id AND category.slug=`+bind(filter.Category)+`)`)
	}
	if filter.Collection != "" {
		conditions = append(conditions, `EXISTS(SELECT 1 FROM collection_titles mapping JOIN editorial_collections collection ON collection.id=mapping.collection_id WHERE mapping.title_id=title.id AND collection.is_visible AND collection.slug=`+bind(filter.Collection)+`)`)
	}
	timestamp := "title.updated_at"
	if filter.Sort == "new" {
		timestamp = "title.created_at"
	}
	orderBy := timestamp + " DESC,title.id DESC"
	if filter.Sort == "name" {
		orderBy = `title.name_kk COLLATE "C",title.id`
	}
	if filter.After != nil {
		if filter.Sort == "name" {
			conditions = append(conditions, `(title.name_kk COLLATE "C",title.id) > (`+bind(filter.After.Name)+` COLLATE "C",`+bind(filter.After.ID)+`)`)
		} else {
			conditions = append(conditions, `(`+timestamp+`,title.id) < (`+bind(filter.After.UpdatedAt)+`,`+bind(filter.After.ID)+`)`)
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
		position := last.UpdatedAt
		if filter.Sort == "new" {
			position = last.CreatedAt
		}
		page.Next = &application.Cursor{UpdatedAt: position, Name: last.Name, ID: last.ID}
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
	failure := repository.pool.QueryRow(ctx, `SELECT title.id,title.slug,title.name_kk,title.name_original,title.aliases,title.type,title.status_original,title.status_translation,title.description_kk,title.direction,title.is_demo,title.authors,title.is_featured,title.is_listed,title.cover_asset_id,title.banner_asset_id,title.publication,title.age_rating,
 coalesce((SELECT avg(score)::float8 FROM ratings JOIN users rating_user ON rating_user.id=ratings.user_id AND NOT rating_user.is_blocked WHERE title_id=title.id),0),(SELECT count(*) FROM ratings JOIN users rating_user ON rating_user.id=ratings.user_id AND NOT rating_user.is_blocked WHERE title_id=title.id) FROM titles title WHERE `+condition, id).Scan(&title.ID, &title.Slug, &title.Name, &title.OriginalName, &title.Aliases, &title.Kind, &title.OriginalStatus, &title.TranslationStatus, &title.Description, &title.Direction, &title.Demo, &title.Authors, &title.Featured, &title.Listed, &detail.CoverID, &detail.BannerID, &title.Publication, &title.AgeRating, &detail.Score, &detail.RatingCount)
	if failure != nil {
		return detail, translate(failure)
	}
	ratingRows, failure := repository.pool.Query(ctx, `SELECT rating.score,count(*) FROM ratings rating JOIN users rating_user ON rating_user.id=rating.user_id AND NOT rating_user.is_blocked WHERE rating.title_id=$1 GROUP BY rating.score`, title.ID)
	if failure != nil {
		return detail, failure
	}
	for ratingRows.Next() {
		var score, count int
		if failure = ratingRows.Scan(&score, &count); failure != nil {
			ratingRows.Close()
			return detail, failure
		}
		if score >= 1 && score <= 10 {
			detail.RatingDistribution[score-1] = count
		}
	}
	failure = ratingRows.Err()
	ratingRows.Close()
	if failure != nil {
		return detail, failure
	}
	rows, failure := repository.pool.Query(ctx, `SELECT genre.id,genre.name_kk,genre.name_ru,genre.slug,COALESCE(category.slug,''),genre.parent_slug,genre.position FROM genres genre LEFT JOIN categories category ON category.id=genre.category_id JOIN title_genres mapping ON mapping.genre_id=genre.id WHERE mapping.title_id=$1 ORDER BY genre.name_kk`, title.ID)
	if failure != nil {
		return detail, failure
	}
	for rows.Next() {
		var genre application.Genre
		if failure = rows.Scan(&genre.ID, &genre.Name, &genre.NameRU, &genre.Slug, &genre.CategorySlug, &genre.ParentSlug, &genre.Position); failure != nil {
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
	detail.Categories = []application.Genre{}
	detail.Collections = []domain.Collection{}
	categoryRows, categoryError := repository.pool.Query(ctx, "SELECT category.id,category.name_kk,category.name_ru,category.slug FROM categories category JOIN title_categories mapping ON mapping.category_id=category.id WHERE mapping.title_id=$1 ORDER BY category.position,category.id", title.ID)
	if categoryError != nil {
		return detail, categoryError
	}
	for categoryRows.Next() {
		var item application.Genre
		if categoryError = categoryRows.Scan(&item.ID, &item.Name, &item.NameRU, &item.Slug); categoryError != nil {
			categoryRows.Close()
			return detail, categoryError
		}
		detail.Categories = append(detail.Categories, item)
	}
	categoryError = categoryRows.Err()
	categoryRows.Close()
	if categoryError != nil {
		return detail, categoryError
	}
	collectionRows, collectionError := repository.pool.Query(ctx, `SELECT collection.id,collection.name_kk,collection.slug,collection.description,collection.position,collection.is_visible FROM editorial_collections collection JOIN collection_titles mapping ON mapping.collection_id=collection.id WHERE mapping.title_id=$1 AND ($2 OR collection.is_visible) ORDER BY collection.position,collection.id`, title.ID, !public)
	if collectionError != nil {
		return detail, collectionError
	}
	for collectionRows.Next() {
		var item domain.Collection
		if collectionError = collectionRows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Position, &item.Visible); collectionError != nil {
			collectionRows.Close()
			return detail, collectionError
		}
		item.TitleIDs = []string{}
		detail.Collections = append(detail.Collections, item)
	}
	collectionError = collectionRows.Err()
	collectionRows.Close()
	if collectionError != nil {
		return detail, collectionError
	}
	projectCondition := `outer_project.title_id=$1`
	failure = repository.pool.QueryRow(ctx, `SELECT
 (SELECT count(DISTINCT chapter.id)`+publicationJoins+`WHERE chapter.title_id=$1 AND `+publicationVisible+`),
 (SELECT count(*) FROM reading_progress progress JOIN users reader ON reader.id=progress.user_id AND NOT reader.is_blocked WHERE progress.title_id=$1),
 (SELECT count(*) FROM comments WHERE title_id=$1 AND release_id IS NULL AND NOT is_hidden AND deleted_at IS NULL)`, title.ID).Scan(&detail.ChapterCount, &detail.ReaderCount, &detail.CommentCount)
	if failure != nil {
		return detail, failure
	}
	if public {
		projectCondition += ` AND EXISTS(SELECT 1` + publicationJoins + ` WHERE publication.project_id=outer_project.id AND ` + publicationVisible + `)`
	}
	detail.Projects, failure = repository.projects(ctx, 0, projectCondition, title.ID)
	return detail, failure
}

func (repository *Repository) UpdateTitle(ctx context.Context, actorID string, title domain.Title) error {
	return repository.withAdmin(ctx, actorID, func(transaction pgx.Tx) error {
		search := strings.ToLower(title.Name + " " + title.OriginalName + " " + strings.Join(title.Aliases, " "))
		result, failure := transaction.Exec(ctx, `UPDATE titles SET slug=$2,name_kk=$3,name_original=$4,aliases=$5,type=$6,status_original=$7,status_translation=$8,description_kk=$9,direction=$10,is_demo=$11,search_text=$12,authors=$13,is_featured=$14,is_listed=$15,publication=$16,age_rating=COALESCE($17::smallint,age_rating),updated_at=now() WHERE id=$1`, title.ID, title.Slug, title.Name, title.OriginalName, title.Aliases, title.Kind, title.OriginalStatus, title.TranslationStatus, title.Description, title.Direction, title.Demo, search, title.Authors, title.Featured, title.Listed, title.Publication, title.AgeRating)
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
func (repository *Repository) ApprovedTeams(ctx context.Context, actorID string, offset ...int) ([]application.Team, error) {
	allowed, failure := repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return nil, failure
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,name,slug,description FROM teams WHERE is_approved ORDER BY name,id LIMIT 100 OFFSET "+pageOffset(offset))
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

func (repository *Repository) Chapters(ctx context.Context, slug string, filter application.ChapterFilter) (application.ChapterPage, error) {
	after, limit := filter.After, filter.Limit
	page := application.ChapterPage{Items: []application.Chapter{}}
	var titleID string
	failure := repository.pool.QueryRow(ctx, `SELECT title.id FROM titles title WHERE title.slug=$1 AND `+titleVisible, slug).Scan(&titleID)
	if failure != nil {
		return page, translate(failure)
	}
	comparison, direction := ">", "ASC"
	if filter.Sort == "desc" {
		comparison, direction = "<", "DESC"
	}
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
	rows, failure := repository.pool.Query(ctx, `SELECT outer_chapter.id,outer_chapter.label,outer_chapter.volume,outer_chapter.title_kk,outer_chapter.order_num FROM chapters outer_chapter WHERE outer_chapter.title_id=$1 AND ($2=0 OR outer_chapter.order_num`+comparison+`$2) AND ($4='' OR lower(outer_chapter.label) LIKE $5 OR lower(outer_chapter.title_kk) LIKE $5) AND EXISTS(SELECT 1`+publicationJoins+`WHERE chapter.id=outer_chapter.id AND `+publicationVisible+`) ORDER BY outer_chapter.order_num `+direction+` LIMIT $3`, titleID, after, limit+1, filter.Query, "%"+escaped+"%")
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
	rows, failure := repository.pool.Query(ctx, "SELECT genre.id,genre.name_kk,genre.name_ru,genre.slug,COALESCE(category.slug,''),genre.parent_slug,genre.position FROM genres genre LEFT JOIN categories category ON category.id=genre.category_id ORDER BY COALESCE(category.position,999),genre.position,genre.name_kk")
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := []application.Genre{}
	for rows.Next() {
		var item application.Genre
		if failure = rows.Scan(&item.ID, &item.Name, &item.NameRU, &item.Slug, &item.CategorySlug, &item.ParentSlug, &item.Position); failure != nil {
			return nil, failure
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) Team(ctx context.Context, slug string) (application.Team, error) {
	var team application.Team
	failure := repository.pool.QueryRow(ctx, "SELECT id,slug,name,description,telegram_url FROM teams WHERE slug=$1 AND is_approved", slug).Scan(&team.ID, &team.Slug, &team.Name, &team.Description, &team.TelegramURL)
	if failure != nil {
		return team, translate(failure)
	}
	rows, failure := repository.pool.Query(ctx, "SELECT COALESCE(NULLIF(account.display_name,''),'Оқырман · '||right(account.id,6)),member.role FROM team_memberships member JOIN users account ON account.id=member.user_id WHERE member.team_id=$1 ORDER BY member.created_at,member.id LIMIT 200", team.ID)
	if failure != nil {
		return team, failure
	}
	team.Members = []application.PublicMember{}
	for rows.Next() {
		var member application.PublicMember
		if failure = rows.Scan(&member.Name, &member.Role); failure != nil {
			rows.Close()
			return team, failure
		}
		team.Members = append(team.Members, member)
	}
	failure = rows.Err()
	rows.Close()
	if failure != nil {
		return team, failure
	}
	team.Projects, failure = repository.projects(ctx, 0, `outer_project.team_id=$1 AND EXISTS(SELECT 1`+publicationJoins+` WHERE publication.project_id=outer_project.id AND `+publicationVisible+`)`, team.ID)
	return team, failure
}

func (repository *Repository) projects(ctx context.Context, offset int, condition string, arguments ...any) ([]application.Project, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT outer_project.id,outer_project.title_id,outer_project.team_id,title.name_kk,title.slug,team.name,team.slug,outer_project.status FROM projects outer_project JOIN titles title ON title.id=outer_project.title_id JOIN teams team ON team.id=outer_project.team_id WHERE `+condition+` ORDER BY title.name_kk,team.name,outer_project.id LIMIT 100 OFFSET `+pageOffset([]int{offset}), arguments...)
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

func (repository *Repository) StudioProjects(ctx context.Context, actorID string, offset ...int) ([]application.Project, error) {
	return repository.projects(ctx, offsetValue(offset), `EXISTS(SELECT 1 FROM team_memberships membership JOIN users account ON account.id=membership.user_id WHERE membership.team_id=outer_project.team_id AND account.id=$1 AND account.is_verified AND NOT account.is_blocked)`, actorID)
}

func (repository *Repository) IsAdmin(ctx context.Context, actorID string) (bool, error) {
	var allowed bool
	failure := repository.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND role='admin' AND is_verified AND NOT is_blocked)", actorID).Scan(&allowed)
	return allowed, failure
}

func (repository *Repository) withAdmin(ctx context.Context, actorID string, mutation func(pgx.Tx) error) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var allowed bool
	failure = transaction.QueryRow(ctx, "SELECT role='admin' AND is_verified AND NOT is_blocked FROM users WHERE id=$1 FOR SHARE", actorID).Scan(&allowed)
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
		_, failure := transaction.Exec(ctx, `INSERT INTO titles(id,slug,name_kk,name_original,aliases,type,status_original,status_translation,description_kk,direction,is_demo,search_text,authors,is_featured,is_listed,publication,age_rating) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,COALESCE($17::smallint,0))`, title.ID, title.Slug, title.Name, title.OriginalName, title.Aliases, title.Kind, title.OriginalStatus, title.TranslationStatus, title.Description, title.Direction, title.Demo, search, title.Authors, title.Featured, title.Listed, title.Publication, title.AgeRating)
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

func (repository *Repository) AdminTitles(ctx context.Context, actorID string, offset ...int) ([]application.Summary, error) {
	return repository.SearchAdminTitles(ctx, actorID, "", offset...)
}
func (repository *Repository) SearchAdminTitles(ctx context.Context, actorID, query string, offset ...int) ([]application.Summary, error) {
	allowed, failure := repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return nil, failure
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT "+summaryColumns+" FROM titles title WHERE ($1='' OR title.name_kk ILIKE '%'||$1||'%' OR title.original_name ILIKE '%'||$1||'%' OR title.slug ILIKE '%'||$1||'%' OR title.id=$1) ORDER BY title.updated_at DESC,title.id DESC LIMIT 100 OFFSET "+pageOffset(offset), query)
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
