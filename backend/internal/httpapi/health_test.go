package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"qmanga/internal/config"
	"qmanga/internal/storage"
	"qmanga/internal/store"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	cfg := config.Config{
		Environment: "test",
	}
	st := &store.Store{}
	stor := &storage.Client{}

	handler := New(cfg, st, stor)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["postgres"] != "ok" {
		t.Errorf("expected postgres 'ok', got %v", resp["postgres"])
	}
	if resp["s3"] != "ok" {
		t.Errorf("expected s3 'ok', got %v", resp["s3"])
	}
}

func TestNotFound_ReturnsKazakhError(t *testing.T) {
	cfg := config.Config{
		Environment: "test",
	}
	handler := New(cfg, &store.Store{}, &storage.Client{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	type errResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	var parsed errResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}

	if parsed.Error.Code != "not_found" {
		t.Errorf("expected code 'not_found', got %q", parsed.Error.Code)
	}
	if parsed.Error.Message != "Табылмады" {
		t.Errorf("expected Kazakh message 'Табылмады', got %q", parsed.Error.Message)
	}
}
