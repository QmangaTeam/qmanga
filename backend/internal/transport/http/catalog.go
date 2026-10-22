package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
)

type CatalogHandler struct{ service *application.Service }

func NewCatalogHandler(service *application.Service) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (handler *CatalogHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	api.GET("/titles", handler.list)
	api.GET("/updates", handler.updates)
	api.GET("/rankings", handler.rankings)
	api.GET("/titles/:slug", handler.detail)
	api.GET("/titles/:slug/chapters", handler.chapters)
	api.GET("/genres", handler.genres)
	api.GET("/taxonomy", handler.taxonomy)
	api.GET("/teams/:slug", handler.team)
	api.GET("/studio/projects", identity.RequireUser(), handler.studioProjects)
	api.GET("/studio/projects/:id", identity.RequireUser(), handler.studioProject)
	api.POST("/studio/projects/:id/chapters", identity.RequireUser(), handler.studioCreateChapter)
	api.GET("/studio/project-requests", identity.RequireUser(), handler.projectRequests)
	api.POST("/studio/project-requests", identity.RequireUser(), handler.requestProject)
	api.POST("/studio/project-requests/:id/cancel", identity.RequireUser(), handler.cancelProjectRequest)
	admin := api.Group("/admin", identity.RequireUser(), RequireRole("admin"))
	admin.GET("/titles", handler.adminTitles)
	admin.GET("/titles/:id", handler.adminDetail)
	admin.GET("/projects/:id", handler.adminProject)
	admin.PUT("/titles/:id", handler.createTitle)
	admin.GET("/catalog-teams", handler.approvedTeams)
	admin.POST("/titles", handler.createTitle)
	admin.POST("/genres", handler.createGenre)
	admin.GET("/taxonomy", handler.taxonomy)
	for _, kind := range []string{"genres", "categories"} {
		if kind != "genres" {
			admin.POST("/"+kind, handler.saveTaxon(kind))
		}
		admin.PUT("/"+kind+"/:id", handler.saveTaxon(kind))
		admin.DELETE("/"+kind+"/:id", handler.deleteTaxon(kind))
	}
	admin.POST("/collections", handler.saveCollection)
	admin.PUT("/collections/:id", handler.saveCollection)
	admin.DELETE("/collections/:id", handler.deleteTaxon("collections"))
	admin.PUT("/titles/:id/categories", handler.setCategories)
	admin.POST("/titles/:id/chapters", handler.createChapter)
	admin.POST("/titles/:id/projects", handler.createProject)
	admin.GET("/project-requests", handler.projectRequests)
	admin.POST("/project-requests/:id/review", handler.reviewProjectRequest)
}

