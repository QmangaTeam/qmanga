package application

import (
	"context"
	"crypto/subtle"
	"errors"
	"qmanga/internal/identity/domain"
	"time"
)

type Session struct {
	Hash, UserID, CSRFHash string
	ExpiresAt              time.Time
}
type Token struct {
	Hash, UserID, Kind string
	ExpiresAt          time.Time
}
type Repository interface {
	Accounts(context.Context, string, string, string) ([]ManagedAccount, error)
	ManageAccount(context.Context, string, string, AccountChange) error
	Audit(context.Context, string, string) ([]AuditEntry, error)
	UpdateProfile(context.Context, string, string) error
	CreateAccount(context.Context, domain.Account) error
	AccountByEmail(context.Context, string) (domain.Account, error)
	SaveSession(context.Context, Session) error
	SessionAccount(context.Context, string, time.Time) (Session, domain.Account, error)
	DeleteSession(context.Context, string) error
	SaveToken(context.Context, Token) error
	ConsumeToken(context.Context, string, string, time.Time, string) error
}
type Credentials interface {
	Hash(string) (string, error)
	Verify(string, string) bool
	RandomToken() (string, error)
	NewID() string
	Digest(string) string
}
type Mailer interface {
	Send(context.Context, string, string, string) error
}
type Service struct {
	repository  Repository
	credentials Credentials
	mailer      Mailer
	origin      string
	now         func() time.Time
}
type LoginResult struct {
	Account                 domain.Account
	SessionToken, CSRFToken string
	ExpiresAt               time.Time
}

