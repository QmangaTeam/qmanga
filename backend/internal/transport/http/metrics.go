package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"qmanga/internal/metrics/application"
	"strconv"
)

type MetricsHandler struct{ service *application.Service }

func NewMetricsHandler(service *application.Service) *MetricsHandler { return &MetricsHandler{service} }
func (handler *MetricsHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	api.GET("/admin/metrics", identity.RequireUser(), RequireRole("admin"), handler.report)
	api.GET("/studio/metrics", identity.RequireUser(), handler.report)
}
func (handler *MetricsHandler) report(request *gin.Context) {
	request.Header("Cache-Control", "private, no-store")
	days := 30
	var err error
	if raw := request.Query("days"); raw != "" {
		days, err = strconv.Atoi(raw)
	}
	if err != nil {
		Error(request, 400, "validation", "Invalid period")
		return
	}
	actor, _ := CurrentAccount(request)
	report, err := handler.service.Report(request.Request.Context(), actor.ID, request.FullPath() == "/api/v1/admin/metrics", days)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalid):
			Error(request, 400, "validation", "Invalid period")
		case errors.Is(err, application.ErrForbidden):
			Error(request, 403, "forbidden", "Access denied")
		default:
			Error(request, 503, "unavailable", "Metrics unavailable")
		}
		return
	}
	items := []gin.H{}
	for _, day := range report.Days {
		items = append(items, gin.H{"date": day.Date, "verified": day.Verified, "publications": day.Publications, "readers": day.Readers, "completions": day.Completions, "ratings": day.Ratings, "libraries": day.Libraries, "subscriptions": day.Subscriptions})
	}
	request.JSON(200, gin.H{"days": items, "started_at": report.StartedAt, "period_readers": report.PeriodReaders})
}
