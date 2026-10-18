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

func TestTrustedGatewayForwardsClientIPButDirectPeerCannotSpoof(testingTarget *testing.T) {
	router := NewRouter(nil)
	if err := TrustProxyHost(router, "localhost"); err != nil {
		testingTarget.Fatal(err)
	}
	router.GET("/client-ip", func(request *gin.Context) { request.String(200, request.ClientIP()) })
	for _, example := range []struct{ peer, want string }{
		{"127.0.0.1:1234", "203.0.113.8"},
		{"198.51.100.5:1234", "198.51.100.5"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/client-ip", nil)
		request.RemoteAddr = example.peer
		request.Header.Set("X-Forwarded-For", "203.0.113.8")
		router.ServeHTTP(recorder, request)
		if recorder.Body.String() != example.want {
			testingTarget.Fatalf("peer %s was attributed to %q, want %q", example.peer, recorder.Body.String(), example.want)
		}
	}
}
