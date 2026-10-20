package application

import (
	"context"
	"qmanga/internal/catalog/domain"
	"strings"
	"time"
	"unicode/utf8"
)

type Genre struct {
	ID, Name, NameRU, Slug   string
	CategorySlug, ParentSlug string
	Position                 int
}
type Summary struct {
	AgeRating                                                        int
	ID, Slug, Name, Kind, OriginalStatus, TranslationStatus, CoverID string
	Demo                                                             bool
	ChapterCount, RatingCount                                        int
	Score                                                            float64
	CreatedAt, UpdatedAt                                             time.Time
}
type Cursor struct {
	UpdatedAt time.Time
	Name, ID  string
}
type Filter struct {
	AgeRatings                                           []int
	HasChapters, Featured                                bool
	Query, Category, Collection, Sort                    string
	Kinds, OriginalStatuses, TranslationStatuses, Genres []string
	Formats, Sources, Countries                          []string
	ChaptersMin, ChaptersMax, YearMin, YearMax           int
	Limit                                                int
	After                                                *Cursor
}
type Page struct {
	Items []Summary
	Next  *Cursor
}
type Detail struct {
	CommentCount       int
	ChapterCount       int
	ReaderCount        int
	RatingDistribution [10]int
	Title              domain.Title
	CoverID            string
	BannerID           string
	Score              float64
	RatingCount        int
	Genres             []Genre
	Categories         []Genre
	Collections        []domain.Collection
	Projects           []Project
}
type Project struct {
	NextRevisionID, NextRevisionState                                              string
	ID, TitleID, TeamID, TitleName, TitleSlug, TeamName, TeamSlug, Status, CoverID string
	ChapterCount, RatingCount                                                      int
	Score                                                                          float64
}
type ProjectFilter struct {
	Query, TeamID string
	Offset        int
}
type ProjectPage struct {
	Items []Project
	Total int
	Next  string
}
type Release struct {
	ID, ProjectID, TeamID, TeamName, TeamSlug string
	PublishedAt                               time.Time
}
type Chapter struct {
	ID, Label, Volume, Name string
	Order                   int
	Releases                []Release
}
type ChapterPage struct {
	Items     []Chapter
	NextOrder *int
}
type ChapterFilter struct {
	Query, Sort  string
	After, Limit int
}
type PublicMember struct{ Name, Role string }
type Team struct {
	Members                     []PublicMember
	ID, Slug, Name, Description string
	TelegramURL                 string
	Projects                    []Project
}
type Update struct {
	ReleaseID, TitleID, TitleName, TitleSlug, ChapterLabel, TeamName, TeamSlug, CoverID string
	Demo                                                                                bool
	PublishedAt                                                                         time.Time
}
type UpdatePage struct {
	Items []Update
	Next  string
}
type Repository interface {
	Rankings(context.Context, string, int) ([]Summary, error)
	CreateProjectRequest(context.Context, string, domain.ProjectRequest) error
	ProjectRequests(context.Context, string, bool, string, ...ProjectRequestFilter) ([]domain.ProjectRequest, error)
	ReviewProjectRequest(context.Context, string, string, string, string, string) (domain.ProjectRequest, error)
	CancelProjectRequest(context.Context, string, string) error
	StudioProject(context.Context, string, string) (StudioProject, error)
	AdminProject(context.Context, string, string) (Project, error)
	CreateStudioChapter(context.Context, string, string, string, string, string, int) (string, error)
	Updates(context.Context, string, string, int) (UpdatePage, error)
	List(context.Context, Filter) (Page, error)
	Detail(context.Context, string) (Detail, error)
	Chapters(context.Context, string, ChapterFilter) (ChapterPage, error)
	Genres(context.Context) ([]Genre, error)
	Team(context.Context, string) (Team, error)
	IsAdmin(context.Context, string) (bool, error)
	CreateTitle(context.Context, string, domain.Title) error
	CreateGenre(context.Context, string, Genre) error
	CreateChapter(context.Context, string, string, string, string, string, int) (string, error)
	CreateProject(context.Context, string, string, string, string) error
	StudioProjects(context.Context, string, ...int) ([]Project, error)
	StudioProjectPage(context.Context, string, ProjectFilter) (ProjectPage, error)
	SearchAdminTitles(context.Context, string, string, ...int) ([]Summary, error)
	AdminTitles(context.Context, string, ...int) ([]Summary, error)
	AdminDetail(context.Context, string, string) (Detail, error)
	UpdateTitle(context.Context, string, domain.Title) error
	ApprovedTeams(context.Context, string, ...int) ([]Team, error)
}
type Service struct {
	repository Repository
	newID      func() string
}

