package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	maxRequests int
	window      time.Duration
	clients     map[string]*clientLimiter
	mu          sync.Mutex
}

type clientLimiter struct {
	requests  []time.Time
	mu        sync.Mutex
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientLimiter),
	}

	go limiter.cleanup()
	return limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, client := range rl.clients {
			client.mu.Lock()
			now := time.Now()
			validRequests := 0
			for _, reqTime := range client.requests {
				if now.Sub(reqTime) < rl.window {
					validRequests++
				}
			}
			if validRequests == 0 {
				delete(rl.clients, ip)
			} else {
				client.requests = client.requests[len(client.requests)-validRequests:]
			}
			client.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rl.mu.Lock()
		client, exists := rl.clients[clientIP]
		if !exists {
			client = &clientLimiter{
				requests: make([]time.Time, 0),
			}
			rl.clients[clientIP] = client
		}
		rl.mu.Unlock()

		client.mu.Lock()
		now := time.Now()
		validRequests := 0
		for _, reqTime := range client.requests {
			if now.Sub(reqTime) < rl.window {
				validRequests++
			}
		}

		if validRequests >= rl.maxRequests {
			client.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		client.requests = append(client.requests, now)
		client.mu.Unlock()

		c.Next()
	}
}
