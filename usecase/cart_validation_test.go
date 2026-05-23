package usecase

import (
	"testing"
)

func TestCartValidation_PositivePrice(t *testing.T) {
	cartItems := []struct {
		name     string
		price    float64
		quantity int
		valid    bool
	}{
		{"item with price", 10.50, 1, true},
		{"item with zero price", 0, 1, false},
		{"item with negative price", -5.00, 1, false},
		{"item with zero quantity", 10.50, 0, false},
		{"item with negative quantity", 10.50, -1, false},
	}

	for _, item := range cartItems {
		t.Run(item.name, func(t *testing.T) {
			isValid := item.price > 0 && item.quantity > 0
			if isValid != item.valid {
				t.Fatalf("expected valid=%v, got %v", item.valid, isValid)
			}
		})
	}
}

/*

 Testa lógica escrita no próprio teste
goisValid := item.price > 0 && item.quantity > 0
if isValid != item.valid {
A validação item.price > 0 && item.quantity > 0 está escrita aqui no teste, não no código de produção. Você está verificando se sua própria expressão condicional funciona — isso nunca vai falhar a menos que você escreva o teste errado.
O teste deveria chamar uma função real do usecase:
gofunc TestCartValidation_AddItem(t *testing.T) {
    uc := setupCartUsecase(t)

    tests := []struct {
        name      string
        productID int
        quantity  int
        wantErr   bool
    }{
        {"quantidade válida", 1, 1, false},
        {"quantidade zero", 1, 0, true},
        {"quantidade negativa", 1, -1, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := uc.AddItem(1, tt.productID, tt.quantity)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddItem() error = %v, wantErr = %v", err, tt.wantErr)
            }
        })
    }
}

🔴 Validação de preço não existe no código de produção
O teste valida price > 0 mas o cart_usecase.go não tem nenhuma validação de preço — só valida quantity. Esse teste está cobrindo um comportamento que não existe na aplicação.

🟡 Nome do arquivo e função são enganosos
cart_validation_test.go com TestCartValidation_PositivePrice sugere que testa validação do carrinho, mas na realidade não testa nada do carrinho. Um desenvolvedor novo vai assumir que essa cobertura existe e não vai escrever testes reais.

*/
