package application

import (
	"context"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

type receiptCredentials struct{ Credentials }

func (*receiptCredentials) Digest(value string) string { return "digest:" + value }

type importRepository struct {
	Repository
	actor, hash string
	complete    bool
}

func (repository *importRepository) RegistrationImportAllowed(_ context.Context, actor, hash string, _ time.Time) (bool, error) {
	return !repository.complete && actor == repository.actor && hash == repository.hash, nil
}
func (repository *importRepository) CompleteRegistrationImport(_ context.Context, actor, hash string, _ time.Time) error {
	if actor == repository.actor && hash == repository.hash {
		repository.complete = true
		return nil
	}
	return domain.ErrForbidden
}
func TestRegistrationImportIsBoundToAccountAndOneTime(t *testing.T) {
	credentials := &receiptCredentials{}
	repository := &importRepository{actor: "new", hash: credentials.Digest(strings.Repeat("a", 43))}
	service := New(repository, credentials, nil, "")
	for _, actor := range []string{"other", "new"} {
		allowed, err := service.RegistrationImportAllowed(context.Background(), actor, strings.Repeat("a", 43))
		if err != nil || allowed != (actor == "new") {
			t.Fatal(actor, allowed, err)
		}
	}
	if err := service.CompleteRegistrationImport(context.Background(), "new", strings.Repeat("a", 43)); err != nil {
		t.Fatal(err)
	}
	if err := service.CompleteRegistrationImport(context.Background(), "new", strings.Repeat("a", 43)); err != nil {
		t.Fatal("completion must be idempotent", err)
	}
	allowed, _ := service.RegistrationImportAllowed(context.Background(), "new", strings.Repeat("a", 43))
	if allowed {
		t.Fatal("receipt reused")
	}
	allowed, _ = service.RegistrationImportAllowed(context.Background(), "new", "invalid")
	if allowed {
		t.Fatal("invalid receipt accepted")
	}
}
