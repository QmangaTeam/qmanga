package http

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"qmanga/internal/platform/readiness"
)

// TrustProxyHost resolves the one internal gateway instead of trusting arbitrary
// client-supplied forwarding headers or an entire Docker address range.
func TrustProxyHost(router *gin.Engine, host string) error {
	addresses, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolving trusted gateway %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("resolving trusted gateway %q: no IP addresses", host)
	}
	trusted := make([]string, 0, len(addresses))
	for _, address := range addresses {
		trusted = append(trusted, address.String())
	}
	return router.SetTrustedProxies(trusted)
}

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
