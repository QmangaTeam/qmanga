package application

import (
	"context"
	"errors"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

type googleRepo struct {
	flow    GoogleFlow
	exists  bool
	claimed bool
	account domain.Account
}

func (r *googleRepo) SaveGoogleFlow(_ context.Context, f GoogleFlow) error {
	r.flow = f
	r.exists = true
	return nil
}
func (r *googleRepo) ConsumeGoogleFlow(_ context.Context, hash string, now time.Time) (GoogleFlow, error) {
	if !r.exists || hash != r.flow.Hash || !now.Before(r.flow.ExpiresAt) {
		return GoogleFlow{}, domain.ErrExpired
	}
	r.exists = false
	return r.flow, nil
}
func (r *googleRepo) GoogleAccount(_ context.Context, i GoogleIdentity, a domain.Account) (domain.Account, error) {
	r.claimed = true
	r.account = a
	return a, domain.ErrConflict
}

type googleCredentials struct{ n int }

func (c *googleCredentials) RandomToken() (string, error) {
	c.n++
	return strings.Repeat(string(rune('a'+c.n)), 43), nil
}
func (c *googleCredentials) Digest(s string) string        { return s }
func (c *googleCredentials) NewID() string                 { return "user" }
func (c *googleCredentials) Hash(s string) (string, error) { return "hash:" + s, nil }
func (c *googleCredentials) Verify(string, string) bool    { return false }

type googleProvider struct {
	identity GoogleIdentity
	failed   bool
}

func (p googleProvider) AuthorizationURL(state, nonce, verifier string) string {
	return state + ":" + nonce + ":" + verifier
}
func (p googleProvider) Exchange(context.Context, string, string, string) (GoogleIdentity, error) {
	if p.failed {
		return GoogleIdentity{}, errors.New("bad token")
	}
	return p.identity, nil
}
func TestGoogleFlowIsOneUseAndRequiresVerifiedStableIdentity(t *testing.T) {
	good := GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "stable-sub", Email: "Reader@Example.com", Verified: true}
	for _, identity := range []GoogleIdentity{good, {Issuer: good.Issuer, Subject: good.Subject, Email: good.Email}, {Issuer: "https://evil.test", Subject: good.Subject, Email: good.Email, Verified: true}, {Issuer: good.Issuer, Email: good.Email, Verified: true}} {
		repo := &googleRepo{}
		service := NewGoogle(repo, &googleCredentials{}, googleProvider{identity: identity}, nil)
		state, _, err := service.Start(context.Background(), "//evil.test", "Оқырман")
		if err != nil {
			t.Fatal(err)
		}
		if repo.flow.ReturnPath != "/" {
			t.Fatal("unsafe redirect saved")
		}
		_, _, err = service.Finish(context.Background(), state, "code")
		if identity == good {
			if !errors.Is(err, domain.ErrConflict) || !repo.claimed || repo.account.Email != "reader@example.com" {
				t.Fatalf("identity not passed safely: %v", err)
			}
		} else if !errors.Is(err, domain.ErrCredentials) || repo.claimed {
			t.Fatalf("untrusted identity accepted: %+v %v", identity, err)
		}
		if _, _, err = service.Finish(context.Background(), state, "code"); !errors.Is(err, domain.ErrExpired) {
			t.Fatalf("replay accepted: %v", err)
		}
	}
}
func TestGoogleCancellationRetainsPageAndConsumesFlow(t *testing.T) {
	repo := &googleRepo{}
	service := NewGoogle(repo, &googleCredentials{}, googleProvider{}, nil)
	state, _, _ := service.Start(context.Background(), "/titles/a?tab=chapters", "")
	_, path, err := service.Finish(context.Background(), state, "")
	if err == nil || path != "/titles/a?tab=chapters" || repo.exists {
		t.Fatalf("bad cancel: %s %v", path, err)
	}
}

