package usecase

import (
	"api/models"
	"api/repository"
	"errors"
	"testing"
)

type mockUserRepository struct {
	users map[string]*models.User
}

func (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
	return m.users[email], nil
}

func (m *mockUserRepository) Create(user models.User) (int, error) {
	if _, exists := m.users[user.Email]; exists {
		return 0, errors.New("email already registered")
	}
	m.users[user.Email] = &user
	return 1, nil
}

func TestSignup_ValidInput(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}

	authUsecase := NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)

	_, _, err := authUsecase.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSignup_InvalidEmail(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}

	authUsecase := NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)

	tests := []string{
		"@invalid",
		"invalid@",
		"invalid",
		"@.com",
	}

	for _, email := range tests {
		_, _, err := authUsecase.Signup(email, "validPassword123")
		if err == nil {
			t.Fatalf("expected error for email %s, got nil", email)
		}
	}
}

func TestSignup_ShortPassword(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}

	authUsecase := NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)

	_, _, err := authUsecase.Signup("test@example.com", "short")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	mockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
	mockRefreshTokenRepo := &repository.RefreshTokenRepository{}

	authUsecase := NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)

	_, _, err := authUsecase.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("first signup failed: %v", err)
	}

	_, _, err = authUsecase.Signup("test@example.com", "validPassword456")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}
