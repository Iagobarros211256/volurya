package controller

import (
	"database/sql"
	"testing"
	"time"
)

// Só verifica que o construtor não retorna nil — o que nunca vai falhar
// a menos que o código tenha um bug óbvio. Não testa nenhum comportamento real do health check.
func TestHealthCheck_ValidDB(t *testing.T) {
	// Mock database that works
	//sql.DB{} vazio não simula falha de Ping
	//O sql.DB sem conexão real vai retornar erro no Ping(),
	// mas de forma imprevisível. Para testar comportamento de DB down/up corretamente, use uma interface:
	mockDB := &sql.DB{}

	controller := NewHealthController(mockDB)
	if controller == nil {
		t.Fatal("controller should not be nil")
	}
}

// Você criou o objeto com "healthy" e verificou se é "healthy".
// Isso nunca vai falhar. Não testa o controller, não testa lógica nenhuma.
func TestHealthResponse_Format(t *testing.T) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "up",
		Version:   "1.0.0",
	}

	if response.Status != "healthy" {
		t.Fatalf("expected status 'healthy', got %s", response.Status)
	}

	if response.Database != "up" {
		t.Fatalf("expected database 'up', got %s", response.Database)
	}

	if response.Version != "1.0.0" {
		t.Fatalf("expected version '1.0.0', got %s", response.Version)
	}
}

//O que esses testes deveriam cobrir:
// 1. DB respondendo — retorna 200 com status healthy
//func TestHealthCheck_DBUp(t *testing.T) {
//    db := setupTestDB(t)  // banco real de teste ou mock com Ping mockado
//    controller := NewHealthController(db)
//
//    w := httptest.NewRecorder()
//    c, _ := gin.CreateTestContext(w)
//    controller.HealthCheck(c)

//    assert.Equal(t, http.StatusOK, w.Code)

//    var resp HealthResponse
//    json.NewDecoder(w.Body).Decode(&resp)
//    assert.Equal(t, "healthy", resp.Status)
//    assert.Equal(t, "up", resp.Database)
//}

// 2. DB fora — retorna 503 com status degraded/unhealthy
//func TestHealthCheck_DBDown(t *testing.T) {
// DB com Ping que falha
//    ...
//    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
//    assert.Equal(t, "unhealthy", resp.Status)
//}

// 3. Timestamp no formato correto
//func TestHealthCheck_TimestampFormat(t *testing.T) {
// Verifica que o timestamp retornado é um RFC3339 válido
//    _, err := time.Parse(time.RFC3339, resp.Timestamp)
//    assert.NoError(t, err)
//}
