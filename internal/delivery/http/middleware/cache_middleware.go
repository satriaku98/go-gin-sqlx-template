package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func CacheMiddleware(redisClient *database.RedisClient, ttl time.Duration, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		key := GetCacheKey(c)
		ctx := context.Background()

		// Check cache
		val, err := redisClient.Client.Get(ctx, key).Result()
		if err == nil {
			c.Header("Content-Type", "application/json")
			c.Header("X-Cache", "HIT")
			c.String(http.StatusOK, val)
			c.Abort()
			return
		}

		// Cache miss
		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		// Save to cache if status is 200
		if c.Writer.Status() == http.StatusOK {
			if err := redisClient.Client.Set(ctx, key, w.body.String(), ttl).Err(); err != nil {
				logger.Errorf(c.Request.Context(), "failed to cache response: %v", err)
			}
		}
	}
}

func GetCacheKey(c *gin.Context) string {
	return fmt.Sprintf("cache:%s", c.Request.URL.RequestURI())
}
