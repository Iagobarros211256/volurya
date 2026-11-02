package db

import "testing"

func TestIsAllowedTestDatabase(t *testing.T) {
	tests := []struct {
		name, database, user string
		allowed              bool
	}{
		{"geral autorizado", "volurya_test", "test", true},
		{"financeiro autorizado", "volurya_payment_test", "test", true},
		{"produção recusada", "volurya_db", "volurya", false},
		{"produção com usuário test", "volurya_db", "test", false},
		{"usuário não autorizado", "volurya_test", "postgres", false},
		{"banco desconhecido", "outro_banco", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedTestDatabase(tt.database, tt.user)
			if got != tt.allowed {
				t.Errorf("permissão = %v, esperado %v", got, tt.allowed)
			}
		})
	}
}
