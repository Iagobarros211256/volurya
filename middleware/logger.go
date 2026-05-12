package middleware

import (
	"api/logger"
	"api/metrics"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := fmt.Sprintf("%08x", rand.Uint32())
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		userID, _ := c.Get("user_id")

		logger.Log.Log(
			c.Request.Context(),
			level,
			"request completed",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
			"user_id", userID,
		)

		// Métricas Prometheus
		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(status),
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration.Seconds())
	}
}
