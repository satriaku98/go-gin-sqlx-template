package middleware

import (
	"time"

	"go-gin-sqlx-template/pkg/logger"

	"github.com/gin-gonic/gin"
)

func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)

		// Create logger with request-specific fields
		requestLogger := log.WithFields(c.Request.Context(), map[string]any{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"latency":     duration.String(),
			"client_ip":   c.ClientIP(),
		})

		// Log with simple message
		requestLogger.Info(c.Request.Context(), "HTTP request completed")
	}
}

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf(c.Request.Context(), "Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"success": false,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Placeholder for authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT or session-based authentication
		c.Next()
	}
}
