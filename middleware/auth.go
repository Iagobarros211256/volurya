package middleware

import (
	"api/auth"
	"log"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "token missing"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			log.Printf("JWT validation failed: %v - token: %s", err, tokenString)
			c.JSON(401, gin.H{"error": "invalid token - " + err.Error()})
			c.Abort()
			return
		}

		// salvar claims no contexto
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
