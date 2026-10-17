package application

import (
	"context"
	"qmanga/internal/reading/domain"
	"time"
)

type Reference struct{ TitleID, Name, Slug, CoverID, ReleaseID, RevisionID, ChapterLabel string }
type RecentEntry struct {
	Reference
	Progress     domain.Progress
	UpdatedAt    time.Time
	PagePosition int
	Availability string
}

func (service *Service) Recent(ctx context.Context, actor string, limit int) ([]RecentEntry, error) {
	if limit == 0 {
		limit = 4
	}
	if limit < 1 || limit > 12 {
		return nil, domain.ErrInvalid
	}
	return service.repository.Recent(ctx, actor, limit)
}
func (service *Service) References(ctx context.Context, ids []string) ([]Reference, error) {
	if len(ids) > 12 {
		return nil, domain.ErrInvalid
	}
	for _, id := range ids {
		if id == "" || len(id) > 36 {
			return nil, domain.ErrInvalid
		}
	}
	if len(ids) == 0 {
		return []Reference{}, nil
	}
	return service.repository.References(ctx, ids)
}
