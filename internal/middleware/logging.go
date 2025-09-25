package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Logger provides structured logging middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID := "unknown"
		if id, exists := param.Keys["request_id"]; exists {
			requestID = fmt.Sprintf("%v", id)
		}

		return fmt.Sprintf("[%s] %s %s %s %d %s \"%s\" %s \"%s\" request_id=%v\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			param.Request.Referer(),
			requestID,
		)
	})
}