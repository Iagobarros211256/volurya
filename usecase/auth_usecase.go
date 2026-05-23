package usecase

import (
	"api/auth"
	"api/config"
	"api/metrics"
	"api/models"
	"api/repository"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo         repository.UserRepositoryInterface
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewAuthUsecase(userRepo repository.UserRepositoryInterface, refreshTokenRepo *repository.RefreshTokenRepository) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (a *AuthUsecase) Signup(email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", errors.New("invalid email format")
	}
	if len(password) < 8 || len(password) > 128 {
		return "", "", errors.New("password must be between 8 and 128 characters")
	}

	existing, err := a.userRepo.GetByEmail(email)
	if err != nil {
		return "", "", err
	}
	if existing != nil {
		return "", "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	user := models.User{
		Email:    email,
		Password: string(hash),
		Role:     "user",
	}

	id, err := a.userRepo.Create(user)
	if err != nil {
		return "", "", err
	}

	metrics.UsersTotal.Inc()

	accessToken, err := auth.GenerateToken(id, user.Role, config.GetAccessTokenDuration())
	if err != nil {
		return "", "", err
	}

	refreshToken, expiresAt, err := auth.GenerateRefreshToken(id, user.Role, config.GetRefreshTokenDuration())
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(id, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *AuthUsecase) Login(email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := a.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return "", "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role, config.GetAccessTokenDuration())
	if err != nil {
		return "", "", err
	}

	refreshToken, expiresAt, err := auth.GenerateRefreshToken(user.ID, user.Role, config.GetRefreshTokenDuration())
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(user.ID, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *AuthUsecase) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	rt, err := a.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil || rt == nil {
		return "", "", errors.New("refresh token not found")
	}

	if rt.Revoked {
		return "", "", errors.New("refresh token revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	if err := a.refreshTokenRepo.Revoke(refreshToken); err != nil {
		return "", "", err
	}

	newAccessToken, err := auth.GenerateToken(claims.UserID, claims.Role, config.GetAccessTokenDuration())
	if err != nil {
		return "", "", err
	}

	newRefreshToken, expiresAt, err := auth.GenerateRefreshToken(claims.UserID, claims.Role, config.GetRefreshTokenDuration())
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(claims.UserID, newRefreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *AuthUsecase) Logout(refreshToken string) error {
	return a.refreshTokenRepo.Revoke(refreshToken)
}

/*


 refreshTokenRepo *repository.RefreshTokenRepository concreto
gotype AuthUsecase struct {
    userRepo         repository.UserRepositoryInterface  // interface ✅
    refreshTokenRepo *repository.RefreshTokenRepository  // concreto ❌
}
Assimetria clara — userRepo tem interface mas refreshTokenRepo não. É por isso que os testes instanciam o repository real. Defina a interface:
gotype RefreshTokenRepository interface {
    Create(userID int, token string, expiresAt time.Time) error
    GetByToken(token string) (*models.RefreshToken, error)
    Revoke(token string) error
}

🔴 Signup verifica email duplicado com SELECT antes do INSERT — race condition
goexisting, err := a.userRepo.GetByEmail(email)
if existing != nil {
    return "", "", errors.New("email already registered")
}
// outro goroutine pode inserir aqui
a.userRepo.Create(user)
Dois signups simultâneos com o mesmo email passam pelo check e ambos tentam o INSERT. O banco vai rejeitar o segundo com violação de UNIQUE, mas o erro vai vazar como erro interno em vez de "email already registered". Confie na constraint do banco:
goid, err := a.userRepo.Create(user)
if err != nil {
    if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
        return "", "", errors.New("email already registered")
    }
    return "", "", err
}

🔴 RefreshToken valida o JWT antes de verificar no banco
goclaims, err := auth.ValidateToken(refreshToken)  // valida JWT
// ...
rt, err := a.refreshTokenRepo.GetByToken(refreshToken)  // verifica no banco
A ordem está invertida. Se o JWT for válido mas o token já foi revogado no banco, o JWT é validado desnecessariamente. Verifique no banco primeiro:
gort, err := a.refreshTokenRepo.GetByToken(refreshToken)
if err != nil || rt == nil {
    return "", "", errors.New("invalid refresh token")
}
if rt.Revoked || time.Now().After(rt.ExpiresAt) {
    return "", "", errors.New("invalid refresh token")
}
// só então valida o JWT
claims, err := auth.ValidateToken(refreshToken)

🟡 Login não limita refresh tokens por usuário
Cada login cria um novo refresh token sem revogar os anteriores. Um usuário com 1000 logins tem 1000 refresh tokens ativos no banco. Considere revogar tokens anteriores no login ou limitar por quantidade.

🟡 Erros internos sem wrap
goreturn "", "", err  // bcrypt, GenerateToken, refreshTokenRepo.Create
Erros de infraestrutura passam direto sem contexto. Em produção é difícil saber qual operação falhou:
goif err != nil {
    return "", "", fmt.Errorf("failed to generate access token: %w", err)
}

🟡 Logout não valida o token antes de revogar
gofunc (a *AuthUsecase) Logout(refreshToken string) error {
    return a.refreshTokenRepo.Revoke(refreshToken)
}
Qualquer string é aceita como refresh token para logout — sem verificar se pertence ao usuário autenticado. Um usuário poderia tentar revogar tokens de outros usuários se souber o valor.

🟡 Role hardcoded como "user" no Signup
gouser := models.User{
    Role: "user",
}
Correto para o caso padrão, mas deveria usar a constante:
goRole: string(models.UserRoleUser),

🟢 metrics.UsersTotal.Inc() só no Signup, não no Login
Já apontado em metrics.go — Login não incrementa nenhuma métrica. Adicione pelo menos um counter de logins para monitoramento de autenticação.

*/
