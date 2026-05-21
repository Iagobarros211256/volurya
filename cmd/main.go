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
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// Iniciar worker pool (4 workers)
	imageProcessor := jobs.NewImageProcessor(4)
	defer imageProcessor.Shutdown()
	notificationHub := notifications.NewHub()

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
	///ping e /swagger sem proteção
	//Swagger em produção expõe toda a estrutura da API. O padrão é desabilitar em produção
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	///metrics exposto sem autenticação  BUG!!!!!
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

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

	// Initialize Stripe
	if err := config.InitStripe(); err != nil {
		slog.Warn("Stripe initialization warning", "error", err)
		// Don't exit - Stripe is optional for development
	}

	// Migrations
	if err := db.RunMigrations(dbConnection); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	//db.BootstrapAdminUser no startup é arriscado
	//Criar usuário admin automaticamente no boot pode ser um vetor de ataque se
	//a lógica não for idempotente e bem protegida. Verifique se verifica existência antes de criar,
	//  e se a senha vem de variável de ambiente.
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
		imageProcessor, // NOVO - passe o processor aqui. esqueci o que isso significa
	)
	authController := controller.NewAuthController(authUseCase)
	orderController := controller.NewOrderController(orderUsecase, notificationHub)
	cartController := controller.NewCartController(cartUsecase, orderUsecase, notificationHub)
	notificationController := controller.NewNotificationController(notificationHub)
	healthController := controller.NewHealthController(dbConnection)
	paymentController := controller.NewPaymentController(paymentUsecase)

	// Rotas públicas
	authLimiter := middleware.NewRateLimiter(5, 1*time.Minute)     // 5 req/min
	refreshLimiter := middleware.NewRateLimiter(10, 1*time.Minute) // 10 req/min

	// Webhook route (public, but needs signature validation)
	//variáveis usadas antes de serem declaradas BUG!!!!
	public.POST("/webhook", paymentController.Webhook)

	//variáveis usadas antes de serem declaradas BUG!!!!
	// Payment routes (protected)
	protected.POST("/checkout", paymentController.Checkout)

	// Rotas protegidas
	////variáveis usadas antes de serem declaradas BUG!!!!
	protected := router.Group("/api")
	protected.Use(auth.Middleware())
	{ //products protected routes
		protected.GET("/products", productController.GetProducts)
		protected.GET("/products/:productId", productController.GetProductById)

		// Product routes requiring admin role
		adminProduct := protected.Group("/products")
		adminProduct.Use(auth.RequireAdminRole())
		{
			adminProduct.POST("", productController.CreateProduct)
			adminProduct.PUT("/:productId", productController.UpdateProduct)
			adminProduct.DELETE("/:productId", productController.Delete)
		}

		protected.POST("/orders", orderController.CreateOrder)
		protected.GET("/events", notificationController.Stream)

		//cart protected routes
		protected.GET("/cart", cartController.GetCart)
		protected.POST("/cart/items", cartController.AddItem)
		protected.PUT("/cart/items/:itemId", cartController.UpdateItem)
		protected.DELETE("/cart/items/:itemId", cartController.RemoveItem)
		protected.POST("/cart/checkout", cartController.Checkout)

		//images protected route (admin only)
		adminProduct.POST("/:productId/image", productController.UploadImage)
	}

	// Porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		// Porta hardcoded como fallback está ok, mas poderia ser constante
	}

	// 👇 HTTP server custom
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	srv.RegisterOnShutdown(notificationHub.Close)

	// HTTPS Redirect em produção
	env := os.Getenv("ENV")
	if env == "production" {
		go func() {
			slog.Info("HTTP redirect server rodando na porta 80")
			redirectServer := &http.Server{
				Addr: ":80",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					target := "https://" + r.Host + r.RequestURI
					http.Redirect(w, r, target, http.StatusMovedPermanently)
				}),
			}
			if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("erro no redirect server", "error", err)
			}
		}()
	}

	// 👇 Rodar servidor em goroutine
	go func() {
		slog.Info("API Volurya rodando na porta", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			//porque esse log ainda existe?  Padronize tudo com slog para manter logs estruturados e consistentes.
			log.Fatalf("erro ao iniciar servidor: %v", err)
		}
	}()

	// 👇 Capturar sinais do sistema (Docker, Render, etc)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Desligando servidor...")

	// 👇 Timeout pra finalizar requisições
	//Se houver requisições longas (upload de imagem, checkout),
	// 5s pode não ser suficiente. Considere 15-30s, ou tornar configurável via env.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Erro no shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Servidor finalizado corretamente")
}