type titleSummaryDTO struct {
	AgeRating         int       `json:"age_rating"`
	ID                string    `json:"id"`
	Slug              string    `json:"slug"`
	Name              string    `json:"name"`
	Kind              string    `json:"kind"`
	OriginalStatus    string    `json:"original_status"`
	TranslationStatus string    `json:"translation_status"`
	CoverURL          string    `json:"cover_url"`
	Demo              bool      `json:"demo"`
	ChapterCount      int       `json:"chapter_count"`
	RatingCount       int       `json:"rating_count"`
	Score             float64   `json:"score"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func coverURL(id string) string {
	if id == "" {
		return ""
	}
	return "/api/v1/media/" + url.PathEscape(id)
}
func summaryView(item application.Summary) titleSummaryDTO {
	return titleSummaryDTO{AgeRating: item.AgeRating, ID: item.ID, Slug: item.Slug, Name: item.Name, Kind: item.Kind, OriginalStatus: item.OriginalStatus, TranslationStatus: item.TranslationStatus, CoverURL: coverURL(item.CoverID), Demo: item.Demo, ChapterCount: item.ChapterCount, RatingCount: item.RatingCount, Score: item.Score, UpdatedAt: item.UpdatedAt}
}

type catalogueCursor struct {
	Version  int                `json:"version"`
	Sort     string             `json:"sort"`
	Position application.Cursor `json:"position"`
}

func nextCursor(cursor *application.Cursor, sort string) *string {
	if cursor == nil {
		return nil
	}
	encoded, _ := json.Marshal(catalogueCursor{Version: 1, Sort: sort, Position: *cursor})
	value := base64.RawURLEncoding.EncodeToString(encoded)
	return &value
}
func queryInteger(request *gin.Context, key string, fallback, maximum int) (int, bool) {
	value := request.Query(key)
	if value == "" {
		return fallback, true
	}
	parsed, failure := strconv.Atoi(value)
	if failure != nil || parsed < 1 || parsed > maximum {
		Error(request, 400, "invalid_query", "Сұрау параметрін тексеріңіз")
		return 0, false
	}
	return parsed, true
}
func queryValues(request *gin.Context, key string) []string {
	values := request.QueryArray(key)
	selected := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			selected = append(selected, trimmed)
		}
	}
	return selected
}
func queryOptionalInteger(request *gin.Context, key string, minimum, maximum int) (int, bool) {
	raw := strings.TrimSpace(request.Query(key))
	if raw == "" {
		return 0, true
	}
	parsed, failure := strconv.Atoi(raw)
	if failure != nil || parsed < minimum || parsed > maximum {
		Error(request, 400, "invalid_query", "Сұрау параметрін тексеріңіз")
		return 0, false
	}
	return parsed, true
}
func (handler *CatalogHandler) list(request *gin.Context) {
	limit, valid := queryInteger(request, "limit", 24, 48)
	if !valid {
		return
	}
	filter := application.Filter{Query: request.Query("q"), Category: request.Query("category"), Collection: request.Query("collection"), Sort: request.DefaultQuery("sort", "updated"), Featured: request.Query("featured") == "true", HasChapters: request.Query("has_chapters") == "true", Limit: limit}
	filter.Kinds = queryValues(request, "kind")
	filter.OriginalStatuses = queryValues(request, "original_status")
	filter.TranslationStatuses = queryValues(request, "translation_status")
	filter.Genres = queryValues(request, "genre")
	filter.Formats = queryValues(request, "format")
	filter.Sources = queryValues(request, "source")
	filter.Countries = queryValues(request, "country")
	for _, raw := range queryValues(request, "age_rating") {
		age, failure := strconv.Atoi(raw)
		if failure != nil {
			Error(request, 400, "invalid_query", "Invalid age rating")
			return
		}
		filter.AgeRatings = append(filter.AgeRatings, age)
	}
	if filter.ChaptersMin, valid = queryOptionalInteger(request, "chapters_min", 0, 1_000_000); !valid {
		return
	}
	if filter.ChaptersMax, valid = queryOptionalInteger(request, "chapters_max", 0, 1_000_000); !valid {
		return
	}
	if filter.YearMin, valid = queryOptionalInteger(request, "year_min", 0, 2200); !valid {
		return
	}
	if filter.YearMax, valid = queryOptionalInteger(request, "year_max", 0, 2200); !valid {
		return
	}
	if raw := request.Query("cursor"); raw != "" {
		decoded, failure := base64.RawURLEncoding.DecodeString(raw)
		var cursor catalogueCursor
		if len(raw) > 2048 || failure != nil || json.Unmarshal(decoded, &cursor) != nil || cursor.Version != 1 || cursor.Sort != filter.Sort {
			Error(request, 400, "invalid_cursor", "Тізім сілтемесі жарамсыз")
			return
		}
		filter.After = &cursor.Position
	}
	page, failure := handler.service.List(request.Request.Context(), filter)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	items := make([]titleSummaryDTO, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, summaryView(item))
	}
	request.JSON(200, gin.H{"items": items, "next_cursor": nextCursor(page.Next, filter.Sort)})
}

type genreDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	NameRU       string `json:"name_ru"`
	Slug         string `json:"slug"`
	CategorySlug string `json:"category_slug"`
	ParentSlug   string `json:"parent_slug"`
	Position     int    `json:"position"`
}

func (handler *CatalogHandler) rankings(request *gin.Context) {
	limit, valid := queryInteger(request, "limit", 16, 16)
	if !valid {
		return
	}
	ranked, failure := handler.service.Rankings(request.Request.Context(), request.Query("kind"), limit)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	items := []titleSummaryDTO{}
	for _, item := range ranked {
		items = append(items, summaryView(item))
	}
	request.JSON(200, gin.H{"items": items, "next_cursor": nil})
}

type projectDTO struct {
	ID        string `json:"id"`
	TitleID   string `json:"title_id"`
	TeamID    string `json:"team_id"`
	TitleName string `json:"title_name"`
	TitleSlug string `json:"title_slug"`
	TeamName  string `json:"team_name"`
	TeamSlug  string `json:"team_slug"`
	Status    string `json:"status"`
}

func projectViews(projects []application.Project) []projectDTO {
	items := make([]projectDTO, 0, len(projects))
	for _, item := range projects {
		items = append(items, projectDTO{item.ID, item.TitleID, item.TeamID, item.TitleName, item.TitleSlug, item.TeamName, item.TeamSlug, item.Status})
	}
	return items
}

type studioProjectDTO struct {
	NextRevisionID    string `json:"next_revision_id,omitempty"`
	NextRevisionState string `json:"next_revision_state,omitempty"`
	projectDTO
	CoverURL     string  `json:"cover_url"`
	ChapterCount int     `json:"chapter_count"`
	RatingCount  int     `json:"rating_count"`
	Score        float64 `json:"score"`
}

func studioProjectViews(projects []application.Project) []studioProjectDTO {
	base := projectViews(projects)
	items := make([]studioProjectDTO, 0, len(projects))
	for i, item := range projects {
		items = append(items, studioProjectDTO{NextRevisionID: item.NextRevisionID, NextRevisionState: item.NextRevisionState, projectDTO: base[i], CoverURL: coverURL(item.CoverID), ChapterCount: item.ChapterCount, RatingCount: item.RatingCount, Score: item.Score})
	}
	return items
}
func genreViews(genres []application.Genre) []genreDTO {
	items := make([]genreDTO, 0, len(genres))
	for _, item := range genres {
		items = append(items, genreDTO{ID: item.ID, Name: item.Name, NameRU: item.NameRU, Slug: item.Slug, CategorySlug: item.CategorySlug, ParentSlug: item.ParentSlug, Position: item.Position})
	}
	return items
}
func (handler *CatalogHandler) detail(request *gin.Context) {
	detail, failure := handler.service.Detail(request.Request.Context(), request.Param("slug"))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, detailView(detail))
}
func detailView(detail application.Detail) gin.H {
	title := detail.Title
	ageRating := 0
	if title.AgeRating != nil {
		ageRating = *title.AgeRating
	}
	return gin.H{"age_rating": ageRating, "id": title.ID, "slug": title.Slug, "name": title.Name, "original_name": title.OriginalName, "publication": title.Publication, "chapter_count": detail.ChapterCount, "comment_count": detail.CommentCount, "reader_count": detail.ReaderCount, "rating_distribution": detail.RatingDistribution, "aliases": title.Aliases, "authors": title.Authors, "featured": title.Featured, "listed": title.Listed, "kind": title.Kind, "original_status": title.OriginalStatus, "translation_status": title.TranslationStatus, "description": title.Description, "direction": title.Direction, "demo": title.Demo, "cover_url": coverURL(detail.CoverID), "banner_url": coverURL(detail.BannerID), "score": detail.Score, "rating_count": detail.RatingCount, "categories": genreViews(detail.Categories), "collections": collectionViews(detail.Collections), "genres": genreViews(detail.Genres), "projects": projectViews(detail.Projects)}
}
func (handler *CatalogHandler) adminDetail(request *gin.Context) {
	account, _ := CurrentAccount(request)
	detail, failure := handler.service.AdminDetail(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, detailView(detail))
}
func (handler *CatalogHandler) approvedTeams(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.ApprovedTeams(request.Request.Context(), account.ID, requestOffset(request))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, gin.H{"id": item.ID, "name": item.Name, "slug": item.Slug})
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}

type releaseDTO struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	TeamID      string    `json:"team_id"`
	TeamName    string    `json:"team_name"`
	TeamSlug    string    `json:"team_slug"`
	PublishedAt time.Time `json:"published_at"`
}
type chapterDTO struct {
	ID       string       `json:"id"`
	Label    string       `json:"label"`
	Volume   string       `json:"volume"`
	Name     string       `json:"name"`
	Order    int          `json:"order"`
	Releases []releaseDTO `json:"releases"`
}

func (handler *CatalogHandler) chapters(request *gin.Context) {
	limit, valid := queryInteger(request, "limit", 50, 100)
	if !valid {
		return
	}
	after := 0
	if request.Query("cursor") != "" {
		after, valid = queryInteger(request, "cursor", 0, 1000000)
		if !valid {
			return
		}
	}
	page, failure := handler.service.Chapters(request.Request.Context(), request.Param("slug"), after, limit, application.ChapterFilter{Team: request.Query("team"), Query: request.Query("q"), Sort: request.DefaultQuery("sort", "asc")})
	if failure != nil {
		catalogError(request, failure)
		return
	}
	items := make([]chapterDTO, 0, len(page.Items))
	for _, chapter := range page.Items {
		view := chapterDTO{ID: chapter.ID, Label: chapter.Label, Volume: chapter.Volume, Name: chapter.Name, Order: chapter.Order, Releases: []releaseDTO{}}
		for _, release := range chapter.Releases {
			view.Releases = append(view.Releases, releaseDTO{release.ID, release.ProjectID, release.TeamID, release.TeamName, release.TeamSlug, release.PublishedAt})
		}
		items = append(items, view)
	}
	var next *string
	if page.NextOrder != nil {
		value := strconv.Itoa(*page.NextOrder)
		next = &value
	}
	request.JSON(200, gin.H{"items": items, "next_cursor": next})
}
func (handler *CatalogHandler) genres(request *gin.Context) {
	items, failure := handler.service.Genres(request.Request.Context())
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, gin.H{"items": genreViews(items), "next_cursor": nil})
}
func (handler *CatalogHandler) team(request *gin.Context) {
	team, failure := handler.service.Team(request.Request.Context(), request.Param("slug"))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	members := []gin.H{}
	for _, member := range team.Members {
		members = append(members, gin.H{"name": member.Name, "role": member.Role})
	}
	request.JSON(200, gin.H{"members": members, "id": team.ID, "slug": team.Slug, "name": team.Name, "description": team.Description, "telegram_url": team.TelegramURL, "projects": projectViews(team.Projects)})
}
func (handler *CatalogHandler) studioProjects(request *gin.Context) {
	account, _ := CurrentAccount(request)
	offset := requestOffset(request)
	if cursor := request.Query("cursor"); cursor != "" {
		value, failure := strconv.Atoi(cursor)
		if failure != nil {
			Error(request, 400, "validation", "Invalid project cursor")
			return
		}
		offset = value
	}
	page, failure := handler.service.StudioProjectPage(request.Request.Context(), account.ID, application.ProjectFilter{Offset: offset, Query: request.Query("q"), TeamID: request.Query("team")})
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, gin.H{"items": studioProjectViews(page.Items), "next_cursor": page.Next, "total": page.Total})
}
func (handler *CatalogHandler) adminTitles(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.SearchAdminTitles(request.Request.Context(), account.ID, request.Query("q"), requestOffset(request))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	views := make([]titleSummaryDTO, 0, len(items))
	for _, item := range items {
		views = append(views, summaryView(item))
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *CatalogHandler) createTitle(request *gin.Context) {
	var input struct {
		AgeRating         *int               `json:"age_rating"`
		Publication       domain.Publication `json:"publication"`
		Slug              string             `json:"slug"`
		Name              string             `json:"name"`
		OriginalName      string             `json:"original_name"`
		Kind              string             `json:"kind"`
		OriginalStatus    string             `json:"original_status"`
		TranslationStatus string             `json:"translation_status"`
		Description       string             `json:"description"`
		Direction         string             `json:"direction"`
		Aliases           []string           `json:"aliases"`
		Authors           []string           `json:"authors"`
		Featured          bool               `json:"featured"`
		Listed            bool               `json:"listed"`
		GenreIDs          []string           `json:"genre_ids"`
		Demo              bool               `json:"demo"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	metadata := domain.Metadata{AgeRating: input.AgeRating, Publication: input.Publication, Slug: input.Slug, Name: input.Name, OriginalName: input.OriginalName, Kind: input.Kind, OriginalStatus: input.OriginalStatus, TranslationStatus: input.TranslationStatus, Description: input.Description, Direction: input.Direction, Authors: input.Authors, Featured: input.Featured, Listed: input.Listed, Aliases: input.Aliases, GenreIDs: input.GenreIDs, Demo: input.Demo}
	var title domain.Title
	var failure error
	if request.Request.Method == "PUT" {
		title, failure = handler.service.UpdateTitle(request.Request.Context(), account.ID, request.Param("id"), metadata)
	} else {
		title, failure = handler.service.CreateTitle(request.Request.Context(), account.ID, metadata)
	}
	if failure != nil {
		catalogError(request, failure)
		return
	}
	status := 201
	if request.Request.Method == "PUT" {
		status = 200
	}
	request.JSON(status, gin.H{"id": title.ID, "slug": title.Slug})
}
func (handler *CatalogHandler) createGenre(request *gin.Context) {
	var input struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	item, failure := handler.service.CreateGenre(request.Request.Context(), account.ID, input.Name, input.Slug)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(201, genreDTO{ID: item.ID, Name: item.Name, Slug: item.Slug})
}
func (handler *CatalogHandler) createChapter(request *gin.Context) {
	var input struct {
		Label  string `json:"label"`
		Volume string `json:"volume"`
		Name   string `json:"name"`
		Order  int    `json:"order"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	id, failure := handler.service.CreateChapter(request.Request.Context(), account.ID, request.Param("id"), input.Label, input.Volume, input.Name, input.Order)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(201, gin.H{"id": id})
}
func (handler *CatalogHandler) createProject(request *gin.Context) {
	var input struct {
		TeamID string `json:"team_id"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	id, failure := handler.service.CreateProject(request.Request.Context(), account.ID, request.Param("id"), input.TeamID)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(201, gin.H{"id": id})
}
func catalogError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrRequestLimit):
		Error(request, 429, "request_limit", "Өтінімдер шегіне жеттіңіз. Алдыңғы өтінімдер қаралғанша күтіңіз.")
	case errors.Is(failure, domain.ErrInvalid):
		Error(request, 422, "validation", "Тайтл деректерін тексеріңіз")
	case errors.Is(failure, domain.ErrForbidden):
		Error(request, 403, "forbidden", "Бұл әрекетке рұқсат жоқ")
	case errors.Is(failure, domain.ErrNotFound):
		Error(request, 404, "not_found", "Тайтл немесе команда табылмады")
	case errors.Is(failure, domain.ErrConflict):
		Error(request, 409, "conflict", "Атауы немесе тарау реті бос емес")
	default:
		log.Printf("catalogue failure: %T", failure)
		Error(request, 503, "unavailable", "Каталог уақытша қолжетімсіз")
	}
}
func (handler *CatalogHandler) updates(request *gin.Context) {
	limit, valid := queryInteger(request, "limit", 24, 48)
	if !valid {
		return
	}
	result, failure := handler.service.Updates(request.Request.Context(), request.Query("after"), request.Query("team"), limit)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	items := []gin.H{}
	for _, item := range result.Items {
		items = append(items, gin.H{"release_id": item.ReleaseID, "title_id": item.TitleID, "title_name": item.TitleName, "title_slug": item.TitleSlug, "chapter_label": item.ChapterLabel, "team_name": item.TeamName, "team_slug": item.TeamSlug, "demo": item.Demo, "published_at": item.PublishedAt, "cover_url": coverURL(item.CoverID)})
	}
	var next *string
	if result.Next != "" {
		next = &result.Next
	}
	request.JSON(200, gin.H{"items": items, "next_cursor": next})
}
