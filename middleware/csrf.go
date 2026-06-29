package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// CSRFTokenProvider gera um token CSRF aleatório e seta como cookie legível por JS.
// Deve ser aplicado antes do CSRFProtection.
func CSRFTokenProvider() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verifica se já existe cookie CSRF válido
		if token, err := c.Cookie(csrfCookieName); err == nil && len(token) == 64 {
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// Gera novo token aleatório
		token, err := generateCSRFToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		secure := os.Getenv("APP_ENV") == "production"

		// Cookie NÃO é HttpOnly — JavaScript precisa ler para enviar no header
		c.SetCookie(
			csrfCookieName,
			token,
			3600,  // 1 hora
			"/",
			"",
			secure,
			false, // não HttpOnly
		)

		c.Set("csrf_token", token)
		c.Next()
	}
}

// CSRFProtection valida o token CSRF via Double Submit Cookie Pattern.
// Rotas públicas de auth (login, signup) são isentas — usuário ainda não tem cookie.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Métodos seguros não precisam de CSRF
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Rotas isentas — usuário ainda não tem cookie CSRF
		exemptPaths := map[string]bool{
			"/api/login":   true,
			"/api/signup":  true,
			"/api/refresh": true,
			"/api/webhook": true, // webhook do Stripe tem validação própria
		}
		if exemptPaths[c.FullPath()] {
			c.Next()
			return
		}

		// Lê token do cookie
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token missing"})
			c.Abort()
			return
		}

		// Lê token do header enviado pelo frontend
		headerToken := c.GetHeader(csrfHeaderName)
		if headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token missing"})
			c.Abort()
			return
		}

		// Compara os dois tokens — devem ser idênticos
		if cookieToken != headerToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateCSRFToken mantido para compatibilidade
func GenerateCSRFToken(sessionID string) string {
	token, _ := generateCSRFToken()
	return token
}
