package http

import (
	"net/http"

	"qmanga/internal/health/application"
	"qmanga/internal/platform/httpserver"
)

type ResponseDTO struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Postgres string `json:"postgres"`
	S3       string `json:"s3"`
}

type Handler struct {
	useCase *application.CheckHealthUseCase
}

func NewHandler(useCase *application.CheckHealthUseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (handlerInstance *Handler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		httpserver.WriteKazakhError(responseWriter, http.StatusMethodNotAllowed, "method_not_allowed", "Әдіске рұқсат етілмеген")
		return
	}

	report, isHealthy := handlerInstance.useCase.Execute(request.Context())

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	responsePayload := ResponseDTO{
		Status:   report.Status,
		Version:  report.Version,
		Postgres: report.Postgres,
		S3:       report.S3,
	}

	httpserver.WriteJSON(responseWriter, statusCode, responsePayload)
}
