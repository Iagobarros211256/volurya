package config

import "testing"

func TestGetPublicSalesConfig(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		phone        string
		payments     string
		wantMode     string
		wantURL      string
		wantPayments bool
	}{
		{
			name:         "vendas pelo whatsapp",
			mode:         "whatsapp",
			phone:        "5585999999999",
			payments:     "false",
			wantMode:     "whatsapp",
			wantURL:      "https://wa.me/5585999999999",
			wantPayments: false,
		},
		{
			name:         "whatsapp sem telefone",
			mode:         "whatsapp",
			phone:        "",
			payments:     "false",
			wantMode:     "whatsapp",
			wantURL:      "",
			wantPayments: false,
		},
		{
			name:         "modo tradicional com pagamentos",
			mode:         "",
			phone:        "",
			payments:     "true",
			wantMode:     "",
			wantURL:      "",
			wantPayments: true,
		},
		{
			name:         "modo normalizado",
			mode:         " WHATSAPP ",
			phone:        "5585999999999",
			payments:     "false",
			wantMode:     "whatsapp",
			wantURL:      "https://wa.me/5585999999999",
			wantPayments: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VOLURYA_SALES_MODE", tt.mode)
			t.Setenv("VOLURYA_SALES_WHATSAPP", tt.phone)
			t.Setenv("PAYMENTS_ENABLED", tt.payments)

			got := GetPublicSalesConfig()

			if got.SalesMode != tt.wantMode {
				t.Errorf("modo=%q; esperado=%q", got.SalesMode, tt.wantMode)
			}
			if got.SalesWhatsAppURL != tt.wantURL {
				t.Errorf("URL=%q; esperado=%q", got.SalesWhatsAppURL, tt.wantURL)
			}
			if got.PaymentsEnabled != tt.wantPayments {
				t.Errorf(
					"pagamentos=%v; esperado=%v",
					got.PaymentsEnabled,
					tt.wantPayments,
				)
			}
		})
	}
}
