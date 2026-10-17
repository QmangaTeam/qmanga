package http

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func requestOffset(request *gin.Context) int {
	value, _ := strconv.Atoi(request.Query("offset"))
	return max(0, min(value, 1000000))
}
