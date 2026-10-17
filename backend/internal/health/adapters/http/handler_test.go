package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"qmanga/internal/health/application"
	"qmanga/internal/platform/httpserver"
)

type dummyPinger struct {
	err error
}

func (pinger *dummyPinger) Ping(contextInstance context.Context) error {
	return pinger.err
}

func TestHealthHandler_ReturnsOKWhenHealthy(testingTarget *testing.T) {
	postgresPinger := &dummyPinger{err: nil}
	s3Pinger := &dummyPinger{err: nil}
	useCase := application.NewCheckHealthUseCase("0.1.0", postgresPinger, s3Pinger)
	healthHandler := NewHandler(useCase)

	httpRequest := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()

	healthHandler.ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusOK {
		testingTarget.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var responsePayload ResponseDTO
	if err := json.Unmarshal(recorder.Body.Bytes(), &responsePayload); err != nil {
		testingTarget.Fatalf("failed to decode response JSON: %v", err)
	}

	if responsePayload.Status != "ok" {
		testingTarget.Errorf("expected status ok, got %s", responsePayload.Status)
	}
	if responsePayload.Postgres != "ok" {
		testingTarget.Errorf("expected postgres ok, got %s", responsePayload.Postgres)
	}
	if responsePayload.S3 != "ok" {
		testingTarget.Errorf("expected s3 ok, got %s", responsePayload.S3)
	}
}

func TestHealthHandler_Returns503WhenDegraded(testingTarget *testing.T) {
	postgresPinger := &dummyPinger{err: errors.New("connection failed")}
	s3Pinger := &dummyPinger{err: nil}
	useCase := application.NewCheckHealthUseCase("0.1.0", postgresPinger, s3Pinger)
	healthHandler := NewHandler(useCase)

	httpRequest := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()

	healthHandler.ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusServiceUnavailable {
		testingTarget.Fatalf("expected status 503, got %d", recorder.Code)
	}

	var responsePayload ResponseDTO
	if err := json.Unmarshal(recorder.Body.Bytes(), &responsePayload); err != nil {
		testingTarget.Fatalf("failed to decode response JSON: %v", err)
	}

	if responsePayload.Status != "degraded" {
		testingTarget.Errorf("expected status degraded, got %s", responsePayload.Status)
	}
	if responsePayload.Postgres != "unavailable" {
		testingTarget.Errorf("expected postgres unavailable, got %s", responsePayload.Postgres)
	}
}

func TestHealthHandler_RejectsNonGetMethod(testingTarget *testing.T) {
	postgresPinger := &dummyPinger{err: nil}
	s3Pinger := &dummyPinger{err: nil}
	useCase := application.NewCheckHealthUseCase("0.1.0", postgresPinger, s3Pinger)
	healthHandler := NewHandler(useCase)

	httpRequest := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()

	healthHandler.ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusMethodNotAllowed {
		testingTarget.Fatalf("expected status 405, got %d", recorder.Code)
	}
}

func TestRouter_Integration(testingTarget *testing.T) {
	postgresPinger := &dummyPinger{err: nil}
	s3Pinger := &dummyPinger{err: nil}
	useCase := application.NewCheckHealthUseCase("0.1.0", postgresPinger, s3Pinger)
	healthHandler := NewHandler(useCase)
	mediaHandler := httpserver.NewMediaHandler()
	router := httpserver.NewRouter(healthHandler, mediaHandler)

	healthRequest := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	healthRecorder := httptest.NewRecorder()
	router.ServeHTTP(healthRecorder, healthRequest)
	if healthRecorder.Code != http.StatusOK {
		testingTarget.Errorf("expected 200 for /api/v1/health, got %d", healthRecorder.Code)
	}

	notFoundRequest := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
	notFoundRecorder := httptest.NewRecorder()
	router.ServeHTTP(notFoundRecorder, notFoundRequest)
	if notFoundRecorder.Code != http.StatusNotFound {
		testingTarget.Errorf("expected 404 for /api/v1/unknown, got %d", notFoundRecorder.Code)
	}

	mediaRequest := httptest.NewRequest(http.MethodGet, "/media/sample.jpg", nil)
	mediaRecorder := httptest.NewRecorder()
	router.ServeHTTP(mediaRecorder, mediaRequest)
	if mediaRecorder.Code != http.StatusNotFound {
		testingTarget.Errorf("expected 404 for /media/sample.jpg, got %d", mediaRecorder.Code)
	}
	if mediaRecorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		testingTarget.Errorf("expected X-Content-Type-Options nosniff, got %s", mediaRecorder.Header().Get("X-Content-Type-Options"))
	}
}
