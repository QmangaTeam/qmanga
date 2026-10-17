package http

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(request *gin.Context) {
		request.Header("X-Content-Type-Options", "nosniff")
		request.Header("Cache-Control", "no-store")
		request.Header("X-Request-ID", uuid.NewString())
		request.Next()
	}
}

func Recover() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(request *gin.Context, failure any) {
		log.Printf("request failed: route=%s", request.FullPath())
		Error(request, http.StatusInternalServerError, "internal", "Сервер қатесі. Кейінірек қайталап көріңіз.")
	})
}

func Error(request *gin.Context, status int, code, message string) {
	request.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
