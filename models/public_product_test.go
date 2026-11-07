package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPublicProductDoesNotExposeUserID(t *testing.T) {
	product := Product{
		ID:          42,
		UserID:      999,
		Name:        "Camiseta Volurya",
		Description: "Produto oficial da banda",
		Price:       79.90,
		Stock:       5,
		ImageURL:    "/static/imagens/camiseta.jpg",
	}

	public := ToPublicProduct(product)

	data, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("serializar produto público: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("interpretar JSON: %v", err)
	}

	if _, exists := result["user_id"]; exists {
		t.Fatal("campo interno user_id exposto no catálogo público")
	}

	if strings.Contains(string(data), `"999"`) {
		t.Fatal("identificador interno encontrado na resposta pública")
	}

	if result["id_product"] != float64(42) {
		t.Errorf("id_product inesperado: %v", result["id_product"])
	}

	if result["name"] != "Camiseta Volurya" {
		t.Errorf("nome inesperado: %v", result["name"])
	}

	if result["price"] != 79.90 {
		t.Errorf("preço inesperado: %v", result["price"])
	}

	if result["stock"] != float64(5) {
		t.Errorf("estoque inesperado: %v", result["stock"])
	}

	if result["image_url"] != "/static/imagens/camiseta.jpg" {
		t.Errorf("imagem inesperada: %v", result["image_url"])
	}
}
