package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"qmanga/internal/platform/readiness"
)

func NewRouter(checker *readiness.CheckHealthUseCase) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.HandleMethodNotAllowed = true
	router.Use(SecurityHeaders(), Recover())
	api := router.Group("/api/v1")
	api.GET("/health", ready(checker))
	api.HEAD("/health", ready(checker))
	api.GET("/live", func(request *gin.Context) { request.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.NoRoute(func(request *gin.Context) { Error(request, http.StatusNotFound, "not_found", "Табылмады") })
	router.NoMethod(func(request *gin.Context) {
		Error(request, http.StatusMethodNotAllowed, "method_not_allowed", "Әдіске рұқсат етілмеген")
	})
	return router
}
