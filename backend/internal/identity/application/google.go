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
	Name, Picture          string
	Verified               bool
}
type GoogleFlow struct {
	LinkUserID, LinkSessionHash                 string
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
type GoogleAvatarImporter interface {
	ImportGoogleAvatar(context.Context, string, string) error
}

type GoogleService struct {
	avatars     GoogleAvatarImporter
	repository  GoogleRepository
	credentials Credentials
	provider    GoogleProvider
	identity    *Service
	now         func() time.Time
}

func NewGoogle(repository GoogleRepository, credentials Credentials, provider GoogleProvider, identity *Service, avatars ...GoogleAvatarImporter) *GoogleService {
	service := &GoogleService{repository: repository, credentials: credentials, provider: provider, identity: identity, now: time.Now}
	if len(avatars) > 0 {
		service.avatars = avatars[0]
	}
	return service
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
	return s.start(ctx, returnPath, nickname, "", "")
}
func (s *GoogleService) start(ctx context.Context, returnPath, nickname, linkUser, linkSession string) (string, string, error) {
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
	flow := GoogleFlow{LinkUserID: linkUser, LinkSessionHash: linkSession, Hash: s.credentials.Digest(state), Nonce: nonce, Verifier: verifier, ReturnPath: SafeAuthReturn(returnPath), Nickname: nickname, ExpiresAt: s.now().Add(10 * time.Minute)}
	if err = s.repository.SaveGoogleFlow(ctx, flow); err != nil {
		return "", "", err
	}
	return state, s.provider.AuthorizationURL(state, nonce, verifier), nil
}
func (s *GoogleService) Finish(ctx context.Context, state, code string) (LoginResult, string, error) {
	return s.finish(ctx, state, code, "")
}
func (s *GoogleService) FinishWithGuestImport(ctx context.Context, state, code, token string) (LoginResult, string, error) {
	if !validRegistrationReceipt(token) {
		return LoginResult{}, "/", domain.ErrInvalid
	}
	return s.finish(ctx, state, code, s.credentials.Digest(token))
}
func (s *GoogleService) finish(ctx context.Context, state, code, receiptHash string) (LoginResult, string, error) {
	if len(state) != 43 || s.provider == nil {
		return LoginResult{}, "/", domain.ErrCredentials
	}
	flow, err := s.repository.ConsumeGoogleFlow(ctx, s.credentials.Digest(state), s.now())
	if err != nil {
		return LoginResult{}, "/", err
	}
	if flow.LinkUserID != "" || flow.LinkSessionHash != "" {
		return LoginResult{}, "/settings", domain.ErrForbidden
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
	candidate := domain.Account{ID: s.credentials.NewID(), Email: email, PasswordHash: hash, Role: "reader", Verified: true, DisplayName: googleDisplayName(identity.Name, email)}
	var account domain.Account
	if receiptHash != "" {
		repository, ok := s.repository.(interface {
			GoogleAccountForImport(context.Context, GoogleIdentity, domain.Account, string, time.Time) (domain.Account, error)
		})
		if !ok {
			return LoginResult{}, flow.ReturnPath, domain.ErrInvalid
		}
		account, err = repository.GoogleAccountForImport(ctx, identity, candidate, receiptHash, s.now().Add(RegistrationImportLifetime))
	} else {
		account, err = s.repository.GoogleAccount(ctx, identity, candidate)
	}
	if err != nil {
		return LoginResult{}, flow.ReturnPath, err
	}
	// Only the freshly inserted candidate receives defaults; stable-sub logins retain edits.
	if account.ID == candidate.ID && s.avatars != nil && identity.Picture != "" {
		// An unavailable picture must not prevent an otherwise verified login.
		_ = s.avatars.ImportGoogleAvatar(ctx, account.ID, identity.Picture)
	}
	result, err := s.identity.newSession(ctx, account)
	return result, flow.ReturnPath, err
}

func googleDisplayName(name, email string) string {
	clean := func(value string) string {
		var result []rune
		for _, r := range value {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || strings.ContainsRune(" _.-", r) {
				result = append(result, r)
			} else if unicode.IsSpace(r) {
				result = append(result, ' ')
			}
		}
		result = []rune(strings.Join(strings.Fields(string(result)), " "))
		if len(result) > 40 {
			result = result[:40]
		}
		normalized, err := domain.NormalizeDisplayName(strings.TrimSpace(string(result)))
		if err != nil {
			return ""
		}
		return normalized
	}
	if value := clean(name); value != "" {
		return value
	}
	if value := clean(strings.SplitN(email, "@", 2)[0]); value != "" {
		return value
	}
	return "Google reader"
}
