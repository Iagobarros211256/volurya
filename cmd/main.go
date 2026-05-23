package main

import (
	"api/auth"
	"api/config"
	"api/controller"
	"api/db"
	"api/jobs"
	"api/logger"
	"api/middleware"
	"api/notifications"
	"api/repository"
	"api/storage"
	"api/usecase"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "api/docs"
)

// @title Volurya API
// @version 1.0
// @description API de e-commerce para a Volurya Shop
// @termsOfService http://swagger.io/terms/

// @contact.name Volurya Support
// @contact.url http://www.volurya.com/support
// @contact.email support@volurya.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.volurya.com
// @BasePath /api
// @schemes https http

func main() {
	logger.Init()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS())
	router.Use(middleware.CSRFProtection())
	router.Use(middleware.CSRFTokenProvider())

	// Worker pool para processamento de imagens
	imageProcessor := jobs.NewImageProcessor(4)
	defer imageProcessor.Shutdown()
	notificationHub := notifications.NewHub()

	router.LoadHTMLGlob("views/templates/*")
	router.Static("/static", "./views/static")

	// Rotas de páginas
	router.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", gin.H{}) })
	router.GET("/store", func(c *gin.Context) {
		c.HTML(200, "store.html", gin.H{"title": "VOLURYA SHOP - Produtos Oficiais"})
	})
	router.GET("/about", func(c *gin.Context) { c.HTML(200, "about.html", gin.H{}) })
	router.GET("/cart", func(c *gin.Context) { c.HTML(200, "cart.html", gin.H{}) })
	router.GET("/login", func(c *gin.Context) { c.HTML(200, "login.html", gin.H{}) })
	router.GET("/signup", func(c *gin.Context) { c.HTML(200, "signup.html", gin.H{}) })

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Swagger apenas fora de produção
	if os.Getenv("APP_ENV") != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	// Métricas protegidas por IP local apenas
	router.GET("/metrics", func(c *gin.Context) {
		ip := c.ClientIP()
		if ip != "127.0.0.1" && ip != "::1" && !strings.HasPrefix(ip, "172.") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	router.HEAD("/", func(c *gin.Context) { c.Status(200) })

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

	// Stripe
	if err := config.InitStripe(); err != nil {
		slog.Warn("Stripe initialization warning", "error", err)
	}

	// Migrations
	if err := db.RunMigrations(dbConnection); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := db.BootstrapAdminUser(dbConnection); err != nil {
		slog.Error("failed to bootstrap admin user", "error", err)
		os.Exit(1)
	}

	// Repositories
	productRepository := repository.NewProductRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	orderRepository := repository.NewOrderRepository(dbConnection)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dbConnection)
	cartRepository := repository.NewCartRepository(dbConnection)
	paymentRepository := repository.NewPaymentRepository(dbConnection)

	// Usecases
	productUsecase := usecase.NewProductUsecase(productRepository)
	authUseCase := usecase.NewAuthUsecase(userRepository, refreshTokenRepository)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	cartUsecase := usecase.NewCartUsecase(cartRepository, productRepository)
	paymentUsecase := usecase.NewPaymentUsecase(orderRepository, paymentRepository, productRepository)

	// Controllers
	productController := controller.NewProductController(
		productUsecase,
		productRepository,
		r2Storage,
		imageProcessor,
	)
	authController := controller.NewAuthController(authUseCase)
	orderController := controller.NewOrderController(orderUsecase, notificationHub)
	cartController := controller.NewCartController(cartUsecase, orderUsecase, notificationHub)
	notificationController := controller.NewNotificationController(notificationHub)
	healthController := controller.NewHealthController(dbConnection)
	paymentController := controller.NewPaymentController(paymentUsecase)

	// Rate limiters
	authLimiter := middleware.NewRateLimiter(5, 1*time.Minute)
	refreshLimiter := middleware.NewRateLimiter(10, 1*time.Minute)

	// Rotas públicas
	public := router.Group("/api")
	{
		public.POST("/signup", authLimiter.Middleware(), authController.Signup)
		public.POST("/login", authLimiter.Middleware(), authController.Login)
		public.POST("/refresh", refreshLimiter.Middleware(), authController.RefreshToken)
		public.POST("/logout", authController.Logout)
		public.GET("/health", healthController.Health)
		public.POST("/webhook", paymentController.Webhook)
	}

	// Rotas protegidas
	protected := router.Group("/api")
	protected.Use(auth.Middleware())
	{
		// Produtos
		protected.GET("/products", productController.GetProducts)
		protected.GET("/products/:productId", productController.GetProductById)

		// Produtos — admin only
		adminProduct := protected.Group("/products")
		adminProduct.Use(auth.RequireAdminRole())
		{
			adminProduct.POST("", productController.CreateProduct)
			adminProduct.PUT("/:productId", productController.UpdateProduct)
			adminProduct.DELETE("/:productId", productController.Delete)
			adminProduct.POST("/:productId/image", productController.UploadImage)
		}

		// Checkout — protegido
		protected.POST("/checkout", paymentController.Checkout)

		// Ordens
		protected.POST("/orders", orderController.CreateOrder)

		// SSE
		protected.GET("/events", notificationController.Stream)

		// Carrinho
		protected.GET("/cart", cartController.GetCart)
		protected.POST("/cart/items", cartController.AddItem)
		protected.PUT("/cart/items/:itemId", cartController.UpdateItem)
		protected.DELETE("/cart/items/:itemId", cartController.RemoveItem)
		protected.POST("/cart/checkout", cartController.Checkout)
	}

	// Porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	srv.RegisterOnShutdown(notificationHub.Close)

	// HTTPS redirect em produção
	if os.Getenv("APP_ENV") == "production" {
		go func() {
			slog.Info("HTTP redirect server rodando na porta 80")
			redirectServer := &http.Server{
				Addr: ":80",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
				}),
			}
			if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("erro no redirect server", "error", err)
			}
		}()
	}

	go func() {
		slog.Info("API Volurya rodando", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("erro ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Desligando servidor...")

	shutdownTimeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro no shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Servidor finalizado corretamente")
}
