package controller

import (
	"api/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController depende de struct concreto, não interface
// Isso impede mocks nos testes — por isso o auth_controller_test.go teve que instanciar o usecase real.
type AuthController struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthController(uc *usecase.AuthUsecase) *AuthController {
	return &AuthController{
		authUsecase: uc,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse declarado mas não usado.
// Todos
// os handlers usam gin.H{} diretamente.
// Use o struct ou remova — além de consistência, facilita a serialização e os testes
type AuthResponse struct {
	Message      string `json:"message,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid email or password"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Router /login [post]
func (a *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
		return
	}

	accessToken, refreshToken, err := a.authUsecase.Login(req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	//Tokens retornados no body em vez de cookies HttpOnly.
	// Tokens no body ficam acessíveis via JavaScript,
	// expondo a XSS. O padrão mais seguro é usar cookies HttpOnly.
	//nao faco ideia do que seja isso. marcar para estudar.
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Signup godoc
// @Summary User registration
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignupRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid data or email already registered"
// @Router /signup [post]
func (a *AuthController) Signup(ctx *gin.Context) {
	var req SignupRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
		return
	}

	accessToken, refreshToken, err := a.authUsecase.Signup(req.Email, req.Password)
	if err != nil {
		//Se o usecase retornar um erro interno (ex: falha no banco),
		//  a mensagem vaza para o cliente. O ideal é mapear erros conhecidos
		//Inclusive, email já cadastrado deveria retornar 409 Conflict, não 400 Bad Request.
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":       "user created",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "refresh_token is required"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Router /refresh [post]
func (a *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	accessToken, refreshToken, err := a.authUsecase.RefreshToken(req.RefreshToken)
	if err != nil {
		//Mesmo problema — erros do usecase/banco vazam para o cliente. Mapeie para mensagens genéricas
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout godoc
// @Summary User logout
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token"
// @Success 200 {object} AuthResponse "Logout successful"
// @Failure 400 {object} ErrorResponse "refresh_token is required"
// @Failure 500 {object} ErrorResponse "Failed to logout"
// @Router /logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	var req LogoutRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	if err := a.authUsecase.Logout(req.RefreshToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// LoginRequest e SignupRequest são idênticos
//type LoginRequest struct {
//    Email    string `json:"email" binding:"required,email"`
//    Password string `json:"password" binding:"required,min=8,max=128"`
//}
//
//type SignupRequest struct {
//    Email    string `json:"email" binding:"required,email"`
//    Password string `json:"password" binding:"required,min=8,max=128"`
//}
//Por ora são iguais, mas faz sentido manter separados —
// SignupRequest eventualmente vai ganhar campos como name,
// confirm_password. Só vale unificar se tiver certeza que nunca vão divergir.

//Logout não limpa os cookies
//Se implementar cookies HttpOnly, o logout precisa expirá-los
