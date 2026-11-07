package config

import "testing"

func TestCartAvailable(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		payments string
		want     bool
	}{
		{
			name:     "whatsapp com pagamentos desabilitados",
			mode:     "whatsapp",
			payments: "false",
			want:     false,
		},
		{
			name:     "whatsapp com pagamentos habilitados",
			mode:     "whatsapp",
			payments: "true",
			want:     false,
		},
		{
			name:     "modo tradicional com pagamentos desabilitados",
			mode:     "",
			payments: "false",
			want:     false,
		},
		{
			name:     "modo tradicional com pagamentos habilitados",
			mode:     "",
			payments: "true",
			want:     true,
		},
		{
			name:     "modo whatsapp em letras maiusculas",
			mode:     "WHATSAPP",
			payments: "true",
			want:     false,
		},
		{
			name:     "pagamentos nao configurados",
			mode:     "",
			payments: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VOLURYA_SALES_MODE", tt.mode)
			t.Setenv("PAYMENTS_ENABLED", tt.payments)

			if got := CartAvailable(); got != tt.want {
				t.Errorf(
					"CartAvailable() = %v; esperado %v",
					got, tt.want,
				)
			}
		})
	}
}
