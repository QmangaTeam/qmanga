package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"qmanga/internal/identity/application"
	"strings"
	"testing"
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
