package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api/db"
)

func TestCreateUser(t *testing.T) {
	// 1️⃣ Conecta no banco de teste
	database := db.SetupTestDB()
	defer database.Close()

	// 2️⃣ Limpa e garante tabela
	if err := db.CleanTestDB(database); err != nil {
		t.Fatalf("failed to clean db: %v", err)
	}

	if err := db.EnsureTablesExist(database); err != nil {
		t.Fatalf("failed to ensure tables: %v", err)
	}

	// 3️⃣ Sobe o router
	r := SetupRouter(database)

	// 4️⃣ Payload do usuário
	payload := map[string]string{
		"name":  "Iago",
		"email": "iago@test.com",
	}

	body, _ := json.Marshal(payload)

	// 5️⃣ Cria request HTTP
	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	// 6️⃣ Executa request
	r.ServeHTTP(resp, req)

	// 7️⃣ Valida status HTTP
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}

	// 8️⃣ Valida que gravou no banco
	var count int
	err := database.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email = $1",
		"iago@test.com",
	).Scan(&count)

	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}
}
