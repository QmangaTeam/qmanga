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

func (r *avatarRepo) ReserveAvatarObject(context.Context, string) error { return nil }
func (r *avatarRepo) AbandonAvatarObject(context.Context, string) error { return nil }
func (r *avatarRepo) AvatarGarbage(context.Context) ([]string, error)   { return nil, nil }
func (r *avatarRepo) ForgetAvatarGarbage(context.Context, string) error { return nil }
func (r *avatarRepo) SaveAvatar(context.Context, Avatar) error          { r.saved = true; return nil }
func (r *avatarRepo) Avatar(context.Context, string, string) (Avatar, error) {
	return Avatar{}, domain.ErrNotFound
}

type avatarStorage struct{}

func (avatarStorage) DeleteDerivative(context.Context, string) error { return nil }

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

type cleanupRepo struct {
	avatarRepo
	failed   bool
	reserved bool
	queue    []string
	forgot   []string
}

func (r *cleanupRepo) ReserveAvatarObject(context.Context, string) error {
	r.reserved = true
	return nil
}
func (r *cleanupRepo) SaveAvatar(_ context.Context, a Avatar) error {
	if r.failed {
		return errors.New("database unavailable")
	}
	r.queue = []string{"avatars/obsolete"}
	return nil
}
func (r *cleanupRepo) AbandonAvatarObject(_ context.Context, key string) error {
	r.queue = append(r.queue, key)
	return nil
}
func (r *cleanupRepo) AvatarGarbage(context.Context) ([]string, error) { return r.queue, nil }
func (r *cleanupRepo) ForgetAvatarGarbage(_ context.Context, key string) error {
	r.forgot = append(r.forgot, key)
	return nil
}

type cleanupStorage struct {
	avatarStorage
	deleted []string
}

func (s *cleanupStorage) DeleteDerivative(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}
func TestAvatarCleansFailedAndReplacedObjects(t *testing.T) {
	for _, failed := range []bool{false, true} {
		repo := &cleanupRepo{failed: failed}
		storage := &cleanupStorage{}
		service := NewAvatars(repo, storage, &avatarProcessor{}, avatarTeams{}, func() string { return "new" })
		_, err := service.Upload(context.Background(), "me", "user", "me", strings.NewReader("image"))
		if !repo.reserved {
			t.Fatal("object not durably reserved")
		}
		expected := "avatars/obsolete"
		if failed {
			expected = "avatars/new"
			if err == nil {
				t.Fatal("save failure ignored")
			}
		}
		if len(storage.deleted) != 1 || storage.deleted[0] != expected || len(repo.forgot) != 1 {
			t.Fatalf("cleanup failed: %+v %+v", storage.deleted, repo.forgot)
		}
	}
}
