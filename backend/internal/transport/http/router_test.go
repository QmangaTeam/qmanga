package http

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"qmanga/internal/platform/readiness"
)

type dependency struct{ failure error }

func (check dependency) Ping(context.Context) error { return check.failure }

func TestGinTransportAndDependencyFailures(testingTarget *testing.T) {
	checker := readiness.NewCheckHealthUseCase("test", dependency{errors.New("private database password")}, dependency{})
	var router *gin.Engine = NewRouter(checker)
	for _, example := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/api/v1/health", 503}, {"HEAD", "/api/v1/health", 503},
		{"GET", "/api/v1/live", 200}, {"POST", "/api/v1/health", 405},
		{"GET", "/api/v1/unknown", 404}, {"GET", "/media/unknown", 404},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(example.method, example.path, nil))
		if recorder.Code != example.status {
			testingTarget.Errorf("%s %s: %d, expected %d", example.method, example.path, recorder.Code, example.status)
		}
		if strings.Contains(recorder.Body.String(), "password") {
			testingTarget.Fatal("dependency details leaked")
		}
		if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
			testingTarget.Fatal("missing security headers")
		}
	}
}
