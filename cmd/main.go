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

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	// DB
	dbConnection, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Repositories
	productRepository := repository.NewProductRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	orderRepository := repository.NewOrderRepository(dbConnection)

	// Usecases
	productUseCase := usecase.NewProductUsecase(productRepository)
	authUseCase := usecase.NewAuthUsecase(userRepository)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)

	// Controllers
	productController := controller.NewProductController(productUseCase)
	authController := controller.NewAuthController(authUseCase)
	orderController := controller.NewOrderController(orderUsecase)

	// Rotas
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api.POST("/signup", authController.Signup)
		api.POST("/login", authController.Login)
	}

	protected := router.Group("/api")
	protected.Use(auth.Middleware())
	{
		protected.GET("/products", productController.GetProducts)
		protected.POST("/products", productController.CreateProduct)
		protected.POST("/orders", orderController.CreateOrder)
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
