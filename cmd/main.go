package main

import (
	"api/auth"
	"api/controller"
	"api/db"
	"api/repository"
	"api/usecase"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()

	server.LoadHTMLGlob("views/templates/*")
	server.Static("/static", "./views/static")

	server.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	// Login
	server.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{})
	})

	// Cadastro
	server.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", gin.H{})
	})

	// Loja pública (visualização de produtos)
	server.GET("/store", func(c *gin.Context) {
		c.HTML(200, "store.html", gin.H{})
	})

	// Painel admin (lista e gerenciamento de produtos)
	server.GET("/admin", func(c *gin.Context) {
		// Aqui você pode adicionar middleware de autenticação no futuro
		// Por enquanto só renderiza
		c.HTML(200, "admin.html", gin.H{}) // ou "products.html", dependendo do nome que você deu
	})

	dbConnection, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.EnsureTablesExist(dbConnection); err != nil {
		log.Fatalf("failed to ensure tables exist: %v", err)
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
	public := server.Group("/api")
	{
		public.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"message": "pong"})
		})
		public.POST("/signup", authController.Signup)
		public.POST("/login", authController.Login)
	}
	// ---------- ROTAS PROTEGIDAS ----------
	protected := server.Group("/api")
	protected.Use(auth.Middleware())
	{
		protected.GET("/products", productController.GetProducts)
		protected.POST("/products", productController.CreateProduct)
		protected.GET("/products/:productId", productController.GetProductById)
		protected.PUT("/products/:productId", productController.UpdateProduct)
		protected.DELETE("/products/:productId", productController.Delete)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("API Volurya rodando na porta", port)

	if err := server.Run(":" + port); err != nil {
		log.Fatal(err)
	}

}
