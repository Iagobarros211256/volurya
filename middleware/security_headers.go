package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adiciona headers de segurança HTTP em todas as respostas.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Impede MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Impede clickjacking via iframes
		c.Header("X-Frame-Options", "DENY")

		// Desativa proteção XSS antiga do browser — conflita com CSP moderno
		c.Header("X-XSS-Protection", "0")

		// Limita informações enviadas no header Referer
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Desativa features de browser não utilizadas
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		c.Next()
	}
}
