package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qmanga/internal/config"
	"qmanga/internal/storage"
	"qmanga/internal/store"
)

type API struct {
	cfg     config.Config
	store   *store.Store
	storage *storage.Client
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Postgres string `json:"postgres"`
	S3       string `json:"s3"`
}

func New(cfg config.Config, st *store.Store, stor *storage.Client) http.Handler {
	api := &API{
		cfg:     cfg,
		store:   st,
		storage: stor,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", api.handleHealth)
	mux.HandleFunc("/media/", api.handleMedia)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Custom handler wrapper to catch 404s with Kazakh JSON error responses for /api/
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if r.URL.Path == "/api/v1/health" {
				mux.ServeHTTP(w, r)
				return
			}
			writeKazakhError(w, http.StatusNotFound, "not_found", "Табылмады")
			return
		}

		mux.ServeHTTP(w, r)
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeKazakhError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Әдіске рұқсат етілмеген")
		return
	}

	ctx := r.Context()
	pgStatus := "ok"
	if err := a.store.Ping(ctx); err != nil {
		pgStatus = "error: " + err.Error()
	}

	s3Status := "ok"
	if err := a.storage.Ping(ctx); err != nil {
		s3Status = "error: " + err.Error()
	}

	isHealthy := pgStatus == "ok" && s3Status == "ok"
	statusCode := http.StatusOK
	statusText := "ok"
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
		statusText = "degraded"
	}

	resp := HealthResponse{
		Status:   statusText,
		Version:  "0.1.0",
		Postgres: pgStatus,
		S3:       s3Status,
	}

	writeJSON(w, statusCode, resp)
}

func (a *API) handleMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeKazakhError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Әдіске рұқсат етілмеген")
		return
	}

	// Task 1 placeholder for media serving
	writeKazakhError(w, http.StatusNotFound, "not_found", "Медиа файлы табылмады")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeKazakhError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	writeJSON(w, status, resp)
}
