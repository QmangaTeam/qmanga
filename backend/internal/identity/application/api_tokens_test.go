package application

import (
	"context"
	"errors"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

type apiTokenRepo struct {
	Repository
	item APIToken
}

func (r *apiTokenRepo) CreateAPIToken(_ context.Context, t APIToken) error { r.item = t; return nil }
func (r *apiTokenRepo) ListAPITokens(context.Context, string) ([]APIToken, error) {
	return []APIToken{r.item}, nil
}
func (r *apiTokenRepo) RevokeAPIToken(context.Context, string, string, time.Time) error { return nil }
func (r *apiTokenRepo) AuthenticateAPIToken(context.Context, string, time.Time) (APIToken, domain.Account, error) {
	return r.item, domain.Account{}, nil
}
func TestAPITokenValidationAndLifetime(t *testing.T) {
	repo := &apiTokenRepo{}
	credentials := &googleCredentials{}
	service := New(repo, credentials, nil, "")
	now := time.Unix(1000, 0)
	service.now = func() time.Time { return now }
	for _, in := range []struct {
		name string
		days int
	}{{"", 1}, {"x", 0}, {"x", 366}, {strings.Repeat("қ", 101), 1}, {"x\ny", 1}} {
		_, _, e := service.CreateAPIToken(context.Background(), "owner", in.name, in.days, true)
		if !errors.Is(e, domain.ErrInvalid) {
			t.Fatalf("invalid token accepted %+v: %v", in, e)
		}
	}
	raw, item, e := service.CreateAPIToken(context.Background(), "owner", "  Codex  ", 365, true)
	if e != nil || item.Name != "Codex" || item.UserID != "owner" || !item.DirectPublish || len(raw) != 47 || !item.ExpiresAt.Equal(now.Add(365*24*time.Hour)) {
		t.Fatalf("created %+v %v", item, e)
	}
}
