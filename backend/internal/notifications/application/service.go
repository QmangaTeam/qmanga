package application

import (
	"context"
	"errors"
	"time"
)

type Notification struct {
	ID, Kind, Title, Body, ReleaseID, TitleSlug, ProjectID, CommentID string
	Read                                                              bool
	CreatedAt                                                         time.Time
}
type Repository interface {
	List(context.Context, string, string, ...bool) ([]Notification, error)
	Read(context.Context, string, string) error
	Unread(context.Context, string) (int, error)
	Deliver(context.Context) (bool, error)
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository} }
func (service *Service) List(ctx context.Context, user, after string, unread ...bool) ([]Notification, error) {
	return service.repository.List(ctx, user, after, unread...)
}
func (service *Service) Read(ctx context.Context, user, id string) error {
	return service.repository.Read(ctx, user, id)
}
func (service *Service) Deliver(ctx context.Context) (bool, error) {
	return service.repository.Deliver(ctx)
}

func (service *Service) Unread(ctx context.Context, user string) (int, error) {
	return service.repository.Unread(ctx, user)
}

// FilteredRepository is additive so older consumers and port doubles remain compatible.
type FilteredRepository interface {
	ListFiltered(context.Context, string, string, bool, string) ([]Notification, error)
	ReadAll(context.Context, string) error
}

func (service *Service) ListFiltered(ctx context.Context, user, after string, unread bool, kind string) ([]Notification, error) {
	if repository, ok := service.repository.(FilteredRepository); ok {
		return repository.ListFiltered(ctx, user, after, unread, kind)
	}
	return service.repository.List(ctx, user, after, unread)
}
func (service *Service) ReadAll(ctx context.Context, user string) error {
	if repository, ok := service.repository.(FilteredRepository); ok {
		return repository.ReadAll(ctx, user)
	}
	return errors.New("bulk read unavailable")
}
