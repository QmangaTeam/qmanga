package application

import (
	"context"
	"errors"
	"io"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
)

type avatarRepo struct{ saved bool }

func (r *avatarRepo) SaveAvatar(context.Context, Avatar) error { r.saved = true; return nil }
func (r *avatarRepo) Avatar(context.Context, string, string) (Avatar, error) {
	return Avatar{}, domain.ErrNotFound
}

type avatarStorage struct{}

func (avatarStorage) PutDerivative(context.Context, string, io.Reader, int64, string) error {
	return nil
}
func (avatarStorage) OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("safe")), nil
}

type avatarProcessor struct{ called bool }

func (p *avatarProcessor) ProcessAvatar(io.Reader) ([]byte, error) {
	p.called = true
	return []byte("safe"), nil
}

type avatarTeams struct{ allowed bool }

func (p avatarTeams) CheckAvatar(context.Context, string, string) error {
	if p.allowed {
		return nil
	}
	return domain.ErrForbidden
}
func TestAvatarAuthorizationPrecedesProcessing(t *testing.T) {
	for _, scenario := range []struct {
		kind, actor, owner string
		team, allowed      bool
	}{{"user", "me", "other", false, false}, {"user", "me", "me", false, true}, {"team", "me", "team", false, false}, {"team", "me", "team", true, true}, {"bucket", "me", "me", true, false}} {
		repo := &avatarRepo{}
		processor := &avatarProcessor{}
		service := NewAvatars(repo, avatarStorage{}, processor, avatarTeams{scenario.team}, func() string { return "immutable-id" })
		_, err := service.Upload(context.Background(), scenario.actor, scenario.kind, scenario.owner, strings.NewReader("source"))
		if scenario.allowed {
			if err != nil || !repo.saved || !processor.called {
				t.Fatalf("allowed upload failed: %v", err)
			}
		} else if !errors.Is(err, domain.ErrForbidden) || repo.saved || processor.called {
			t.Fatalf("unauthorized upload processed: %+v %v", scenario, err)
		}
	}
}
