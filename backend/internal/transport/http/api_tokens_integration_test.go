package http

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	"qmanga/internal/identity/application"
	"qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestAPITokenHTTPPostgres(testingTarget *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		testingTarget.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, failure := postgres.New(ctx, databaseURL)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Close()
	schema := "identity_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, failure = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	database, failure := postgres.New(ctx, parsed.String())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		testingTarget.Fatal(failure)
	}

	origin := "http://localhost:8080"
	credentials := identityCrypto.New()
	service := application.New(identityPostgres.New(database.Pool()), credentials, nil, origin)
	owner, e := service.CreateBootstrapAdmin(ctx, "tokens@test.local", "password123")
	if e != nil {
		testingTarget.Fatal(e)
	}
	login, e := service.Login(ctx, "tokens@test.local", "password123")
	if e != nil {
		testingTarget.Fatal(e)
	}
	handler := NewIdentityHandler(service, origin, false)
	router := NewRouter(nil)
	api := router.Group("/api/v1")
	handler.Register(api)
	handler.RegisterAPITokens(api)
	api.POST("/studio/revisions/:id/publish", handler.RequireUser(), expectedAccount(), func(c *gin.Context) {
		id, ok := DirectPublishToken(c)
		if !ok || id == "" {
			c.Status(403)
			return
		}
		c.Status(204)
	})
	send := func(method, path, body, bearer string, cookie, csrf bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Qmanga-Account", owner)
		if bearer != "" {
			r.Header.Set("Authorization", bearer)
		}
		if cookie {
			r.AddCookie(&http.Cookie{Name: "qmanga_session", Value: login.SessionToken})
		}
		if csrf {
			r.Header.Set("X-CSRF-Token", login.CSRFToken)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	expect := func(w *httptest.ResponseRecorder, status int) {
		testingTarget.Helper()
		if w.Code != status {
			testingTarget.Fatalf("status %d want %d body %s", w.Code, status, w.Body.String())
		}
	}
	body := `{"name":"Codex","expires_in_days":2,"direct_publish":true}`
	expect(send("POST", "/admin/api-tokens", body, "", true, false), 403)
	created := send("POST", "/admin/api-tokens", body, "", true, true)
	expect(created, 201)
	var result struct {
		Token string               `json:"token"`
		Item  application.APIToken `json:"item"`
	}
	if e = json.Unmarshal(created.Body.Bytes(), &result); e != nil {
		testingTarget.Fatal(e)
	}
	if len(result.Token) != 47 {
		testingTarget.Fatal("missing secret")
	}
	var hash string
	if e = database.Pool().QueryRow(ctx, "SELECT token_hash FROM api_tokens WHERE id=$1", result.Item.ID).Scan(&hash); e != nil || hash != credentials.Digest(result.Token) || hash == result.Token {
		testingTarget.Fatal("hash storage", e)
	}
	listed := send("GET", "/admin/api-tokens", "", "", true, false)
	expect(listed, 200)
	if strings.Contains(listed.Body.String(), result.Token) || strings.Contains(listed.Body.String(), hash) {
		testingTarget.Fatal("list disclosed secret")
	}
	bearer := "Bearer " + result.Token
	expect(send("GET", "/auth/me", "", bearer, false, false), 200)
	for _, headers := range []map[string][]string{
		{"Authorization": {bearer, bearer}},
		{"Authorization": {bearer}, "X-Qmanga-Account": {"other"}},
		{"Authorization": {""}},
	} {
		r := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		r.AddCookie(&http.Cookie{Name: "qmanga_session", Value: login.SessionToken})
		for k, v := range headers {
			r.Header[k] = v
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != 401 && w.Code != 403 {
			testingTarget.Fatalf("conflicting authorization accepted: %d", w.Code)
		}
	}
	r := httptest.NewRequest("POST", "/api/v1/studio/revisions/x/publish", nil)
	r.Header.Set("Authorization", bearer)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	expect(w, 204)

	expect(send("POST", "/studio/revisions/x/publish", "", bearer, false, false), 204)
	expect(send("GET", "/admin/api-tokens", "", bearer, true, false), 403)
	expect(send("GET", "/auth/me", "", "Bearer invalid", true, false), 401)
	for _, sql := range []string{"UPDATE users SET role='reader' WHERE id=$1", "UPDATE users SET role='admin',is_verified=false WHERE id=$1", "UPDATE users SET is_verified=true,is_blocked=true WHERE id=$1"} {
		if _, e = database.Pool().Exec(ctx, sql, owner); e != nil {
			testingTarget.Fatal(e)
		}
		expect(send("GET", "/auth/me", "", bearer, false, false), 401)
	}
	if _, e = database.Pool().Exec(ctx, "UPDATE users SET role='admin',is_verified=true,is_blocked=false WHERE id=$1", owner); e != nil {
		testingTarget.Fatal(e)
	}

	// Expiry is enforced on every request, independently of a previous successful use.
	if _, e = database.Pool().Exec(ctx, "UPDATE api_tokens SET expires_at=created_at-interval '1 second',created_at=created_at-interval '2 seconds' WHERE id=$1", result.Item.ID); e != nil {
		testingTarget.Fatal(e)
	}
	expect(send("GET", "/auth/me", "", bearer, false, false), 401)
	if _, e = database.Pool().Exec(ctx, "UPDATE api_tokens SET expires_at=now()+interval '1 day' WHERE id=$1", result.Item.ID); e != nil {
		testingTarget.Fatal(e)
	}
	// Revocation is owner scoped even for another current administrator.
	other, e := service.CreateBootstrapAdmin(ctx, "other@test.local", "password123")
	if e != nil {
		testingTarget.Fatal(e)
	}
	if e = service.RevokeAPIToken(ctx, other, result.Item.ID); e == nil {
		testingTarget.Fatal("another owner revoked token")
	}
	expect(send("GET", "/auth/me", "", bearer, false, false), 200)
	expect(send("DELETE", "/admin/api-tokens/"+result.Item.ID, "", "", true, true), 204)
	expect(send("GET", "/auth/me", "", bearer, false, false), 401)
}
