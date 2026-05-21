package controller

import (
	"api/models"
	"api/repository"
	"api/usecase"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockProductRepository e mockProductUsecase declarados mas não usados
// Estão no arquivo mas nenhum teste os usa. Provavelmente são resquícios de refatoração.
// Remova ou mova para o arquivo correto (product_controller_test.go).
type mockProductRepository struct {
	products map[int]*models.Product
}

func (m *mockProductRepository) GetByID(id int) (*models.Product, error) {
	return m.products[id], nil
}

type mockProductUsecase struct{}

func (m *mockProductUsecase) GetProducts(limit, cursor string) ([]*models.Product, string, bool, error) {
	return []*models.Product{}, "", false, nil
}

func TestAuthController_Login_InvalidEmail(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}
	authUsecase := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)
	authController := NewAuthController(authUsecase)

	payload := LoginRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	authController.Login(c)
	//Apenas o status code é verificado.
	//  O body da resposta nunca é validado.
	// Um teste mais robusto verificaria o formato do erro
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAuthController_Login_MissingFields(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	//Isso usa a implementação real do repositório nos testes do controller —
	// que provavelmente precisa de banco de dados. O correto é usar uma interface mockada,
	// igual ao que foi feito com mockUserRepository
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}
	authUsecase := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)
	authController := NewAuthController(authUsecase)

	payload := map[string]string{
		"email": "test@example.com",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	authController.Login(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

type mockUserRepository struct {
	users map[string]*models.User
}

func (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
	return m.users[email], nil
}

func (m *mockUserRepository) Create(user models.User) (int, error) {
	m.users[user.Email] = &user
	return 1, nil
}

// Setup duplicado nos dois testes
//As 6 linhas de setup se repetem identicamente. Extrai um helper:
//gofunc setupAuthController(t *testing.T) *AuthController {
//    t.Helper()
//    mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
//    mockRefreshTokenRepo := &mockRefreshTokenRepository{}
//    authUsecase := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)
//    return NewAuthController(authUsecase)
//}
// Cobertura muito limitada
//Os testes só cobrem validação de input inválido. Faltam os casos mais importantes:

//Login com credenciais corretas → 200 + tokens no cookie/body
//Login com email válido mas usuário inexistente → 401
//Login com senha errada → 401
//Refresh token válido → 200
//Refresh token expirado → 401

//Considere usar table-driven tests (nao sei bem sobre essa parte estudar depois)
//Os dois testes são variações do mesmo cenário. O padrão idiomático em Go:
//gotests := []struct {
//    name           string
//    payload        any
//    expectedStatus int
//}{
//    {"email inválido", LoginRequest{Email: "invalid", Password: "pass123"}, http.StatusBadRequest},
//    {"sem senha", map[string]string{"email": "a@b.com"}, http.StatusBadRequest},
//}

//for _, tt := range tests {
//    t.Run(tt.name, func(t *testing.T) {
//        // ...
//    })
//}
