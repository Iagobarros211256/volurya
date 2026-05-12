package main

import (
	"api/auth"
	"api/controller"
	"api/db"
	"api/logger"
	"api/middleware"
	"api/repository"
	"api/storage"
	"api/usecase"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	logger.Init()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

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

	router.GET("/about", func(c *gin.Context) {
		c.HTML(200, "about.html", gin.H{})
	})

	router.GET("/cart", func(c *gin.Context) {
		c.HTML(200, "cart.html", gin.H{})
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

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	// DB
	dbConnection, err := db.ConnectDB()
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Storage R2
	r2Storage, err := storage.NewR2Storage()
	if err != nil {
		slog.Error("failed to initialize R2 storage", "error", err)
		os.Exit(1)
	}

	// Migrations
	if err := db.RunMigrations(dbConnection); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Repositories
	productRepository := repository.NewProductRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	orderRepository := repository.NewOrderRepository(dbConnection)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dbConnection)
	cartRepository := repository.NewCartRepository(dbConnection)

	// Usecases
	productUseCase := usecase.NewProductUsecase(productRepository)
	authUseCase := usecase.NewAuthUsecase(userRepository, refreshTokenRepository)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	cartUsecase := usecase.NewCartUsecase(cartRepository, productRepository)

	// Controllers
	productController := controller.NewProductController(productUseCase, r2Storage)
	authController := controller.NewAuthController(authUseCase)
	orderController := controller.NewOrderController(orderUsecase)
	cartController := controller.NewCartController(cartUsecase, orderUsecase)

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
	{ //products protected routes
		protected.GET("/products", productController.GetProducts)
		protected.POST("/products", productController.CreateProduct)
		protected.POST("/orders", orderController.CreateOrder)
		protected.GET("/products/:productId", productController.GetProductById)
		protected.PUT("/products/:productId", productController.UpdateProduct)
		protected.DELETE("/products/:productId", productController.Delete)

		//cart protected routes
		protected.GET("/cart", cartController.GetCart)
		protected.POST("/cart/items", cartController.AddItem)
		protected.PUT("/cart/items/:itemId", cartController.UpdateItem)
		protected.DELETE("/cart/items/:itemId", cartController.RemoveItem)
		protected.POST("/cart/checkout", cartController.Checkout)

		//images protected route
		protected.POST("/products/:productId/image", productController.UploadImage)
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
		slog.Info("API Volurya rodando na porta", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro ao iniciar servidor: %v", err)
		}
	}()

	// 👇 Capturar sinais do sistema (Docker, Render, etc)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Desligando servidor...")

	// 👇 Timeout pra finalizar requisições
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Erro no shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Servidor finalizado corretamente")
}
