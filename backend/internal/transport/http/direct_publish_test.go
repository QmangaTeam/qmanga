package http

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	identity "qmanga/internal/identity/application"
)

func TestDirectPublishRejectsSessionAndUnprivilegedToken(t *testing.T) {
	for _, item := range []struct {
		name  string
		token *identity.APIToken
	}{
		{name: "session"},
		{name: "ordinary bearer", token: &identity.APIToken{ID: "token"}},
	} {
		t.Run(item.name, func(t *testing.T) {
			router := gin.New()
			handler := NewPublishingHandler(nil)
			router.POST("/publish", func(c *gin.Context) {
				if item.token != nil {
					c.Set("api_token", *item.token)
				}
				handler.directPublish(c)
			})
			response := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/publish", strings.NewReader(`{"version":1}`))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(response, request)
			if response.Code != 403 {
				t.Fatalf("status=%d", response.Code)
			}
		})
	}
}

func TestDirectCoverRejectsSessionAndOrdinaryToken(t *testing.T) {
	for _, token := range []*identity.APIToken{nil, {ID: "ordinary"}} {
		router := gin.New()
		handler := NewPublishingHandler(nil)
		router.PUT("/cover", func(c *gin.Context) {
			if token != nil {
				c.Set("api_token", *token)
			}
			handler.selectCover(c)
		})
		response := httptest.NewRecorder()
		request := httptest.NewRequest("PUT", "/cover", strings.NewReader(`{"page_id":"page"}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != 409 || !strings.Contains(response.Body.String(), "moderation_required") {
			t.Fatalf("got %d %s", response.Code, response.Body.String())
		}
	}
}
