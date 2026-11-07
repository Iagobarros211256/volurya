package config

import (
	"net/url"
	"os"
	"strings"
)

// SalesWhatsAppURL retorna a URL pública de contato para vendas.
// Se o modo WhatsApp não estiver habilitado ou o telefone
// não for válido, retorna uma string vazia.
func SalesWhatsAppURL() string {
	if strings.ToLower(strings.TrimSpace(
		os.Getenv("VOLURYA_SALES_MODE"),
	)) != "whatsapp" {
		return ""
	}

	phone := strings.TrimSpace(os.Getenv("VOLURYA_SALES_WHATSAPP"))

	// Exige número internacional brasileiro: 55 + DDD + telefone.
	if len(phone) != 12 && len(phone) != 13 {
		return ""
	}
	if !strings.HasPrefix(phone, "55") {
		return ""
	}
	for _, digit := range phone {
		if digit < '0' || digit > '9' {
			return ""
		}
	}

	return (&url.URL{
		Scheme: "https",
		Host:   "wa.me",
		Path:   "/" + phone,
	}).String()
}
