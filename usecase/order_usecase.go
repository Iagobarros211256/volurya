package usecase

import (
	"api/metrics"
	"api/models"
	"api/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type OrderUsecase struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
}

func NewOrderUsecase(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

type PagSeguroChargeRequest struct {
	ReferenceID string `json:"reference_id"`
	Description string `json:"description"`
	Amount      struct {
		Value    int    `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"payment_method"`
	NotificationURL string `json:"notification_url"`
}

type PagSeguroChargeResponse struct {
	ID          string `json:"id"`
	PaymentLink string `json:"payment_link"`
	Status      string `json:"status"`
}

func (ou *OrderUsecase) CreateOrder(userID int, productID int, quantity int) (string, error) {
	if quantity <= 0 {
		return "", ErrInvalidQuantity
	}

	// Busca o produto usando o ProductRepository
	product, err := ou.productRepo.GetProductById(productID)
	if err != nil {
		return "", fmt.Errorf("produto não encontrado: %w", err)
	}

	if product.Stock < quantity {
		return "", fmt.Errorf("%w: requested %d, available %d", ErrInsufficientStock, quantity, product.Stock)
	}

	// Cria a ordem no banco
	order := models.Order{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Total:     product.Price * float64(quantity),
		Status:    "pending",
	}

	orderID, err := ou.orderRepo.CreateOrder(order)
	if err != nil {
		return "", fmt.Errorf("falha ao criar ordem: %w", err)
	}

	metrics.OrdersTotal.Inc()

	// Configuração PagSeguro
	isSandbox := os.Getenv("PAGSEGURO_SANDBOX") != "false"
	baseURL := "https://sandbox.api.pagseguro.com"
	if !isSandbox {
		baseURL = "https://api.pagseguro.com"
	}

	payload := PagSeguroChargeRequest{
		ReferenceID: fmt.Sprintf("ORDER-%d", orderID),
		Description: fmt.Sprintf("Compra Volurya - Produto %d", productID),
		Amount: struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		}{
			Value:    int(order.Total * 100), // centavos
			Currency: "BRL",
		},
		PaymentMethod: struct {
			Type string `json:"type"`
		}{
			Type: "PIX",
		},
		NotificationURL: os.Getenv("PAGSEGURO_WEBHOOK_URL"),
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", baseURL+"/charges", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("PAGSEGURO_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha ao chamar PagSeguro: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("PagSeguro retornou status %d", resp.StatusCode)
	}

	var chargeResp PagSeguroChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return "", err
	}

	// Atualiza a ordem com o ID da cobrança
	err = ou.orderRepo.UpdateOrderChargeID(orderID, chargeResp.ID)
	if err != nil {
		return "", fmt.Errorf("falha ao atualizar ordem: %w", err)
	}

	return chargeResp.PaymentLink, nil
}

/*

 PagSeguro e Stripe coexistindo no mesmo projeto
Esse arquivo usa PagSeguro, mas payment_controller.go, payment_usecase.go e as migrations usam Stripe. O projeto tem dois provedores de pagamento ativos simultaneamente — qual é o real? Isso explica o comentário desatualizado no order_controller.go. O código do PagSeguro deveria ser removido completamente.

🔴 HTTP client direto no usecase
goclient := &http.Client{}
resp, err := client.Do(req)
Lógica de infraestrutura (HTTP, API externa) dentro do usecase viola separação de responsabilidades e torna testes impossíveis sem chamadas reais ao PagSeguro. Extraia para um gateway:
gotype PaymentGateway interface {
    CreateCharge(orderID int, amount int, currency string) (string, error)
}

🔴 json.Marshal com erro ignorado
gojsonPayload, _ := json.Marshal(payload)
Se a serialização falhar, jsonPayload é nil e a requisição HTTP vai com body vazio:
gojsonPayload, err := json.Marshal(payload)
if err != nil {
    return "", fmt.Errorf("failed to marshal payload: %w", err)
}

🔴 Token de API lido a cada chamada
goreq.Header.Set("Authorization", "Bearer "+os.Getenv("PAGSEGURO_TOKEN"))
os.Getenv a cada request é ineficiente e não valida se o token está configurado. Valide no startup e injete via construtor.

🔴 Sem rollback se PagSeguro falhar
goorderID, err := ou.orderRepo.CreateOrder(order)  // ordem criada
// ...
resp, err := client.Do(req)
if err != nil {
    return "", fmt.Errorf("falha ao chamar PagSeguro: %w", err)  // ordem órfã
}
Mesmo problema do payment_controller.go — ordem criada sem payment link fica órfã no banco.

🔴 Estoque não decrementado após criar ordem
goif product.Stock < quantity {
    return "", ErrInsufficientStock
}
// cria ordem...
// estoque nunca é decrementado
Dois usuários podem comprar o último item simultaneamente — ambos passam na validação de estoque, ambos criam ordens, estoque fica negativo. Deveria usar UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1 atomicamente.

🔴 Dependências concretas
goorderRepo   *repository.OrderRepository
productRepo *repository.ProductRepository
Padrão recorrente — sem interfaces.

🟡 isSandbox com lógica invertida
goisSandbox := os.Getenv("PAGSEGURO_SANDBOX") != "false"
Por padrão é sandbox (quando env não está definida). O padrão seguro deveria ser produção exigindo opt-in explícito para sandbox, não o contrário — evita cobranças reais acidentais em dev, mas também evita não cobrar em produção por env não configurada:
goisSandbox := os.Getenv("PAGSEGURO_SANDBOX") == "true"

🟡 Mensagens de erro em português
goreturn "", fmt.Errorf("produto não encontrado: %w", err)
return "", fmt.Errorf("falha ao criar ordem: %w", err)
Mistura com inglês nos outros usecases — padronize.

🟡 int(order.Total * 100) — imprecisão de float
Já apontado no payment_controller.go — use math.Round:
goamountCents := int64(math.Round(order.Total * 100))

*/
