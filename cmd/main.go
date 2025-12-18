package main

import (
	"api/controller"
	"api/db"
	"api/repository"
	"api/usecase"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	server := gin.Default()

	dbConnection, err := db.ConnectDB()
	if err != nil {
		panic(err)
	}

	//Camada de repository
	ProductRepository := repository.NewProductRepository(dbConnection)
	//Camada usecase
	ProductUseCase := usecase.NewProductUseCase(ProductRepository)
	//Camada de controllers
	ProductController := controller.NewProductController(ProductUseCase)

	//mensagem teste para saber se o server is up
	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//rotas crud
	server.GET("/products", ProductController.GetProducts)
	server.POST("/products", ProductController.CreateProduct)
	server.GET("/products/:productId", ProductController.GetProductById)
	server.PUT("/products/:productId", ProductController.UpdateProduct)
	server.DELETE("/products/:productId", ProductController.Delete)
	log.Println("API Volurya rodando na porta 8080")

	server.Run(":8000")

}
