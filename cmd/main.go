package main

import (
	"api/auth"
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

	// ---------- REPOSITORIES ----------
	productRepository := repository.NewProductRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)

	// ---------- USECASES ----------
	productUseCase := usecase.NewProductUsecase(productRepository)
	authUseCase := usecase.NewAuthUsecase(userRepository)

	// ---------- CONTROLLERS ----------
	productController := controller.NewProductController(productUseCase)
	authController := controller.NewAuthController(authUseCase)

	// ---------- ROTAS PÚBLICAS ----------
	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})
	server.POST("/signup", authController.Signup)
	server.POST("/login", authController.Login)

	// ---------- ROTAS PROTEGIDAS ----------
	protected := server.Group("/")
	protected.Use(auth.Middleware())
	{
		protected.GET("/products", productController.GetProducts)
		protected.POST("/products", productController.CreateProduct)
		protected.GET("/products/:productId", productController.GetProductById)
		protected.PUT("/products/:productId", productController.UpdateProduct)
		protected.DELETE("/products/:productId", productController.Delete)
	}

	log.Println("API Volurya rodando na porta 8000")
	database := db.SetupDB() // DB de dev/prod
	r := SetupRouter(database)
	r.Run()
	server.Run(":8000")
}
