package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"qmanga/internal/platform/readiness"
)

func ready(checker *readiness.CheckHealthUseCase) gin.HandlerFunc {
	return func(request *gin.Context) {
		if checker == nil {
			Error(request, http.StatusServiceUnavailable, "unavailable", "Қызмет уақытша қолжетімсіз")
			return
		}
		report, healthy := checker.Execute(request.Request.Context())
		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}
		if request.Request.Method == http.MethodHead {
			request.Status(status)
			return
		}
		request.JSON(status, gin.H{"status": report.Status, "version": report.Version, "postgres": report.Postgres, "s3": report.S3})
	}
}
