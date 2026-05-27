package middleware

import (
	"api/logger"
	"api/metrics"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID gera um UUID único por requisição e loga quando completa.
// Substitui tanto request_id.go quanto o logger.go anterior.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		args := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		}

		if userID, exists := c.Get("user_id"); exists {
			args = append(args, "user_id", userID)
		}

		logger.Log.Log(c.Request.Context(), level, "request completed", args...)

		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(status),
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration.Seconds())
	}
}

// RequestLogger é um alias para compatibilidade com main.go
func RequestLogger() gin.HandlerFunc {
	return RequestID()
}
