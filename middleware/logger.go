package middleware

import (
	"api/logger"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Gera request ID único
		requestID := fmt.Sprintf("%08x", rand.Uint32())
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Processa a requisição
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Nível do log baseado no status
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		// Pega user_id se autenticado
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
	}
}