func TestGoogleReturnPathRejectsBrowserNormalizedRedirects(t *testing.T) {
	for _, value := range []string{"//evil.example", "/\t/evil.example", "/\n/evil.example", "/%09/evil.example", "/%2f/evil.example", "/%252f/evil.example", "/\\evil.example", "/%5cevil.example", "https://evil.example", "/ path"} {
		if got := SafeAuthReturn(value); got != "/" {
			t.Errorf("unsafe redirect %q retained: %q", value, got)
		}
	}
	for _, value := range []string{"/", "/catalog?query=reader", "/titles/a?tab=chapters", "/read/chapter#page=2"} {
		if got := SafeAuthReturn(value); got != value {
			t.Errorf("local redirect rejected: %q", value)
		}
	}
}

func TestGoogleRegistrationDefaultsNicknameWithoutManualInput(t *testing.T) {
	repo := &googleRepo{}
	service := NewGoogle(repo, &googleCredentials{}, googleProvider{identity: GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "stable", Email: "reader@example.com", Verified: true}}, nil)
	state, _, err := service.Start(context.Background(), "/settings", "")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = service.Finish(context.Background(), state, "code")
	if !errors.Is(err, domain.ErrConflict) || repo.account.DisplayName != "reader" {
		t.Fatalf("Google nickname default missing: %q %v", repo.account.DisplayName, err)
	}
}

type googleLoginRepo struct {
	googleRepo
	Repository
	existing domain.Account
}

func (r *googleLoginRepo) GoogleAccount(_ context.Context, _ GoogleIdentity, candidate domain.Account) (domain.Account, error) {
	if r.existing.ID != "" {
		return r.existing, nil
	}
	return candidate, nil
}
func (r *googleLoginRepo) SaveSession(context.Context, Session) error { return nil }

type googleAvatarRecorder struct {
	calls          int
	owner, picture string
}

func (a *googleAvatarRecorder) ImportGoogleAvatar(_ context.Context, owner, picture string) error {
	a.calls++
	a.owner = owner
	a.picture = picture
	return errors.New("picture unavailable")
}
func TestGoogleDefaultsOnlyApplyToNewAccount(t *testing.T) {
	for _, existing := range []domain.Account{{}, {ID: "existing", Email: "edited@example.com", DisplayName: "Edited nickname", Verified: true, Role: "reader"}} {
		t.Run(existing.ID, func(t *testing.T) {
			repo := &googleLoginRepo{existing: existing}
			creds := &googleCredentials{}
			avatars := &googleAvatarRecorder{}
			service := NewGoogle(repo, creds, googleProvider{identity: GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "stable", Email: "reader@example.com", Name: "  Әли 😊 O’Connor  ", Picture: "https://lh3.googleusercontent.com/a/photo", Verified: true}}, New(repo, creds, nil, ""), avatars)
			state, _, err := service.Start(context.Background(), "/catalog", "")
			if err != nil {
				t.Fatal(err)
			}
			result, path, err := service.Finish(context.Background(), state, "code")
			if err != nil || path != "/catalog" || result.SessionToken == "" {
				t.Fatalf("login failed: %v", err)
			}
			if existing.ID == "" {
				if result.Account.DisplayName != "Әли OConnor" || result.Account.Email != "reader@example.com" || avatars.calls != 1 || avatars.owner != result.Account.ID {
					t.Fatalf("new defaults absent: %+v %+v", result.Account, avatars)
				}
			} else if result.Account != existing || avatars.calls != 0 {
				t.Fatalf("existing edits overwritten: %+v %+v", result.Account, avatars)
			}
		})
	}
}
func TestGoogleNicknameHandlesUnusualNames(t *testing.T) {
	for _, input := range []struct{ name, email, want string }{{"A", "x@example.com", "Google reader"}, {"😊", "reader@example.com", "reader"}, {strings.Repeat("名", 50), "x@example.com", strings.Repeat("名", 40)}, {"  Name\n Reader ", "x@example.com", "Name Reader"}} {
		if got := googleDisplayName(input.name, input.email); got != input.want {
			t.Errorf("nickname=%q want=%q", got, input.want)
		}
	}
}
