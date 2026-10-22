package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

type profileIdentityRepository struct {
	application.Repository
	failure error
}

func (r profileIdentityRepository) SessionAccount(context.Context, string, time.Time) (application.Session, domain.Account, error) {
	return application.Session{CSRFHash: identityCrypto.New().Digest(strings.Repeat("c", 43))}, domain.Account{ID: "owner", Email: "secret@test.local", Verified: true}, r.failure
}

type profileRepositoryStub struct {
	application.ProfileRepository
	viewer   string
	saved    bool
	followed bool
}

func (r *profileRepositoryStub) PublicProfile(_ context.Context, viewer, target string) (application.PublicProfile, error) {
	r.viewer = viewer
	return application.PublicProfile{ID: target, IsOwner: viewer == target}, nil
}
func (r *profileRepositoryStub) SavePreferences(_ context.Context, _ string, p application.Preferences) (application.Preferences, error) {
	r.saved = true
	p.Version++
	return p, nil
}
func (r *profileRepositoryStub) FollowProfile(context.Context, string, string, bool) error {
	r.followed = true
	return nil
}
func TestProfileOptionalIdentityAndPrivateCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []struct {
		name, cookie string
		failure      error
		status       int
		viewer       string
	}{{"guest", "", nil, 200, ""}, {"invalid", "short", nil, 200, ""}, {"expired", strings.Repeat("s", 43), domain.ErrCredentials, 200, ""}, {"owner", strings.Repeat("s", 43), nil, 200, "owner"}, {"auth outage", strings.Repeat("s", 43), errors.New("offline"), 503, ""}} {
		t.Run(scenario.name, func(t *testing.T) {
			repo := &profileRepositoryStub{}
			id := NewIdentityHandler(application.New(profileIdentityRepository{failure: scenario.failure}, identityCrypto.New(), nil, "https://app.test"), "https://app.test", true)
			router := gin.New()
			NewProfileHandler(application.NewProfiles(repo), id).Register(router.Group("/api/v1"))
			req := httptest.NewRequest("GET", "/api/v1/profiles/owner", nil)
			if scenario.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "qmanga_session", Value: scenario.cookie})
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != scenario.status || repo.viewer != scenario.viewer {
				t.Fatalf("%d %s viewer=%s", response.Code, response.Body, repo.viewer)
			}
			if response.Header().Get("Cache-Control") != "private, no-store" || strings.Contains(response.Body.String(), "email") {
				t.Fatal("private data/cache failure")
			}
		})
	}
}
func TestProfileWritesRequireAccountCSRFAndCompletePayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &profileRepositoryStub{}
	id := NewIdentityHandler(application.New(profileIdentityRepository{}, identityCrypto.New(), nil, "https://app.test"), "https://app.test", true)
	router := gin.New()
	NewProfileHandler(application.NewProfiles(repo), id).Register(router.Group("/api/v1"))
	for _, scenario := range []struct {
		path, account, csrf, body string
		status                    int
	}{{"/account/preferences", "owner", "", `{}`, 403}, {"/account/preferences", "other", strings.Repeat("c", 43), `{}`, 409}, {"/account/preferences", "owner", strings.Repeat("c", 43), `{"version":0}`, 422}, {"/profiles/target/follow", "owner", strings.Repeat("c", 43), `{}`, 422}, {"/profiles/target/follow", "owner", strings.Repeat("c", 43), `{"following":true}`, 200}} {
		req := httptest.NewRequest("PUT", "/api/v1"+scenario.path, strings.NewReader(scenario.body))
		req.AddCookie(&http.Cookie{Name: "qmanga_session", Value: strings.Repeat("s", 43)})
		req.Header.Set("Origin", "https://app.test")
		req.Header.Set("X-CSRF-Token", scenario.csrf)
		req.Header.Set("X-Qmanga-Account", scenario.account)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != scenario.status {
			t.Fatalf("%s: %d %s", scenario.path, res.Code, res.Body)
		}
	}
	if repo.saved || !repo.followed {
		t.Fatal("unexpected writes")
	}
}
