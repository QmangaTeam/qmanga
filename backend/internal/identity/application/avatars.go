package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"qmanga/internal/identity/domain"
	"time"
)

const MaxAvatarBytes int64 = 5 << 20

type Avatar struct {
	Kind, OwnerID, Key string
	Bytes              int64
	UpdatedAt          time.Time
}
type AvatarRepository interface {
	TryAvatarCleanupGuard(context.Context) (func(), bool, error)
	ReserveAvatarObject(context.Context, string) error
	AbandonAvatarObject(context.Context, string) error
	AvatarGarbage(context.Context) ([]string, error)
	ForgetAvatarGarbage(context.Context, string) error
	SaveAvatar(context.Context, Avatar) error
	Avatar(context.Context, string, string) (Avatar, error)
}
type AvatarStorage interface {
	DeleteDerivative(context.Context, string) error
	PutDerivative(context.Context, string, io.Reader, int64, string) error
	OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error)
}
type AvatarProcessor interface {
	ProcessAvatar(io.Reader) ([]byte, error)
}
type AvatarTeams interface {
	CheckAvatar(context.Context, string, string) error
}
type AvatarService struct {
	repository AvatarRepository
	storage    AvatarStorage
	processor  AvatarProcessor
	teams      AvatarTeams
	newID      func() string
}

func NewAvatars(repository AvatarRepository, storage AvatarStorage, processor AvatarProcessor, teams AvatarTeams, newID func() string) *AvatarService {
	return &AvatarService{repository, storage, processor, teams, newID}
}
func (s *AvatarService) Upload(ctx context.Context, actor, kind, owner string, source io.Reader) (Avatar, error) {
	if actor == "" || (kind != "user" && kind != "team") || owner == "" {
		return Avatar{}, domain.ErrForbidden
	}
	if kind == "user" && actor != owner {
		return Avatar{}, domain.ErrForbidden
	}
	if kind == "team" {
		if err := s.teams.CheckAvatar(ctx, actor, owner); err != nil {
			return Avatar{}, domain.ErrForbidden
		}
	}
	raw, err := s.processor.ProcessAvatar(io.LimitReader(source, MaxAvatarBytes+1))
	if err != nil {
		return Avatar{}, domain.ErrInvalid
	}
	avatar := Avatar{Kind: kind, OwnerID: owner, Key: "avatars/" + s.newID(), Bytes: int64(len(raw)), UpdatedAt: time.Now().UTC()}
	if err = s.repository.ReserveAvatarObject(ctx, avatar.Key); err != nil {
		return Avatar{}, err
	}
	if err = s.storage.PutDerivative(ctx, avatar.Key, bytes.NewReader(raw), avatar.Bytes, "image/jpeg"); err != nil {
		s.cleanupAvatar(avatar.Key)
		return Avatar{}, err
	}
	err = s.repository.SaveAvatar(ctx, avatar)
	if err != nil {
		s.cleanupAvatar(avatar.Key)
	} else {
		s.cleanupAvatar("")
	}
	return avatar, err
}
func (s *AvatarService) Open(ctx context.Context, kind, owner string) (Avatar, io.ReadCloser, error) {
	if kind != "user" && kind != "team" {
		return Avatar{}, nil, domain.ErrNotFound
	}
	avatar, err := s.repository.Avatar(ctx, kind, owner)
	if err != nil {
		return avatar, nil, err
	}
	stream, err := s.storage.OpenDerivative(ctx, avatar.Key, 0, -1)
	return avatar, stream, err
}

// Pending objects and obsolete versions have durable deletion records. Never delete a current key.
func (s *AvatarService) cleanupAvatar(failedKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if failedKey != "" {
		_ = s.repository.AbandonAvatarObject(ctx, failedKey)
	}
	_ = s.DrainAvatarGarbage(ctx)
}

// DrainAvatarGarbage retries durable cleanup independently of uploads, with bounded work and time.
func (s *AvatarService) DrainAvatarGarbage(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	release, acquired, err := s.repository.TryAvatarCleanupGuard(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	} // An online backup owns the deletion guard.
	defer release()
	keys, err := s.repository.AvatarGarbage(ctx)
	if err != nil {
		return err
	}
	var failures []error
	for index, key := range keys {
		if index >= 20 {
			break
		}
		if err = ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		if err = s.storage.DeleteDerivative(ctx, key); err != nil {
			failures = append(failures, err)
			continue
		}
		if err = s.repository.ForgetAvatarGarbage(ctx, key); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
