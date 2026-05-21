package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			// aqui você pode logar err
			//Esse comentário indica que o log nunca foi implementado.Provavelmente esqueci e deixei por isso mesmo.
			//  Com o slog já disponível no projeto, seria trivial:
			//goslog.Warn("JWT validation failed", "error", err, "path", c.FullPath())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}
		//c.Set armazena interface{} — risco de type assertion em runtime
		//Se por qualquer motivo o middleware não rodou (rota mal configurada, por exemplo),
		// isso vai causar panic. O padrão mais defensivo é usar helpers tipados.
		//nao entendo bem dessa parte ainda. marcar como estudar tbm e analisar trade offs.
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	if c.Request.Method == http.MethodGet && c.FullPath() == "/api/events" {
		return c.Query("access_token")
		//Tokens em query string aparecem em logs de servidor, histórico do browser,
		// headers Referer e ferramentas de analytics. Para SSE/WebSocket o padrão mais seguro é enviar o token no primeiro frame/mensagem, ou usar um ticket de curta duração (token OTP de uso único, válido por ~30s, trocado por um token real no servidor).
		// Se precisar manter por ora, pelo menos documente o risco com um comentário // TODO: security risk.
		//nao entendo bem quale o trade off nessa parte. vou pesquisar sobre e tomar alguma decisao
	}

	return ""
}

// RequireAdminRole depende de Middleware() ter rodado antes
// Não há verificação explícita dessa dependência. Se RequireAdminRole for registrado em uma rota sem Middleware(),
//
//	vai retornar forbidden para todos silenciosamente em vez de unauthorized.
//
// Considere verificar a ausência de role separadamente
func RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
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
