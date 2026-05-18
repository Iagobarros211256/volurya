package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CSRFToken generates and validates CSRF tokens
// For simplicity, we use a basic token based on session + secret
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow safe methods without CSRF check
		if c.Request.Method == http.MethodGet || 
		   c.Request.Method == http.MethodHead || 
		   c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Only check CSRF for form submissions (not JSON APIs with Authorization header)
		// JSON APIs should use Authorization header which is CSRF-safe
		if c.ContentType() == "application/json" {
			c.Next()
			return
		}

		// For form submissions, verify CSRF token
		token := c.PostForm("_csrf_token")
		if token == "" {
			token = c.GetHeader("X-CSRF-Token")
		}

		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Validate token (basic validation - in production use session-based tokens)
		if !isValidCSRFToken(token) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token invalid",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a CSRF token
func GenerateCSRFToken(sessionID string) string {
	data := sessionID + "csrf_secret_key"
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// isValidCSRFToken validates a CSRF token
func isValidCSRFToken(token string) bool {
	// Basic validation: check if token is hex string of correct length
	if len(token) != 64 { // SHA256 hex length
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}

// CSRFTokenMiddleware adds CSRF token to context for templates
func CSRFTokenProvider() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a token based on request
		token := GenerateCSRFToken(c.ClientIP())
		c.Set("csrf_token", token)
		c.Next()
	}
}
