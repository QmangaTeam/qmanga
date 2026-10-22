package application

import (
	"context"
	"qmanga/internal/identity/domain"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type APIToken struct {
	ID            string     `json:"id"`
	UserID        string     `json:"-"`
	Name          string     `json:"name"`
	Prefix        string     `json:"prefix"`
	Hash          string     `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	DirectPublish bool       `json:"direct_publish"`
}
type APITokenRepository interface {
	CreateAPIToken(context.Context, APIToken) error
	ListAPITokens(context.Context, string) ([]APIToken, error)
	RevokeAPIToken(context.Context, string, string, time.Time) error
	AuthenticateAPIToken(context.Context, string, time.Time) (APIToken, domain.Account, error)
}

func (s *Service) apiTokens() (APITokenRepository, error) {
	r, ok := s.repository.(APITokenRepository)
	if !ok {
		return nil, domain.ErrForbidden
	}
	return r, nil
}
func (s *Service) CreateAPIToken(ctx context.Context, owner, name string, days int, direct bool) (string, APIToken, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 || strings.IndexFunc(name, unicode.IsControl) >= 0 || days < 1 || days > 365 {
		return "", APIToken{}, domain.ErrInvalid
	}
	r, err := s.apiTokens()
	if err != nil {
		return "", APIToken{}, err
	}
	raw, err := s.credentials.RandomToken()
	if err != nil {
		return "", APIToken{}, err
	}
	raw = "qma_" + raw
	now := s.now()
	item := APIToken{ID: s.credentials.NewID(), UserID: owner, Name: name, Prefix: raw[:12], Hash: s.credentials.Digest(raw), CreatedAt: now, ExpiresAt: now.Add(time.Duration(days) * 24 * time.Hour), DirectPublish: direct}
	if err = r.CreateAPIToken(ctx, item); err != nil {
		return "", APIToken{}, err
	}
	return raw, item, nil
}
func (s *Service) ListAPITokens(ctx context.Context, owner string) ([]APIToken, error) {
	r, e := s.apiTokens()
	if e != nil {
		return nil, e
	}
	return r.ListAPITokens(ctx, owner)
}
func (s *Service) RevokeAPIToken(ctx context.Context, owner, id string) error {
	r, e := s.apiTokens()
	if e != nil {
		return e
	}
	return r.RevokeAPIToken(ctx, owner, id, s.now())
}
func (s *Service) AuthenticateAPIToken(ctx context.Context, raw string) (APIToken, domain.Account, error) {
	if len(raw) != 47 || !strings.HasPrefix(raw, "qma_") {
		return APIToken{}, domain.Account{}, domain.ErrCredentials
	}
	r, e := s.apiTokens()
	if e != nil {
		return APIToken{}, domain.Account{}, e
	}
	return r.AuthenticateAPIToken(ctx, s.credentials.Digest(raw), s.now())
}
