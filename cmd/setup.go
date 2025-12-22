package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// health não precisa do banco
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// aqui futuramente entram rotas que usam db
	// ex:
	// userRepo := repository.NewUserRepository(db)
	// userController := controller.NewUserController(userRepo)
	// r.POST("/users", userController.Create)

	return r
}
