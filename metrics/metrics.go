package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Total de requisições HTTP
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volurya_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Duração das requisições HTTP
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "volurya_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Usuários cadastrados
	UsersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "volurya_users_total",
			Help: "Total number of registered users",
		},
	)

	// Produtos criados
	ProductsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "volurya_products_total",
			Help: "Total number of products created",
		},
	)

	// Ordens criadas
	OrdersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "volurya_orders_total",
			Help: "Total number of orders created",
		},
	)

	// Itens adicionados ao carrinho
	CartItemsAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "volurya_cart_items_added_total",
			Help: "Total number of items added to cart",
		},
	)

	// Uploads de imagem
	ImageUploadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "volurya_image_uploads_total",
			Help: "Total number of image uploads",
		},
	)
)
