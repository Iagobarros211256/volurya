package usecase

import (
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
	// Busca o produto usando o ProductRepository
	product, err := ou.productRepo.GetProductById(productID)
	if err != nil {
		return "", fmt.Errorf("produto não encontrado: %w", err)
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
