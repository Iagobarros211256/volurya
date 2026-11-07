package config

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CartPageGuard redireciona visitantes para o catálogo
// quando o carrinho estiver indisponível.
//
// A função não autoriza pagamentos; essa proteção
// permanece nos endpoints do backend.
func CartPageGuard(c *gin.Context) {
	if !CartAvailable() {
		c.Redirect(http.StatusFound, "/store")
		c.Abort()
		return
	}

	c.Next()
}
