package application

import (
	"context"
	"qmanga/internal/catalog/domain"
	"strings"
	"time"
	"unicode/utf8"
)

type Genre struct{ ID, Name, Slug string }
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
	AgeRating                                                                         *int
	HasChapters                                                                       bool
	Featured                                                                          bool
	Query, Kind, Genre, Category, Collection, OriginalStatus, TranslationStatus, Sort string
	Limit                                                                             int
	After                                                                             *Cursor
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
	ProjectRequests(context.Context, string, bool, string) ([]domain.ProjectRequest, error)
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
	if filter.AgeRating != nil && !domain.ValidAgeRating(*filter.AgeRating) {
		return Page{}, domain.ErrInvalid
	}
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
	if filter.Kind != "" && filter.Kind != "manga" && filter.Kind != "manhwa" && filter.Kind != "manhua" {
		return Page{}, domain.ErrInvalid
	}
	if filter.OriginalStatus != "" && filter.OriginalStatus != "ongoing" && filter.OriginalStatus != "completed" && filter.OriginalStatus != "hiatus" && filter.OriginalStatus != "cancelled" {
		return Page{}, domain.ErrInvalid
	}
	if filter.TranslationStatus != "" && filter.TranslationStatus != "ongoing" && filter.TranslationStatus != "completed" && filter.TranslationStatus != "paused" && filter.TranslationStatus != "stopped" {
		return Page{}, domain.ErrInvalid
	}
	if len(filter.Category) > 64 || len(filter.Collection) > 64 || len(filter.Genre) > 64 || (filter.After != nil && (filter.After.ID == "" || len(filter.After.ID) > 36 || utf8.RuneCountInString(filter.After.Name) > 200 || filter.After.UpdatedAt.IsZero())) {
		return Page{}, domain.ErrInvalid
	}
	return service.repository.List(ctx, filter)
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
