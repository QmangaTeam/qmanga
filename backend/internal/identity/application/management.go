package application

import (
	"context"
	"qmanga/internal/identity/domain"
	"strings"
	"time"
	"unicode/utf8"
)

type ManagedAccount struct {
	ID, Name, Role string
	Blocked        bool
	Version        int64
}
type AccountChange struct {
	Role, Reason string
	Blocked      bool
	Version      int64
}
type AuditEntry struct {
	ActorName, TargetName, TargetPath               string
	Before, After                                   map[string]any
	ID, Actor, Action, EntityType, EntityID, Reason string
	CreatedAt                                       time.Time
}

func (service *Service) Accounts(ctx context.Context, actor, query, after string) ([]ManagedAccount, error) {
	if len(query) > 160 || len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	return service.repository.Accounts(ctx, actor, strings.TrimSpace(query), after)
}
func (service *Service) ManageAccount(ctx context.Context, actor, id string, input AccountChange) error {
	if len(id) > 36 || input.Version < 0 || utf8.RuneCountInString(strings.TrimSpace(input.Reason)) < 8 || utf8.RuneCountInString(input.Reason) > 2000 {
		return domain.ErrInvalid
	}
	if input.Role != "reader" && input.Role != "moderator" && input.Role != "admin" {
		return domain.ErrInvalid
	}
	return service.repository.ManageAccount(ctx, actor, id, input)
}

type AuditFilter struct{ Actor, Object, Action, From, To string }

func (service *Service) Audit(ctx context.Context, actor, after string, filters ...AuditFilter) ([]AuditEntry, error) {
	if len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	var filter AuditFilter
	if len(filters) > 0 {
		filter = filters[0]
	}
	if len(filter.Actor) > 160 || len(filter.Object) > 160 || len(filter.Action) > 80 {
		return nil, domain.ErrInvalid
	}
	for _, value := range []string{filter.From, filter.To} {
		if value != "" {
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				return nil, domain.ErrInvalid
			}
		}
	}
	return service.repository.Audit(ctx, actor, after, filter)
}
