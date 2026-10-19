package application

import (
	"context"
	"qmanga/internal/reading/domain"
	"strings"
)

// Statuses reads only the visible titles for one account; it never scans the library.
func (service *Service) Statuses(ctx context.Context, actor string, ids []string) (map[string]string, error) {
	if len(ids) > 60 {
		return nil, domain.ErrInvalid
	}
	unique := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || len(id) > 128 || strings.TrimSpace(id) != id {
			return nil, domain.ErrInvalid
		}
		if !seen[id] {
			unique = append(unique, id)
			seen[id] = true
		}
	}
	if len(unique) == 0 {
		return map[string]string{}, nil
	}
	repository, ok := service.repository.(interface {
		Statuses(context.Context, string, []string) (map[string]string, error)
	})
	if !ok {
		return map[string]string{}, nil
	}
	return repository.Statuses(ctx, actor, unique)
}
