package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			slog.Warn("JWT validation failed", "error", err, "path", c.FullPath())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// extractToken tenta extrair o JWT de múltiplas fontes na ordem de prioridade:
// 1. Cookie HttpOnly (mais seguro — não acessível por JavaScript)
// 2. Header Authorization: Bearer (para clientes API/mobile)
func extractToken(c *gin.Context) string {
	// 1. Cookie HttpOnly — fonte principal após migração
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}

	// 2. Header Authorization — mantido para compatibilidade com clientes API
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

// RequireAdminRole verifica explicitamente tanto autenticação quanto role admin
func RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verifica se o middleware de autenticação rodou
		_, authenticated := c.Get("user_id")
		if !authenticated {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden: admin role required",
			})
			return
		}

		c.Next()
	}
}

// GetUserID extrai o user_id do contexto de forma tipada e segura
func GetUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := val.(int)
	return id, ok
}
