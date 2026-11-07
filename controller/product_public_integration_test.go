//go:build integration

package controller

import (
	"api/db"
	"api/repository"
	"api/usecase"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPublicProductCatalogWithoutAuthentication(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"public-catalog@example.invalid",
		"test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products
		 (user_id, name, description, price, stock)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		userID,
		"Camiseta Volurya",
		"Produto oficial",
		79.90,
		5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	productRepo := repository.NewProductRepository(conn)
	productUsecase := usecase.NewProductUsecase(productRepo)
	productController := NewProductController(
		productUsecase,
		productRepo,
		nil,
		nil,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mesmos handlers públicos, sem middleware de autenticação.
	router.GET("/api/products", productController.GetProducts)
	router.GET("/api/products/:productId", productController.GetProductById)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/products?limit=12",
		nil,
	)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status esperado=200, obtido=%d, corpo=%s",
			rec.Code,
			rec.Body.String(),
		)
	}

	var response struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			HasMore bool `json:"has_more"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if len(response.Data) != 1 {
		t.Fatalf(
			"esperado 1 produto, obtidos %d",
			len(response.Data),
		)
	}

	product := response.Data[0]

	if _, exists := product["user_id"]; exists {
		t.Fatal("API pública expôs user_id")
	}

	if product["id_product"] != float64(productID) {
		t.Errorf("ID incorreto: %v", product["id_product"])
	}

	if product["name"] != "Camiseta Volurya" {
		t.Errorf("nome incorreto: %v", product["name"])
	}

	if product["stock"] != float64(5) {
		t.Errorf("estoque incorreto: %v", product["stock"])
	}

	if response.Pagination.HasMore {
		t.Error("paginação indicou produtos adicionais indevidamente")
	}
}

func TestPublicProductDetailWithoutAuthentication(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"public-detail@example.invalid",
		"test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products
		 (user_id, name, description, price, stock)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, "Camiseta Volurya", "Merch oficial", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	productRepo := repository.NewProductRepository(conn)
	productController := NewProductController(
		usecase.NewProductUsecase(productRepo),
		productRepo,
		nil,
		nil,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products/:productId", productController.GetProductById)

	t.Run("produto existente", func(t *testing.T) {
		path := "/api/products/" + strconv.Itoa(productID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, corpo=%s", rec.Code, rec.Body.String())
		}

		var product map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
			t.Fatal(err)
		}

		if _, exists := product["user_id"]; exists {
			t.Fatal("consulta pública expôs user_id")
		}

		if product["id_product"] != float64(productID) {
			t.Errorf("ID incorreto: %v", product["id_product"])
		}

		if product["name"] != "Camiseta Volurya" {
			t.Errorf("nome incorreto: %v", product["name"])
		}
	})

	t.Run("produto inexistente", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/products/2147483647",
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf(
				"esperado 404, obtido=%d, corpo=%s",
				rec.Code,
				rec.Body.String(),
			)
		}
	})
}
