package config

import (
	"os"
	"strings"
)

// CartAvailable determina se o carrinho pode ser apresentado
// ao visitante. A autorização das operações de pagamento
// continua sendo responsabilidade do backend.
func CartAvailable() bool {
	salesMode := strings.ToLower(
		strings.TrimSpace(os.Getenv("VOLURYA_SALES_MODE")),
	)

	return salesMode != "whatsapp" &&
		os.Getenv("PAYMENTS_ENABLED") == "true"
}
