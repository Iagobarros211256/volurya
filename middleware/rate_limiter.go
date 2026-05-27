package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const maxClientsLimit = 10000

type clientLimiter struct {
	requests []time.Time
	mu       sync.Mutex
}

type RateLimiter struct {
	maxRequests int
	window      time.Duration
	clients     sync.Map // substitui map + mutex global
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Proteção contra DoS por memória
		count := 0
		rl.clients.Range(func(_, _ any) bool {
			count++
			return count < maxClientsLimit
		})
		if count >= maxClientsLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "server busy"})
			c.Abort()
			return
		}

		// Carrega ou cria o clientLimiter para este IP
		val, _ := rl.clients.LoadOrStore(clientIP, &clientLimiter{
			requests: make([]time.Time, 0),
		})
		client := val.(*clientLimiter)

		client.mu.Lock()
		defer client.mu.Unlock()

		now := time.Now()

		// Filtra requests dentro da janela
		valid := client.requests[:0]
		for _, t := range client.requests {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		client.requests = valid

		if len(client.requests) >= rl.maxRequests {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		client.requests = append(client.requests, now)
		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.clients.Range(func(key, val any) bool {
			client := val.(*clientLimiter)

			client.mu.Lock()
			valid := client.requests[:0]
			for _, t := range client.requests {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			client.requests = valid
			empty := len(client.requests) == 0
			client.mu.Unlock()

			if empty {
				rl.clients.Delete(key)
			}
			return true
		})
	}
}