func New(repository Repository, credentials Credentials, mailer Mailer, origin string) *Service {
	return &Service{repository: repository, credentials: credentials, mailer: mailer, origin: origin, now: time.Now}
}
func (service *Service) Register(ctx context.Context, email, password, displayName string) error {
	name, failure := domain.NormalizeDisplayName(displayName)
	if failure != nil || name == "" {
		return domain.ErrInvalid
	}
	if validationError := domain.ValidatePassword(password); validationError != nil {
		return validationError
	}
	normalized, validationError := domain.NormalizeEmail(email)
	if validationError != nil {
		return validationError
	}
	hashed, hashError := service.credentials.Hash(password)
	if hashError != nil {
		return hashError
	}
	account, createError := domain.NewAccount(service.credentials.NewID(), normalized, hashed)
	if createError != nil {
		return createError
	}
	account.DisplayName = name
	if saveError := service.repository.CreateAccount(ctx, account); saveError != nil {
		if errors.Is(saveError, domain.ErrConflict) {
			return nil
		}
		return saveError
	}
	return service.sendToken(ctx, account, "verify")
}
func (service *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	normalized, validationError := domain.NormalizeEmail(email)
	if validationError != nil || len(password) > 512 {
		return LoginResult{}, domain.ErrCredentials
	}
	account, findError := service.repository.AccountByEmail(ctx, normalized)
	if findError != nil {
		service.credentials.Verify(password, "")
		if errors.Is(findError, domain.ErrNotFound) {
			return LoginResult{}, domain.ErrCredentials
		}
		return LoginResult{}, findError
	}
	if !service.credentials.Verify(password, account.PasswordHash) {
		return LoginResult{}, domain.ErrCredentials
	}
	token, tokenError := service.credentials.RandomToken()
	if tokenError != nil {
		return LoginResult{}, tokenError
	}
	csrf, csrfError := service.credentials.RandomToken()
	if csrfError != nil {
		return LoginResult{}, csrfError
	}
	expiration := service.now().Add(30 * 24 * time.Hour)
	session := Session{Hash: service.credentials.Digest(token), UserID: account.ID, CSRFHash: service.credentials.Digest(csrf), ExpiresAt: expiration}
	if saveError := service.repository.SaveSession(ctx, session); saveError != nil {
		return LoginResult{}, saveError
	}
	return LoginResult{Account: account, SessionToken: token, CSRFToken: csrf, ExpiresAt: expiration}, nil
}
func (service *Service) Authenticate(ctx context.Context, token string) (Session, domain.Account, error) {
	if len(token) != 43 {
		return Session{}, domain.Account{}, domain.ErrCredentials
	}
	return service.repository.SessionAccount(ctx, service.credentials.Digest(token), service.now())
}
func (service *Service) ValidCSRF(session Session, token string) bool {
	return len(token) == 43 && subtle.ConstantTimeCompare([]byte(service.credentials.Digest(token)), []byte(session.CSRFHash)) == 1
}
func (service *Service) Logout(ctx context.Context, token string) error {
	return service.repository.DeleteSession(ctx, service.credentials.Digest(token))
}
func (service *Service) RequestToken(ctx context.Context, email, kind string) error {
	if kind != "verify" && kind != "reset" {
		return domain.ErrInvalid
	}
	normalized, validationError := domain.NormalizeEmail(email)
	if validationError != nil {
		return nil
	}
	account, findError := service.repository.AccountByEmail(ctx, normalized)
	if errors.Is(findError, domain.ErrNotFound) {
		return nil
	}
	if findError != nil {
		return findError
	}
	if kind == "verify" && account.Verified {
		return nil
	}
	return service.sendToken(ctx, account, kind)
}
func (service *Service) sendToken(ctx context.Context, account domain.Account, kind string) error {
	if service.mailer == nil {
		return errors.New("mail delivery unavailable")
	}
	rawToken, tokenError := service.credentials.RandomToken()
	if tokenError != nil {
		return tokenError
	}
	expiration := service.now().Add(time.Hour)
	page, subject := "/reset-password", "Qmanga — құпиясөзді жаңарту"
	if kind == "verify" {
		expiration = service.now().Add(24 * time.Hour)
		page = "/verify-email"
		subject = "Qmanga — поштаны растау"
	}
	if saveError := service.repository.SaveToken(ctx, Token{Hash: service.credentials.Digest(rawToken), UserID: account.ID, Kind: kind, ExpiresAt: expiration}); saveError != nil {
		return saveError
	}
	return service.mailer.Send(ctx, account.Email, subject, "Сілтемені ашыңыз:\n"+service.origin+page+"?token="+rawToken+"\n\nБұл әрекетті сіз бастамасаңыз, хатты елемеңіз.")
}
func (service *Service) VerifyEmail(ctx context.Context, token string) error {
	if len(token) != 43 {
		return domain.ErrExpired
	}
	return service.repository.ConsumeToken(ctx, service.credentials.Digest(token), "verify", service.now(), "")
}
func (service *Service) ResetPassword(ctx context.Context, token, password string) error {
	if len(token) != 43 {
		return domain.ErrExpired
	}
	if validationError := domain.ValidatePassword(password); validationError != nil {
		return validationError
	}
	hashed, hashError := service.credentials.Hash(password)
	if hashError != nil {
		return hashError
	}
	return service.repository.ConsumeToken(ctx, service.credentials.Digest(token), "reset", service.now(), hashed)
}
func (service *Service) CreateBootstrapAdmin(ctx context.Context, email, password string) (string, error) {
	if validationError := domain.ValidatePassword(password); validationError != nil {
		return "", validationError
	}
	hashed, hashError := service.credentials.Hash(password)
	if hashError != nil {
		return "", hashError
	}
	account, createError := domain.NewAccount(service.credentials.NewID(), email, hashed)
	if createError != nil {
		return "", createError
	}
	account.Role = "admin"
	account.Verified = true
	return account.ID, service.repository.CreateAccount(ctx, account)
}
func (service *Service) UpdateProfile(ctx context.Context, actor, name string) error {
	normalized, failure := domain.NormalizeDisplayName(name)
	if failure != nil {
		return failure
	}
	return service.repository.UpdateProfile(ctx, actor, normalized)
}
