package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(responseWriter http.ResponseWriter, statusCode int, payload interface{}) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(statusCode)
	_ = json.NewEncoder(responseWriter).Encode(payload)
}

func WriteKazakhError(responseWriter http.ResponseWriter, statusCode int, errorCode string, errorMessage string) {
	responsePayload := ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: errorMessage,
		},
	}
	WriteJSON(responseWriter, statusCode, responsePayload)
}

func NewMediaHandler() http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			WriteKazakhError(responseWriter, http.StatusMethodNotAllowed, "method_not_allowed", "Әдіске рұқсат етілмеген")
			return
		}
		WriteKazakhError(responseWriter, http.StatusNotFound, "not_found", "Медиа файлы табылмады")
	})
}

func NewRouter(healthHandler http.Handler, mediaHandler http.Handler) http.Handler {
	routerMux := http.NewServeMux()

	if healthHandler != nil {
		routerMux.Handle("/api/v1/health", healthHandler)
	}

	if mediaHandler != nil {
		routerMux.Handle("/media/", mediaHandler)
	}

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestPath := request.URL.Path
		if strings.HasPrefix(requestPath, "/api/") {
			if requestPath == "/api/v1/health" {
				routerMux.ServeHTTP(responseWriter, request)
				return
			}
			WriteKazakhError(responseWriter, http.StatusNotFound, "not_found", "Табылмады")
			return
		}

		routerMux.ServeHTTP(responseWriter, request)
	})
}
