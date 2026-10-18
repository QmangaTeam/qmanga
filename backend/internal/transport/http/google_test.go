package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
	"time"
)

func TestGoogleDisabledByDefaultAndCallbackRequiresCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	identity := NewIdentityHandler(nil, "https://app.test", true)
	NewGoogleHandler(nil, identity).Register(router.Group("/api/v1"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/auth/google/status", nil))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"enabled":false`) {
		t.Fatalf("disabled state: %d %s", response.Code, response.Body)
	}
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/auth/google/start", nil))
	if response.Code != 404 {
		t.Fatal("disabled endpoint exposed")
	}
	router = gin.New()
	NewGoogleHandler(&application.GoogleService{}, identity).Register(router.Group("/api/v1"))
	for _, cookie := range []string{"", strings.Repeat("b", 43)} {
		request := httptest.NewRequest("GET", "/api/v1/auth/google/callback?state="+strings.Repeat("a", 43)+"&code=attacker", nil)
		if cookie != "" {
			request.AddCookie(&http.Cookie{Name: "qmanga_google_state", Value: cookie})
		}
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != 400 {
			t.Fatal("callback without matching browser state accepted")
		}
		for _, cookie := range response.Result().Cookies() {
			if cookie.Name == "qmanga_google_state" && (!cookie.HttpOnly || !cookie.Secure || cookie.MaxAge != -1) {
				t.Fatal("unsafe state cookie")
			}
		}
	}
}

type completionGoogleService struct {
	destination string
	failure     error
}

func (s completionGoogleService) Start(context.Context, string, string) (string, string, error) {
	return "", "", nil
}
func (s completionGoogleService) Finish(context.Context, string, string) (application.LoginResult, string, error) {
	return application.LoginResult{Account: domain.Account{DisplayName: "Reader"}, SessionToken: "ordinary-session", CSRFToken: "ordinary-csrf", ExpiresAt: time.Now().Add(time.Hour)}, s.destination, s.failure
}
func TestGoogleSuccessUsesBrowserAccountTransitionAndCancellationDoesNot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []struct {
		destination, want string
		cancel            bool
	}{{"/read/release?mode=strip", "/auth/complete?returnTo=%2Fread%2Frelease%3Fmode%3Dstrip", false}, {"/auth/complete?returnTo=/auth/complete", "/auth/complete?returnTo=%2F", false}, {"//evil.example", "/auth/complete?returnTo=%2F", false}, {"/catalog?genre=adventure", "/catalog?genre=adventure", true}} {
		router := gin.New()
		identity := NewIdentityHandler(nil, "https://app.test", true)
		service := completionGoogleService{destination: scenario.destination}
		if scenario.cancel {
			service.failure = domain.ErrCredentials
		}
		handler := &GoogleHandler{service: service, identity: identity}
		handler.Register(router.Group("/api/v1"))
		path := "/api/v1/auth/google/callback?state=" + strings.Repeat("a", 43) + "&code=code"
		if scenario.cancel {
			path += "&error=access_denied"
		}
		request := httptest.NewRequest("GET", path, nil)
		request.AddCookie(&http.Cookie{Name: "qmanga_google_state", Value: strings.Repeat("a", 43)})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		want := scenario.want
		if !scenario.cancel {
			want += "&nonce=" + strings.Repeat("a", 43)
		}
		if response.Code != 303 || response.Header().Get("Location") != want {
			t.Fatalf("%s -> %d %s", scenario.destination, response.Code, response.Header().Get("Location"))
		}
		session := false
		for _, cookie := range response.Result().Cookies() {
			if cookie.Name == "qmanga_session" {
				session = true
			}
		}
		if session == scenario.cancel {
			t.Fatal("session cookie success/cancel mismatch")
		}
		proof := false
		for _, cookie := range response.Result().Cookies() {
			if cookie.Name == "qmanga_external_complete" {
				proof = true
				expected := sha256.Sum256([]byte(strings.Repeat("a", 43) + "\x00ordinary-session"))
				if cookie.Value != hex.EncodeToString(expected[:]) {
					t.Fatal("proof is not bound to nonce and new session")
				}
				if !cookie.HttpOnly || !cookie.Secure || len(cookie.Value) != 64 || cookie.MaxAge > 60 {
					t.Fatal("unsafe completion proof")
				}
			}
		}
		if proof == scenario.cancel {
			t.Fatal("completion proof success/cancel mismatch")
		}

	}
}
