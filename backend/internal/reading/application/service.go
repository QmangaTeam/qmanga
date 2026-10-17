package application

import (
	"context"
	"qmanga/internal/reading/domain"
	"time"
)

type Entry struct {
	Version                                      int64
	Removed                                      bool
	TitleID, Name, Slug, Status, PreferredTeamID string
	Favorite, Notify                             bool
	Unseen                                       int
	UpdatedAt                                    time.Time
	Progress                                     domain.Progress
}
type Selection struct {
	Version                          int64
	Removed                          bool
	TitleID, Status, PreferredTeamID string
	Favorite, Notify                 bool
}
type Result struct {
	Progress domain.Progress
	Outcome  string
}
type Repository interface {
	Selection(context.Context, string, string) (*Entry, error)
	List(context.Context, string, string) ([]Entry, error)
	Save(context.Context, string, Selection) error
	Remove(context.Context, string, string) error
	Progress(context.Context, string, string) (domain.Progress, error)
	Advance(context.Context, string, domain.Event) (Result, error)
}
type Service struct{ repository Repository }

func (service *Service) Selection(ctx context.Context, actor, title string) (*Entry, error) {
	return service.repository.Selection(ctx, actor, title)
}

func New(repository Repository) *Service { return &Service{repository} }
func (service *Service) List(ctx context.Context, actor, after string) ([]Entry, error) {
	return service.repository.List(ctx, actor, after)
}
func (service *Service) Save(ctx context.Context, actor string, input Selection) error {
	if input.Version < 0 || !domain.ValidStatus(input.Status) || input.TitleID == "" {
		return domain.ErrInvalid
	}
	return service.repository.Save(ctx, actor, input)
}
func (service *Service) Remove(ctx context.Context, actor, title string) error {
	return service.repository.Remove(ctx, actor, title)
}
func (service *Service) Progress(ctx context.Context, actor, title string) (domain.Progress, error) {
	return service.repository.Progress(ctx, actor, title)
}
func (service *Service) Advance(ctx context.Context, actor string, event domain.Event) (Result, error) {
	if _, failure := domain.Advance(domain.Progress{Version: event.BaseVersion}, event, 0, false); failure != nil {
		return Result{}, failure
	}
	return service.repository.Advance(ctx, actor, event)
}
