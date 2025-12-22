package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter cria o router da API com todas as rotas
func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// rota de health check (não precisa do banco)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// rota POST /users - grava no banco
	r.POST("/users", func(c *gin.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		// valida payload
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// insere no banco
		_, err := db.Exec(
			"INSERT INTO users (name, email) VALUES ($1, $2)",
			user.Name,
			user.Email,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		// retorna sucesso
		c.JSON(http.StatusCreated, gin.H{"message": "user created"})
	})

	return r
}
