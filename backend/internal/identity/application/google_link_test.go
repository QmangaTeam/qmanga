package application

import (
	"context"
	"errors"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

type googleLinkRepo struct {
	googleRepo
	Repository
	session  Session
	actor    domain.Account
	revoked  bool
	links    int
	identity GoogleIdentity
}

func (r *googleLinkRepo) SessionAccount(_ context.Context, hash string, now time.Time) (Session, domain.Account, error) {
	if r.revoked || hash != r.session.Hash || !now.Before(r.session.ExpiresAt) {
		return Session{}, domain.Account{}, domain.ErrCredentials
	}
	return r.session, r.actor, nil
}
func (r *googleLinkRepo) LinkGoogleIdentity(_ context.Context, actor, session string, identity GoogleIdentity, now time.Time) error {
	if r.revoked || actor != r.actor.ID || session != r.session.Hash || !now.Before(r.session.ExpiresAt) {
		return domain.ErrForbidden
	}
	r.links++
	r.identity = identity
	return nil
}
func (r *googleLinkRepo) GoogleLinked(context.Context, string) (bool, error) { return r.links > 0, nil }
func linkService() (*GoogleService, *googleLinkRepo) {
	repo := &googleLinkRepo{session: Session{Hash: strings.Repeat("s", 43), UserID: "owner", ExpiresAt: time.Now().Add(time.Hour)}, actor: domain.Account{ID: "owner", Email: "manual@example.com", DisplayName: "Manual", Role: "admin", Verified: true}}
	creds := &googleCredentials{}
	return NewGoogle(repo, creds, googleProvider{identity: GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "google", Email: "google@example.com", Verified: true}}, New(repo, creds, nil, "")), repo
}
func TestGoogleLinkBindsActorAndExactLiveSession(t *testing.T) {
	for _, scenario := range []string{"success", "logout", "switched", "other_session", "expired", "unverified", "cancel", "bad_claims"} {
		t.Run(scenario, func(t *testing.T) {
			service, repo := linkService()
			state, _, err := service.StartLink(context.Background(), "owner", strings.Repeat("s", 43))
			if err != nil {
				t.Fatal(err)
			}
			token, code := strings.Repeat("s", 43), "code"
			switch scenario {
			case "logout":
				repo.revoked = true
			case "switched":
				repo.actor.ID = "another"
			case "other_session":
				token = "another-session"
			case "expired":
				repo.session.ExpiresAt = time.Now().Add(-time.Minute)
			case "unverified":
				repo.actor.Verified = false
			case "cancel":
				code = ""
			case "bad_claims":
				service.provider = googleProvider{identity: GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "google", Email: "google@example.com"}}
			}
			err = service.FinishLink(context.Background(), state, code, token)
			if scenario == "success" {
				if err != nil || repo.links != 1 {
					t.Fatalf("link absent: %v", err)
				}
			} else if err == nil || repo.links != 0 {
				t.Fatalf("invalid link accepted: %v", err)
			}
			if err = service.FinishLink(context.Background(), state, "code", strings.Repeat("s", 43)); err == nil {
				t.Fatal("link replay accepted")
			}
		})
	}
}
func TestGoogleLinkStartRequiresVerifiedMatchingAccount(t *testing.T) {
	for _, scenario := range []string{"mismatch", "unverified", "revoked"} {
		t.Run(scenario, func(t *testing.T) {
			service, repo := linkService()
			actor := "owner"
			switch scenario {
			case "mismatch":
				actor = "another"
			case "unverified":
				repo.actor.Verified = false
			case "revoked":
				repo.revoked = true
			}
			if _, _, err := service.StartLink(context.Background(), actor, strings.Repeat("s", 43)); err == nil || repo.exists {
				t.Fatal("invalid link start saved")
			}
		})
	}
}
func TestGoogleLinkAndLoginFlowsCannotBeConfused(t *testing.T) {
	service, repo := linkService()
	state, _, _ := service.Start(context.Background(), "/", "")
	if err := service.FinishLink(context.Background(), state, "code", strings.Repeat("s", 43)); err == nil || repo.links != 0 {
		t.Fatal("login used for linking")
	}
	state, _, _ = service.StartLink(context.Background(), "owner", strings.Repeat("s", 43))
	if _, _, err := service.Finish(context.Background(), state, "code"); !errors.Is(err, domain.ErrForbidden) || repo.claimed {
		t.Fatalf("link used for login: %v", err)
	}
}
