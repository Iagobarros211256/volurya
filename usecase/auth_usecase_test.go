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

/*

repository.RefreshTokenRepository{} instanciado diretamente
Mesmo problema do auth_controller_test.go — usa a implementação real que precisa de banco:
gomockRefreshTokenRepo := &repository.RefreshTokenRepository{}
Crie um mock:
gotype mockRefreshTokenRepository struct{}

func (m *mockRefreshTokenRepository) Create(userID int, token string, expiresAt time.Time) error {
    return nil
}

🔴 Setup duplicado em todos os 4 testes
gomockUserRepo := &mockUserRepository{users: make(map[string]*models.User)}
mockRefreshTokenRepo := &repository.RefreshTokenRepository{}
authUsecase := NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo)
Extrai helper:
gofunc setupAuthUsecase(t *testing.T) *AuthUsecase {
    t.Helper()
    return NewAuthUsecase(
        &mockUserRepository{users: make(map[string]*models.User)},
        &mockRefreshTokenRepository{},
    )
}

🟡 TestSignup_InvalidEmail não usa table-driven test corretamente
gofor _, email := range tests {
    _, _, err := authUsecase.Signup(email, "validPassword123")
    if err == nil {
        t.Fatalf("expected error for email %s, got nil", email)
    }
}
t.Fatalf para no primeiro falso — os outros emails não são testados. Use t.Errorf ou t.Run:
gofor _, email := range tests {
    t.Run("email="+email, func(t *testing.T) {
        _, _, err := setupAuthUsecase(t).Signup(email, "validPassword123")
        if err == nil {
            t.Errorf("expected error for email %s", email)
        }
    })
}

🟡 Faltam testes de Login e RefreshToken
Os testes cobrem apenas Signup. Faltam:

Login com credenciais corretas → sucesso
Login com email inexistente → erro
Login com senha errada → erro
RefreshToken válido → novos tokens
RefreshToken expirado → erro
RefreshToken revogado → erro
Logout → token revogado


🟡 mockUserRepository.GetByEmail retorna nil, nil para não encontrado
gofunc (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
    return m.users[email], nil
}
Se o email não existe, retorna nil, nil — propaga o anti-padrão do repository real. O mock deveria retornar ErrUserNotFound quando não encontrado, para testar o comportamento correto do usecase.

🟢 Senha não verificada no mock de Create
gofunc (m *mockUserRepository) Create(user models.User) (int, error) {
    m.users[user.Email] = &user
    return 1, nil
}
O mock armazena a senha em texto puro. Se o usecase esquecer de hashear a senha antes de chamar Create, o teste passa mesmo assim. Adicione verificação:
goif !strings.HasPrefix(user.Password, "$2a$") {
    return 0, errors.New("password not hashed")
}


*/
