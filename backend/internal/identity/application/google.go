package application

import (
	"context"
	"errors"
	"net/url"
	"qmanga/internal/identity/domain"
	"strings"
	"time"
	"unicode"
)

var ErrNicknameRequired = errors.New("nickname required")

type GoogleIdentity struct {
	Issuer, Subject, Email string
	Verified               bool
}
type GoogleFlow struct {
	Hash, Nonce, Verifier, ReturnPath, Nickname string
	ExpiresAt                                   time.Time
}
type GoogleRepository interface {
	SaveGoogleFlow(context.Context, GoogleFlow) error
	ConsumeGoogleFlow(context.Context, string, time.Time) (GoogleFlow, error)
	GoogleAccount(context.Context, GoogleIdentity, domain.Account) (domain.Account, error)
}
type GoogleProvider interface {
	AuthorizationURL(string, string, string) string
	Exchange(context.Context, string, string, string) (GoogleIdentity, error)
}
type GoogleService struct {
	repository  GoogleRepository
	credentials Credentials
	provider    GoogleProvider
	identity    *Service
	now         func() time.Time
}

func NewGoogle(repository GoogleRepository, credentials Credentials, provider GoogleProvider, identity *Service) *GoogleService {
	return &GoogleService{repository, credentials, provider, identity, time.Now}
}
func SafeAuthReturn(value string) string {
	if len(value) > 2048 {
		return "/"
	}
	// Browser navigation can discard controls and decode separators; check each form.
	candidate := value
	for attempt := 0; attempt < 3; attempt++ {
		if !strings.HasPrefix(candidate, "/") || strings.HasPrefix(candidate, "//") || strings.Contains(candidate, "\\") {
			return "/"
		}
		for _, character := range candidate {
			if unicode.IsControl(character) || unicode.IsSpace(character) {
				return "/"
			}
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.User != nil {
			return "/"
		}
		decoded, err := url.PathUnescape(candidate)
		if err != nil {
			return "/"
		}
		if decoded == candidate {
			break
		}
		candidate = decoded
	}
	return value
}
func (s *GoogleService) Start(ctx context.Context, returnPath, nickname string) (string, string, error) {
	if s.provider == nil {
		return "", "", domain.ErrNotFound
	}
	if nickname != "" {
		if _, err := domain.NormalizeDisplayName(nickname); err != nil {
			return "", "", err
		}
	}
	state, err := s.credentials.RandomToken()
	if err != nil {
		return "", "", err
	}
	nonce, err := s.credentials.RandomToken()
	if err != nil {
		return "", "", err
	}
	verifier, err := s.credentials.RandomToken()
	if err != nil {
		return "", "", err
	}
	flow := GoogleFlow{Hash: s.credentials.Digest(state), Nonce: nonce, Verifier: verifier, ReturnPath: SafeAuthReturn(returnPath), Nickname: nickname, ExpiresAt: s.now().Add(10 * time.Minute)}
	if err = s.repository.SaveGoogleFlow(ctx, flow); err != nil {
		return "", "", err
	}
	return state, s.provider.AuthorizationURL(state, nonce, verifier), nil
}
func (s *GoogleService) Finish(ctx context.Context, state, code string) (LoginResult, string, error) {
	if len(state) != 43 || s.provider == nil {
		return LoginResult{}, "/", domain.ErrCredentials
	}
	flow, err := s.repository.ConsumeGoogleFlow(ctx, s.credentials.Digest(state), s.now())
	if err != nil {
		return LoginResult{}, "/", err
	}
	if code == "" {
		return LoginResult{}, flow.ReturnPath, domain.ErrCredentials
	}
	identity, err := s.provider.Exchange(ctx, code, flow.Nonce, flow.Verifier)
	if err != nil {
		return LoginResult{}, flow.ReturnPath, domain.ErrCredentials
	}
	email, err := domain.NormalizeEmail(identity.Email)
	if err != nil || !identity.Verified || identity.Subject == "" || identity.Issuer != "https://accounts.google.com" {
		return LoginResult{}, flow.ReturnPath, domain.ErrCredentials
	}
	identity.Email = email
	secret, err := s.credentials.RandomToken()
	if err != nil {
		return LoginResult{}, flow.ReturnPath, err
	}
	hash, err := s.credentials.Hash(secret)
	if err != nil {
		return LoginResult{}, flow.ReturnPath, err
	}
	candidate := domain.Account{ID: s.credentials.NewID(), Email: email, PasswordHash: hash, Role: "reader", Verified: true, DisplayName: flow.Nickname}
	account, err := s.repository.GoogleAccount(ctx, identity, candidate)
	if err != nil {
		return LoginResult{}, flow.ReturnPath, err
	}
	result, err := s.identity.newSession(ctx, account)
	return result, flow.ReturnPath, err
}
