package main

import (
	"api/auth"
	"api/controller"
	"api/db"
	"api/repository"
	"api/usecase"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.LoadHTMLGlob("views/templates/*")
	router.Static("/static", "./views/static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	router.GET("/store", func(c *gin.Context) {
		c.HTML(200, "store.html", gin.H{
			"title": "VOLURYA SHOP - Produtos Oficiais",
		})
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{})
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", gin.H{})
	})

	// Ping de saúde
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	// DB
	dbConnection, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Migrations
	if err := db.RunMigrations(dbConnection); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Repositories
	productRepository := repository.NewProductRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	orderRepository := repository.NewOrderRepository(dbConnection)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dbConnection)

	// Usecases
	productUseCase := usecase.NewProductUsecase(productRepository)
	authUseCase := usecase.NewAuthUsecase(userRepository, refreshTokenRepository)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)

	// Controllers
	productController := controller.NewProductController(productUseCase)
	authController := controller.NewAuthController(authUseCase)
	orderController := controller.NewOrderController(orderUsecase)

	// Rotas públicas
	public := router.Group("/api")
	{
		public.POST("/signup", authController.Signup)
		public.POST("/login", authController.Login)
		public.POST("/refresh", authController.RefreshToken)
		public.POST("/logout", authController.Logout)
	}

	// Rotas protegidas
	protected := router.Group("/api")
	protected.Use(auth.Middleware())
	{
		protected.GET("/products", productController.GetProducts)
		protected.POST("/products", productController.CreateProduct)
		protected.POST("/orders", orderController.CreateOrder)
		protected.GET("/products/:productId", productController.GetProductById)
		protected.PUT("/products/:productId", productController.UpdateProduct)
		protected.DELETE("/products/:productId", productController.Delete)
	}

	// Porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 👇 HTTP server custom
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 👇 Rodar servidor em goroutine
	go func() {
		log.Println("API Volurya rodando na porta", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro ao iniciar servidor: %v", err)
		}
	}()

	// 👇 Capturar sinais do sistema (Docker, Render, etc)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Desligando servidor...")

	// 👇 Timeout pra finalizar requisições
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Erro no shutdown:", err)
	}

	log.Println("Servidor finalizado corretamente")
}
