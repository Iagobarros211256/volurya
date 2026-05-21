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

/*


Métricas declaradas mas não todas usadas
Visto nos controllers, apenas ImageUploadsTotal é incrementado (metrics.ImageUploadsTotal.Inc()). UsersTotal, ProductsTotal, OrdersTotal e CartItemsAdded provavelmente não são incrementados em lugar nenhum — métricas que nunca mudam não têm valor. Verifique e conecte nos usecases/controllers corretos.

🟡 HttpRequestsTotal e HttpRequestDuration não são usados no middleware
O middleware/logger.go provavelmente não incrementa essas métricas. Sem isso, as métricas HTTP mais importantes ficam zeradas:
go// middleware/logger.go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start)

        metrics.HttpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            strconv.Itoa(c.Writer.Status()),
        ).Inc()

        metrics.HttpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration.Seconds())
    }
}

🟡 Sem métricas de negócio importantes
Para um e-commerce, faltam métricas críticas:
go// Pagamentos
PaymentsSucceeded = promauto.NewCounter(...)
PaymentsFailed    = promauto.NewCounter(...)

// Checkouts
CheckoutsTotal    = promauto.NewCounter(...)

// Erros
ErrorsTotal = promauto.NewCounterVec(..., []string{"type"})

🟡 prometheus.DefBuckets pode não ser ideal para sua API
goBuckets: prometheus.DefBuckets,
// [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10] segundos
Se a maioria das requisições responde em menos de 50ms, os buckets padrão têm resolução baixa onde importa. Ajuste para sua realidade:
goBuckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2},

🟢 Variáveis globais exportadas
Mesmo padrão do stripe.go — qualquer pacote pode modificar os contadores diretamente. Para métricas isso é menos crítico que para chaves de API, mas encapsular em funções seria mais limpo:
gofunc IncOrdersTotal() {
    OrdersTotal.Inc()
}


*/