func New(repository Repository, newID func() string) *Service {
	return &Service{repository: repository, newID: newID}
}
func (service *Service) List(ctx context.Context, filter Filter) (Page, error) {
	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	if utf8.RuneCountInString(filter.Query) > 80 {
		return Page{}, domain.ErrInvalid
	}
	if filter.Sort == "" {
		filter.Sort = "updated"
	}
	if filter.Sort != "updated" && filter.Sort != "name" && filter.Sort != "new" {
		return Page{}, domain.ErrInvalid
	}
	if filter.Limit == 0 {
		filter.Limit = 24
	}
	filter.Limit = min(max(filter.Limit, 1), 48)
	var valid bool
	if filter.Kinds, valid = selection(filter.Kinds, 3); !valid {
		return Page{}, domain.ErrInvalid
	}
	for _, kind := range filter.Kinds {
		if !domain.ValidKind(kind) {
			return Page{}, domain.ErrInvalid
		}
	}
	if filter.OriginalStatuses, valid = selection(filter.OriginalStatuses, 4); !valid {
		return Page{}, domain.ErrInvalid
	}
	for _, status := range filter.OriginalStatuses {
		if !domain.ValidOriginalStatus(status) {
			return Page{}, domain.ErrInvalid
		}
	}
	if filter.TranslationStatuses, valid = selection(filter.TranslationStatuses, 4); !valid {
		return Page{}, domain.ErrInvalid
	}
	for _, status := range filter.TranslationStatuses {
		if !domain.ValidTranslationStatus(status) {
			return Page{}, domain.ErrInvalid
		}
	}
	if filter.Formats, valid = selection(filter.Formats, 8); !valid {
		return Page{}, domain.ErrInvalid
	}
	if filter.Sources, valid = selection(filter.Sources, 16); !valid {
		return Page{}, domain.ErrInvalid
	}
	if filter.Countries, valid = selection(filter.Countries, 16); !valid {
		return Page{}, domain.ErrInvalid
	}
	if len(filter.Category) > 64 || len(filter.Collection) > 64 || (filter.After != nil && (filter.After.ID == "" || len(filter.After.ID) > 36 || utf8.RuneCountInString(filter.After.Name) > 200 || filter.After.UpdatedAt.IsZero())) {
		return Page{}, domain.ErrInvalid
	}
	if len(filter.AgeRatings) > 0 {
		seenRatings := map[int]bool{}
		ratings := make([]int, 0, len(filter.AgeRatings))
		for _, rating := range filter.AgeRatings {
			if !domain.ValidAgeRating(rating) {
				return Page{}, domain.ErrInvalid
			}
			if seenRatings[rating] {
				continue
			}
			seenRatings[rating] = true
			ratings = append(ratings, rating)
		}
		filter.AgeRatings = ratings
	}
	if filter.ChaptersMin < 0 || filter.ChaptersMax < 0 || filter.ChaptersMax > 1_000_000 || (filter.ChaptersMin > 0 && filter.ChaptersMax > 0 && filter.ChaptersMin > filter.ChaptersMax) {
		return Page{}, domain.ErrInvalid
	}
	if filter.YearMin < 0 || filter.YearMax < 0 || filter.YearMin > 2200 || filter.YearMax > 2200 || (filter.YearMin > 0 && filter.YearMin < 1000) || (filter.YearMax > 0 && filter.YearMax < 1000) || (filter.YearMin > 0 && filter.YearMax > 0 && filter.YearMin > filter.YearMax) {
		return Page{}, domain.ErrInvalid
	}
	if filter.Genres, valid = selection(filter.Genres, 12); !valid {
		return Page{}, domain.ErrInvalid
	}
	for _, slug := range filter.Genres {
		if !domain.ValidSlug(slug) {
			return Page{}, domain.ErrInvalid
		}
	}
	return service.repository.List(ctx, filter)
}

// selection trims values, drops empty entries, removes duplicates and rejects
// a group that still holds more than limit distinct values.
func selection(values []string, limit int) ([]string, bool) {
	seen := make(map[string]bool, len(values))
	selected := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if utf8.RuneCountInString(trimmed) > 160 {
			return nil, false
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		selected = append(selected, trimmed)
	}
	if len(selected) > limit {
		return nil, false
	}
	return selected, true
}

