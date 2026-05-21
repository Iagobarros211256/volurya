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

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAuthController_Login_MissingFields(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
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
