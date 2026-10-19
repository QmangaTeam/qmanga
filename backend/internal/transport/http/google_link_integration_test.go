package http

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

type linkingProvider struct {
	identity       application.GoogleIdentity
	beforeExchange func()
}

func (p *linkingProvider) AuthorizationURL(state, nonce, verifier string) string {
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
}
func (p *linkingProvider) Exchange(context.Context, string, string, string) (application.GoogleIdentity, error) {
	if p.beforeExchange != nil {
		p.beforeExchange()
	}
	return p.identity, nil
}
func TestGoogleExplicitLinkHTTPWithPostgres(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	base, err := postgres.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "google_link_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := postgres.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := identityPostgres.New(db.Pool())
	credentials := identityCrypto.New()
	service := application.New(repo, credentials, nil, "https://app.test")
	owner := domain.Account{ID: "link-owner", Email: "owner@example.test", PasswordHash: "hash", Role: "admin", Verified: true, DisplayName: "Edited profile"}
	other := domain.Account{ID: "other-owner", Email: "other@example.test", PasswordHash: "hash", Role: "reader", Verified: true, DisplayName: "Other profile"}
	rawA, rawB, csrf := strings.Repeat("a", 43), strings.Repeat("b", 43), strings.Repeat("c", 43)
	for _, account := range []domain.Account{owner, other} {
		if err = repo.CreateAccount(ctx, account); err != nil {
			t.Fatal(err)
		}
	}
	saveSession := func(raw, actor string) {
		t.Helper()
		_ = repo.DeleteSession(ctx, credentials.Digest(raw))
		if err := repo.SaveSession(ctx, application.Session{Hash: credentials.Digest(raw), UserID: actor, CSRFHash: credentials.Digest(csrf), ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	saveSession(rawA, owner.ID)
	saveSession(rawB, other.ID)
	provider := &linkingProvider{identity: application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "stable-google", Email: owner.Email, Verified: true, Name: "Google name"}}
	google := application.NewGoogle(repo, credentials, provider, service)
	identity := NewIdentityHandler(service, "https://app.test", true)
	router := NewRouter(nil)
	identity.Register(router.Group("/api/v1"))
	NewGoogleHandler(google, identity).Register(router.Group("/api/v1"))
	send := func(method, path, raw, account, token, origin string, cookies []*http.Cookie) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, "/api/v1/auth/google/"+path, strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", origin)
		request.Header.Set("X-Qmanga-Account", account)
		request.Header.Set("X-CSRF-Token", token)
		if raw != "" {
			request.AddCookie(&http.Cookie{Name: "qmanga_session", Value: raw})
		}
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	for _, scenario := range []struct {
		method, path, raw, account, csrf, origin string
		want                                     int
	}{{"GET", "link", "", "", "", "", 401}, {"GET", "link", rawA, "", "", "", 409}, {"GET", "link", rawA, other.ID, "", "", 409}, {"GET", "link", rawA, owner.ID, "", "", 200}, {"POST", "link/start", "", owner.ID, csrf, "https://app.test", 401}, {"POST", "link/start", rawA, owner.ID, "", "https://app.test", 403}, {"POST", "link/start", rawA, owner.ID, csrf, "https://evil.test", 403}, {"POST", "link/start", rawA, other.ID, csrf, "https://app.test", 409}} {
		response := send(scenario.method, scenario.path, scenario.raw, scenario.account, scenario.csrf, scenario.origin, nil)
		if response.Code != scenario.want {
			t.Fatalf("link route protection %s %s: %d want %d", scenario.method, scenario.path, response.Code, scenario.want)
		}
	}
	start := func(raw, actor string) (string, []*http.Cookie) {
		t.Helper()
		response := send("POST", "link/start", raw, actor, csrf, "https://app.test", nil)
		if response.Code != 200 {
			t.Fatalf("start: %d %s", response.Code, response.Body)
		}
		var state string
		for _, cookie := range response.Result().Cookies() {
			if cookie.Name == "qmanga_google_state" {
				state = cookie.Value
			}
			if !cookie.HttpOnly || !cookie.Secure {
				t.Fatal("unsafe link cookie")
			}
		}
		if len(state) != 43 {
			t.Fatal("missing link state")
		}
		return state, response.Result().Cookies()
	}
	for _, scenario := range []string{"logout", "switch", "expired", "unverified", "blocked", "cancel", "claims", "revoked-during-exchange", "stripped-marker"} {
		t.Run(scenario, func(t *testing.T) {
			saveSession(rawA, owner.ID)
			state, cookies := start(rawA, owner.ID)
			raw, code := rawA, "code"
			switch scenario {
			case "logout":
				raw = ""
			case "switch":
				raw = rawB
			case "expired":
				_, err = db.Pool().Exec(ctx, "UPDATE sessions SET expires_at=now()-interval '1 second' WHERE id=$1", credentials.Digest(rawA))
			case "unverified":
				_, err = db.Pool().Exec(ctx, "UPDATE users SET is_verified=false WHERE id=$1", owner.ID)
			case "blocked":
				_, err = db.Pool().Exec(ctx, "UPDATE users SET is_blocked=true WHERE id=$1", owner.ID)
			case "cancel":
				code = ""
			case "claims":
				provider.identity.Verified = false
			case "revoked-during-exchange":
				provider.beforeExchange = func() {
					if err := repo.DeleteSession(ctx, credentials.Digest(rawA)); err != nil {
						t.Fatal(err)
					}
				}
			case "stripped-marker":
				cookies = cookies[:0]
				cookies = append(cookies, &http.Cookie{Name: "qmanga_google_state", Value: state})
			}
			if err != nil {
				t.Fatal(err)
			}
			response := send("GET", "callback?state="+state+"&code="+code, raw, "", "", "", cookies)
			if response.Code != 303 || strings.Contains(response.Header().Get("Location"), "success") {
				t.Fatalf("invalid callback accepted: %d %s", response.Code, response.Header().Get("Location"))
			}
			if linked, err := repo.GoogleLinked(ctx, owner.ID); err != nil || linked {
				t.Fatal("invalid callback created link", err)
			}
			for _, cookie := range response.Result().Cookies() {
				if cookie.Name == "qmanga_session" || cookie.Name == "qmanga_csrf" || cookie.Name == "qmanga_external_complete" || cookie.Name == "qmanga_registration_import" {
					t.Fatal("link changed account/session state")
				}
			}
			if _, err = db.Pool().Exec(ctx, "UPDATE users SET is_verified=true,is_blocked=false WHERE id=$1", owner.ID); err != nil {
				t.Fatal(err)
			}
			provider.identity.Verified = true
			provider.beforeExchange = nil
		})
	}
	saveSession(rawA, owner.ID)
	state, cookies := start(rawA, owner.ID)
	response := send("GET", "callback?state="+state+"&code=code", rawA, "", "", "", cookies)
	if response.Code != 303 || response.Header().Get("Location") != "/settings?google_link=success" {
		t.Fatalf("link failed: %d %s", response.Code, response.Header().Get("Location"))
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "qmanga_session" || cookie.Name == "qmanga_csrf" || cookie.Name == "qmanga_external_complete" {
			t.Fatal("successful link replaced session")
		}
	}
	response = send("GET", "callback?state="+state+"&code=code", rawA, "", "", "", cookies)
	if response.Header().Get("Location") == "/settings?google_link=success" {
		t.Fatal("callback replay accepted")
	}
	status := send("GET", "link", rawA, owner.ID, "", "", nil)
	if status.Code != 200 || !strings.Contains(status.Body.String(), `"linked":true`) || !strings.Contains(status.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("private linked status: %d %s", status.Code, status.Body)
	}
	state, cookies = start(rawB, other.ID)
	response = send("GET", "callback?state="+state+"&code=code", rawB, "", "", "", cookies)
	if response.Header().Get("Location") != "/settings?google_link=conflict" {
		t.Fatal("foreign identity transfer accepted")
	}
	// A subsequent ordinary Google sign-in resolves the original account, with no replacement profile.
	state, _, err = google.Start(ctx, "/catalog", "")
	if err != nil {
		t.Fatal(err)
	}
	response = send("GET", "callback?state="+state+"&code=code", "", "", "", "", []*http.Cookie{{Name: "qmanga_google_state", Value: state}})
	if response.Code != 303 || !strings.HasPrefix(response.Header().Get("Location"), "/auth/complete?") {
		t.Fatalf("linked sign-in failed: %d %s", response.Code, response.Header().Get("Location"))
	}
	var sessionToken string
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "qmanga_session" {
			sessionToken = cookie.Value
		}
	}
	_, actual, err := service.Authenticate(ctx, sessionToken)
	if err != nil || actual != owner {
		t.Fatalf("linked login changed owner: %+v %v", actual, err)
	}
	var count int
	if err = db.Pool().QueryRow(ctx, "SELECT count(*) FROM users").Scan(&count); err != nil || count != 2 {
		t.Fatal("link/login created replacement user", err)
	}
	// A forged link marker on a normal sign-in cannot turn it into linking.
	state, _, err = google.Start(ctx, "/catalog", "")
	if err != nil {
		t.Fatal(err)
	}
	response = send("GET", "callback?state="+state+"&code=code", rawB, "", "", "", []*http.Cookie{{Name: "qmanga_google_state", Value: state}, {Name: "qmanga_google_link", Value: state}})
	if response.Header().Get("Location") != "/settings?google_link=error" {
		t.Fatal("normal flow accepted for linking")
	}
	var linked bool
	status = send("GET", "link", rawB, other.ID, "", "", nil)
	var payload struct {
		Linked bool `json:"linked"`
	}
	if err = json.Unmarshal(status.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	linked = payload.Linked
	if linked {
		t.Fatal("forged marker linked foreign user")
	}
}
