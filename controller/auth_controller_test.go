package controller

import (
	"api/repository"
	"api/usecase"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api/models"

	"github.com/gin-gonic/gin"
)

// --- Mocks ---

type mockUserRepo struct {
	users map[string]*models.User
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) Create(user models.User) (int, error) {
	if _, exists := m.users[user.Email]; exists {
		return 0, errors.New("email already registered")
	}
	user.ID = len(m.users) + 1
	m.users[user.Email] = &user
	return user.ID, nil
}

type mockRefreshTokenRepo struct {
	tokens map[string]*models.RefreshToken
}

func newMockRefreshRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{tokens: make(map[string]*models.RefreshToken)}
}

func (m *mockRefreshTokenRepo) Create(userID int, token string, expiresAt time.Time) error {
	m.tokens[token] = &models.RefreshToken{UserID: userID, Token: token, ExpiresAt: expiresAt}
	return nil
}

func (m *mockRefreshTokenRepo) GetByToken(token string) (*models.RefreshToken, error) {
	rt, ok := m.tokens[token]
	if !ok {
		return nil, nil
	}
	return rt, nil
}

func (m *mockRefreshTokenRepo) Revoke(token string) error {
	if rt, ok := m.tokens[token]; ok {
		rt.Revoked = true
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByUser(userID int) error {
	for _, rt := range m.tokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired() (int64, error) {
	var count int64
	for key, rt := range m.tokens {
		if rt.Revoked || time.Now().After(rt.ExpiresAt) {
			delete(m.tokens, key)
			count++
		}
	}
	return count, nil
}

// --- Helpers ---

func setupAuthController() *AuthController {
	userRepo := &mockUserRepo{users: make(map[string]*models.User)}
	refreshRepo := newMockRefreshRepo()
	uc := usecase.NewAuthUsecase(userRepo, refreshRepo)
	return NewAuthController(uc)
}

func setupAuthControllerWithUser(email, password string) (*AuthController, error) {
	userRepo := &mockUserRepo{users: make(map[string]*models.User)}
	refreshRepo := newMockRefreshRepo()
	uc := usecase.NewAuthUsecase(userRepo, refreshRepo)
	_, _, err := uc.Signup(email, password)
	if err != nil {
		return nil, err
	}
	return NewAuthController(uc), nil
}

func performRequest(controller *AuthController, method, path string, body any, handler func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

// --- Testes de Login ---

func TestLogin_InvalidEmail(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/login",
		map[string]string{"email": "not-an-email", "password": "password123"},
		ctrl.Login,
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/login",
		map[string]string{"email": "test@example.com"},
		ctrl.Login,
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/login",
		LoginRequest{Email: "nobody@example.com", Password: "password123"},
		ctrl.Login,
	)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl, err := setupAuthControllerWithUser("test@example.com", "correctPassword123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	w := performRequest(ctrl, http.MethodPost, "/api/login",
		LoginRequest{Email: "test@example.com", Password: "wrongPassword"},
		ctrl.Login,
	)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	ctrl, err := setupAuthControllerWithUser("test@example.com", "correctPassword123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	w := performRequest(ctrl, http.MethodPost, "/api/login",
		LoginRequest{Email: "test@example.com", Password: "correctPassword123"},
		ctrl.Login,
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	// Verifica cookie HttpOnly
	cookies := w.Result().Cookies()
	var hasAccessToken bool
	for _, c := range cookies {
		if c.Name == "access_token" && c.HttpOnly {
			hasAccessToken = true
		}
	}
	if !hasAccessToken {
		t.Fatal("expected HttpOnly access_token cookie")
	}
	// Verifica body
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] != "login successful" {
		t.Fatalf("unexpected message: %s", resp["message"])
	}
}

// --- Testes de Signup ---

func TestSignup_ValidInput(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/signup",
		SignupRequest{Email: "new@example.com", Password: "password123"},
		ctrl.Signup,
	)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestSignup_InvalidEmail(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/signup",
		SignupRequest{Email: "not-an-email", Password: "password123"},
		ctrl.Signup,
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	ctrl, err := setupAuthControllerWithUser("existing@example.com", "password123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	w := performRequest(ctrl, http.MethodPost, "/api/signup",
		SignupRequest{Email: "existing@example.com", Password: "password123"},
		ctrl.Signup,
	)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestSignup_PasswordTooShort(t *testing.T) {
	ctrl := setupAuthController()
	w := performRequest(ctrl, http.MethodPost, "/api/signup",
		SignupRequest{Email: "test@example.com", Password: "short"},
		ctrl.Signup,
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// Garante que o repositório concreto implementa a interface
var _ repository.RefreshTokenRepositoryInterface = (*mockRefreshTokenRepo)(nil)
