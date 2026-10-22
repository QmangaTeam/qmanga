package application

import (
	"context"
	"qmanga/internal/identity/domain"
	"strings"
	"time"
	"unicode/utf8"
)

// Preferences version zero represents an account which has never saved preferences.
type Preferences struct {
	Version                                                     int64
	Bio                                                         string
	PublicProfile, ShowLibrary, ShowActivity, AllowFollow       bool
	NotifyReleases, NotifyReplies, NotifyFollowers, EmailDigest bool
}

func DefaultPreferences() Preferences {
	return Preferences{AllowFollow: true, NotifyReleases: true, NotifyReplies: true}
}
func ValidatePreferences(p Preferences) error {
	if p.Version < 0 || !utf8.ValidString(p.Bio) || utf8.RuneCountInString(p.Bio) > 500 || strings.ContainsRune(p.Bio, 0) {
		return domain.ErrInvalid
	}
	return nil
}

type ProfileLibrary struct{ TitleID, Slug, Name, Status, CoverURL string }
type ProfileActivity struct {
	Label, Href string
	CreatedAt   time.Time
}
type PublicProfile struct {
	ID, DisplayName, Bio, AvatarURL                                             string
	CreatedAt                                                                   time.Time
	PublicProfile, ShowLibrary, ShowActivity, AllowFollow, IsOwner, IsFollowing bool
	FollowersCount, FollowingCount                                              int64
	Library                                                                     []ProfileLibrary
	Activity                                                                    []ProfileActivity
}
type ProfileRepository interface {
	Preferences(context.Context, string) (Preferences, error)
	SavePreferences(context.Context, string, Preferences) (Preferences, error)
	PublicProfile(context.Context, string, string) (PublicProfile, error)
	FollowProfile(context.Context, string, string, bool) error
}
type ProfileService struct{ repository ProfileRepository }

func NewProfiles(repository ProfileRepository) *ProfileService { return &ProfileService{repository} }
func (s *ProfileService) Preferences(ctx context.Context, actor string) (Preferences, error) {
	return s.repository.Preferences(ctx, actor)
}
func (s *ProfileService) SavePreferences(ctx context.Context, actor string, p Preferences) (Preferences, error) {
	if err := ValidatePreferences(p); err != nil {
		return Preferences{}, err
	}
	return s.repository.SavePreferences(ctx, actor, p)
}
func (s *ProfileService) Profile(ctx context.Context, viewer, target string) (PublicProfile, error) {
	if len(target) > 36 || target == "" {
		return PublicProfile{}, domain.ErrNotFound
	}
	return s.repository.PublicProfile(ctx, viewer, target)
}
func (s *ProfileService) Follow(ctx context.Context, actor, target string, following bool) error {
	if actor == target || target == "" || len(target) > 36 {
		return domain.ErrInvalid
	}
	return s.repository.FollowProfile(ctx, actor, target, following)
}
