package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	"qmanga/internal/identity/application"
	"qmanga/internal/platform/postgres"
)

type capturedMail struct{ messages []string }

func (mailer *capturedMail) Send(ctx context.Context, to, subject, body string) error {
	mailer.messages = append(mailer.messages, body)
	return nil
}
func (mailer *capturedMail) token() string {
	parts := strings.Split(mailer.messages[len(mailer.messages)-1], "token=")
	return parts[1][:43]
}

func TestIdentityFlowWithPostgres(testingTarget *testing.T) {
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
	mailer := &capturedMail{}
	service := application.New(identityPostgres.New(database.Pool()), identityCrypto.New(), mailer, "http://localhost:8080")
	handler := NewIdentityHandler(service, "http://localhost:8080", false)
	router := NewRouter(nil)
	handler.Register(router.Group("/api/v1"))
	send := func(method, path string, body map[string]string, cookies []*http.Cookie, csrf, origin string) *httptest.ResponseRecorder {
		encoded, _ := json.Marshal(body)
		request := httptest.NewRequest(method, "/api/v1/auth/"+path, strings.NewReader(string(encoded)))
		request.Header.Set("Origin", origin)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-CSRF-Token", csrf)
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}
	origin := "http://localhost:8080"
	expect := func(response *httptest.ResponseRecorder, status int) {
		testingTarget.Helper()
		if response.Code != status {
			testingTarget.Fatalf("status %d expected %d; response %s", response.Code, status, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "password_hash") || strings.Contains(response.Body.String(), "csrf_hash") {
			testingTarget.Fatal("credential hash leaked")
		}
	}
	credentials := map[string]string{"email": "Reader@Example.com", "password": "correct-password-123", "display_name": "Оқырман"}
	expect(send("POST", "register", credentials, nil, "", "http://foreign.example"), 403)
	expect(send("POST", "register", credentials, nil, "", origin), 202)
	expect(send("POST", "register", credentials, nil, "", origin), 202)
	verification := mailer.token()
	expect(send("POST", "verify-email", map[string]string{"token": verification}, nil, "", origin), 200)
	expect(send("POST", "verify-email", map[string]string{"token": verification}, nil, "", origin), 410)
	expect(send("POST", "login", map[string]string{"email": "reader@example.com", "password": "wrong"}, nil, "", origin), 401)
	login := send("POST", "login", credentials, nil, "", origin)
	expect(login, 200)
	cookies := login.Result().Cookies()
	for _, cookie := range cookies {
		if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
			testingTarget.Fatal("unsafe session cookie")
		}
	}
	me := send("GET", "me", nil, cookies, "", origin)
	expect(me, 200)
	if !strings.Contains(me.Body.String(), "\"email_verified\":true") {
		testingTarget.Fatal("email not verified")
	}
	expect(send("POST", "logout", nil, cookies, "", origin), 403)
	csrfResponse := send("GET", "csrf", nil, cookies, "", origin)
	expect(csrfResponse, 200)
	var csrfPayload struct {
		Token string `json:"csrf_token"`
	}
	if failure = json.Unmarshal(csrfResponse.Body.Bytes(), &csrfPayload); failure != nil {
		testingTarget.Fatal(failure)
	}

	var identityPayload struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if failure = json.Unmarshal(me.Body.Bytes(), &identityPayload); failure != nil {
		testingTarget.Fatal(failure)
	}
	profile := map[string]string{"account_id": identityPayload.User.ID, "display_name": "Қазақ оқырманы"}
	expect(send("POST", "profile", profile, cookies, "", origin), 403)
	expect(send("POST", "profile", profile, cookies, csrfPayload.Token, origin), 200)
	me = send("GET", "me", nil, cookies, "", origin)
	if !strings.Contains(me.Body.String(), "Қазақ оқырманы") {
		testingTarget.Fatal("profile name not persisted")
	}
	profile["account_id"] = "another-account"
	expect(send("POST", "profile", profile, cookies, csrfPayload.Token, origin), 409)
	expect(send("POST", "logout", nil, cookies, csrfPayload.Token, "http://foreign.example"), 403)
	expect(send("POST", "forgot-password", map[string]string{"email": "unknown@example.com"}, nil, "", origin), 202)
	expect(send("POST", "forgot-password", map[string]string{"email": "reader@example.com"}, nil, "", origin), 202)
	resetToken := mailer.token()
	expect(send("POST", "reset-password", map[string]string{"token": resetToken, "password": "changed-password-456"}, nil, "", origin), 200)
	expect(send("POST", "reset-password", map[string]string{"token": resetToken, "password": "changed-password-456"}, nil, "", origin), 410)
	expect(send("GET", "me", nil, cookies, "", origin), 401)
	expect(send("POST", "login", credentials, nil, "", origin), 401)
	changed := map[string]string{"email": "reader@example.com", "password": "changed-password-456"}
	login = send("POST", "login", changed, nil, "", origin)
	expect(login, 200)
	cookies = login.Result().Cookies()
	expect(send("GET", "me", nil, cookies, "", origin), 200)
	if _, failure = database.Pool().Exec(ctx, "UPDATE sessions SET expires_at=now()-interval '1 second'"); failure != nil {
		testingTarget.Fatal(failure)
	}
	expect(send("GET", "me", nil, cookies, "", origin), 401)
}
