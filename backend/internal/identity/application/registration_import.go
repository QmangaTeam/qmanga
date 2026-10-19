package application

import (
	"context"
	"encoding/base64"
	"qmanga/internal/identity/domain"
	"time"
)

const RegistrationImportLifetime = 7 * 24 * time.Hour

func validRegistrationReceipt(token string) bool {
	if len(token) != 43 {
		return false
	}
	value, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(value) == 32
}

// The same opaque browser receipt survives registration retries; it is not an authentication token.
func (service *Service) RegistrationImportToken(existing string) (string, error) {
	if validRegistrationReceipt(existing) {
		return existing, nil
	}
	return service.credentials.RandomToken()
}
func (service *Service) RegisterWithGuestImport(ctx context.Context, email, password, name, token string) error {
	if !validRegistrationReceipt(token) {
		return domain.ErrInvalid
	}
	return service.register(ctx, email, password, name, service.credentials.Digest(token))
}
func (service *Service) RegistrationImportAllowed(ctx context.Context, actor, token string) (bool, error) {
	if !validRegistrationReceipt(token) {
		return false, nil
	}
	repository, ok := service.repository.(interface {
		RegistrationImportAllowed(context.Context, string, string, time.Time) (bool, error)
	})
	if !ok {
		return false, nil
	}
	return repository.RegistrationImportAllowed(ctx, actor, service.credentials.Digest(token), service.now())
}
func (service *Service) CompleteRegistrationImport(ctx context.Context, actor, token string) error {
	if !validRegistrationReceipt(token) {
		return domain.ErrForbidden
	}
	repository, ok := service.repository.(interface {
		CompleteRegistrationImport(context.Context, string, string, time.Time) error
	})
	if !ok {
		return domain.ErrForbidden
	}
	return repository.CompleteRegistrationImport(ctx, actor, service.credentials.Digest(token), service.now())
}
