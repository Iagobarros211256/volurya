package config

import "testing"

func TestSalesWhatsAppURL(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		phone string
		want  string
	}{
		{
			name:  "modo desativado",
			mode:  "",
			phone: "5585999999999",
			want:  "",
		},
		{
			name:  "numero ausente",
			mode:  "whatsapp",
			phone: "",
			want:  "",
		},
		{
			name:  "numero com caracteres invalidos",
			mode:  "whatsapp",
			phone: "+55 (85) 99999-9999",
			want:  "",
		},
		{
			name:  "numero internacional invalido",
			mode:  "whatsapp",
			phone: "1585999999999",
			want:  "",
		},
		{
			name:  "numero valido",
			mode:  "whatsapp",
			phone: "5585999999999",
			want:  "https://wa.me/5585999999999",
		},
		{
			name:  "modo com letras maiusculas",
			mode:  "WHATSAPP",
			phone: "5585999999999",
			want:  "https://wa.me/5585999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VOLURYA_SALES_MODE", tt.mode)
			t.Setenv("VOLURYA_SALES_WHATSAPP", tt.phone)

			got := SalesWhatsAppURL()
			if got != tt.want {
				t.Errorf(
					"SalesWhatsAppURL() = %q; esperado %q",
					got, tt.want,
				)
			}
		})
	}
}