func (service *Service) Rankings(ctx context.Context, kind string, limit int) ([]Summary, error) {
	if kind != "" && kind != "manga" && kind != "manhwa" && kind != "manhua" {
		return nil, domain.ErrInvalid
	}
	if limit == 0 {
		limit = 16
	}
	return service.repository.Rankings(ctx, kind, min(max(limit, 1), 16))
}
func (service *Service) Detail(ctx context.Context, slug string) (Detail, error) {
	return service.repository.Detail(ctx, slug)
}
func (service *Service) Chapters(ctx context.Context, slug string, after, limit int, options ...ChapterFilter) (ChapterPage, error) {
	filter := ChapterFilter{}
	if len(options) > 0 {
		filter = options[0]
	}
	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	if filter.Sort == "" {
		filter.Sort = "asc"
	}
	if utf8.RuneCountInString(filter.Query) > 80 || (filter.Sort != "asc" && filter.Sort != "desc") {
		return ChapterPage{}, domain.ErrInvalid
	}
	if limit == 0 {
		limit = 50
	}
	filter.After = max(after, 0)
	filter.Limit = min(max(limit, 1), 100)
	return service.repository.Chapters(ctx, slug, filter)
}
func (service *Service) Genres(ctx context.Context) ([]Genre, error) {
	return service.repository.Genres(ctx)
}
func (service *Service) Team(ctx context.Context, slug string) (Team, error) {
	return service.repository.Team(ctx, slug)
}
func (service *Service) admin(ctx context.Context, actorID string) error {
	allowed, failure := service.repository.IsAdmin(ctx, actorID)
	if failure != nil {
		return failure
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}
func (service *Service) CreateTitle(ctx context.Context, actorID string, input domain.Metadata) (domain.Title, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return domain.Title{}, failure
	}
	title, failure := domain.NewTitle(service.newID(), input)
	if failure != nil {
		return title, failure
	}
	return title, service.repository.CreateTitle(ctx, actorID, title)
}
func (service *Service) CreateGenre(ctx context.Context, actorID, name, slug string) (Genre, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return Genre{}, failure
	}
	if !domain.ValidGenre(name, slug) {
		return Genre{}, domain.ErrInvalid
	}
	genre := Genre{ID: service.newID(), Name: strings.TrimSpace(name), Slug: slug}
	return genre, service.repository.CreateGenre(ctx, actorID, genre)
}
func (service *Service) CreateChapter(ctx context.Context, actorID, titleID, label, volume, name string, order int) (string, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return "", failure
	}
	if failure := domain.ValidateChapter(label, order); failure != nil {
		return "", failure
	}
	if utf8.RuneCountInString(volume) > 32 || utf8.RuneCountInString(name) > 200 {
		return "", domain.ErrInvalid
	}
	return service.repository.CreateChapter(ctx, actorID, titleID, strings.TrimSpace(label), volume, name, order)
}
func (service *Service) CreateProject(ctx context.Context, actorID, titleID, teamID string) (string, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return "", failure
	}
	id := service.newID()
	return id, service.repository.CreateProject(ctx, actorID, id, titleID, teamID)
}
func (service *Service) StudioProjects(ctx context.Context, actorID string, offset ...int) ([]Project, error) {
	return service.repository.StudioProjects(ctx, actorID, offset...)
}
func (service *Service) StudioProjectPage(ctx context.Context, actor string, filter ProjectFilter) (ProjectPage, error) {
	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	if utf8.RuneCountInString(filter.Query) > 80 || len(filter.TeamID) > 36 || filter.Offset < 0 || filter.Offset > 100000 {
		return ProjectPage{}, domain.ErrInvalid
	}
	return service.repository.StudioProjectPage(ctx, actor, filter)
}
func (service *Service) SearchAdminTitles(ctx context.Context, actorID, query string, offset ...int) ([]Summary, error) {
	if utf8.RuneCountInString(query) > 160 {
		return nil, domain.ErrInvalid
	}
	if failure := service.admin(ctx, actorID); failure != nil {
		return nil, failure
	}
	return service.repository.SearchAdminTitles(ctx, actorID, strings.TrimSpace(query), offset...)
}

func (service *Service) AdminTitles(ctx context.Context, actorID string, offset ...int) ([]Summary, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return nil, failure
	}
	return service.repository.AdminTitles(ctx, actorID, offset...)
}

func (service *Service) AdminDetail(ctx context.Context, actorID, id string) (Detail, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return Detail{}, failure
	}
	return service.repository.AdminDetail(ctx, actorID, id)
}
func (service *Service) UpdateTitle(ctx context.Context, actorID, id string, input domain.Metadata) (domain.Title, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return domain.Title{}, failure
	}
	title, failure := domain.NewTitle(id, input)
	if failure != nil {
		return title, failure
	}
	return title, service.repository.UpdateTitle(ctx, actorID, title)
}
func (service *Service) ApprovedTeams(ctx context.Context, actorID string, offset ...int) ([]Team, error) {
	if failure := service.admin(ctx, actorID); failure != nil {
		return nil, failure
	}
	return service.repository.ApprovedTeams(ctx, actorID, offset...)
}
func (service *Service) Updates(ctx context.Context, after, team string, limit int) (UpdatePage, error) {
	if len(after) > 36 || len(team) > 128 {
		return UpdatePage{}, domain.ErrInvalid
	}
	if limit == 0 {
		limit = 24
	}
	return service.repository.Updates(ctx, after, team, min(max(limit, 1), 48))
}
