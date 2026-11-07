package config

import (
	"os"
	"strings"
)

// PublicSalesConfig contém somente informações públicas
// necessárias para a navegação e o catálogo.
type PublicSalesConfig struct {
	SalesMode        string `json:"sales_mode"`
	SalesWhatsAppURL string `json:"sales_whatsapp_url"`
	PaymentsEnabled  bool   `json:"payments_enabled"`
}

// GetPublicSalesConfig consulta a configuração do ambiente.
// Não inclui credenciais nem segredos de pagamento.
func GetPublicSalesConfig() PublicSalesConfig {
	mode := strings.ToLower(strings.TrimSpace(
		os.Getenv("VOLURYA_SALES_MODE"),
	))

	return PublicSalesConfig{
		SalesMode:        mode,
		SalesWhatsAppURL: SalesWhatsAppURL(),
		PaymentsEnabled:  os.Getenv("PAYMENTS_ENABLED") == "true",
	}
}
