package usecase

import (
	"api/models"
	"errors"
	"testing"
	"time"
)

// mockUserRepository implementa repository.UserRepositoryInterface
type mockUserRepository struct {
	users map[string]*models.User
}

func (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepository) Create(user models.User) (int, error) {
	if _, exists := m.users[user.Email]; exists {
		return 0, errors.New("email already registered")
	}
	user.ID = len(m.users) + 1
	m.users[user.Email] = &user
	return user.ID, nil
}

// mockRefreshTokenRepository implementa repository.RefreshTokenRepositoryInterface
type mockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepository {
	return &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *mockRefreshTokenRepository) Create(userID int, token string, expiresAt time.Time) error {
	m.tokens[token] = &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}
	return nil
}

func (m *mockRefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	rt, ok := m.tokens[token]
	if !ok {
		return nil, nil
	}
	return rt, nil
}

func (m *mockRefreshTokenRepository) Revoke(token string) error {
	rt, ok := m.tokens[token]
	if !ok {
		return nil
	}
	rt.Revoked = true
	return nil
}

func (m *mockRefreshTokenRepository) RevokeAllByUser(userID int) error {
	for _, rt := range m.tokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

func newTestAuthUsecase() *AuthUsecase {
	return NewAuthUsecase(
		&mockUserRepository{users: make(map[string]*models.User)},
		newMockRefreshTokenRepo(),
	)
}

func TestSignup_ValidInput(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSignup_InvalidEmail(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("not-an-email", "validPassword123")
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
}

func TestSignup_PasswordTooShort(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("test@example.com", "short")
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("first signup failed: %v", err)
	}
	_, _, err = uc.Signup("test@example.com", "validPassword123")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	_, _, err = uc.Login("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	_, _, err = uc.Login("test@example.com", "wrongPassword")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	uc := newTestAuthUsecase()
	_, _, err := uc.Login("nobody@example.com", "validPassword123")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestLogout_RevokesToken(t *testing.T) {
	uc := newTestAuthUsecase()
	_, refreshToken, err := uc.Signup("test@example.com", "validPassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	err = uc.Logout(refreshToken)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
}
