package controller

import (
	"api/auth"
	"api/usecase"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthController(uc *usecase.AuthUsecase) *AuthController {
	return &AuthController{authUsecase: uc}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

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

	setAuthCookies(ctx, accessToken, refreshToken)
	ctx.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (a *AuthController) Signup(ctx *gin.Context) {
	var req SignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
		return
	}

	accessToken, refreshToken, err := a.authUsecase.Signup(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	setAuthCookies(ctx, accessToken, refreshToken)
	ctx.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (a *AuthController) RefreshToken(ctx *gin.Context) {
	// Lê refresh token do cookie
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	accessToken, newRefreshToken, err := a.authUsecase.RefreshToken(refreshToken)
	if err != nil {
		clearAuthCookies(ctx)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	setAuthCookies(ctx, accessToken, newRefreshToken)
	ctx.JSON(http.StatusOK, gin.H{"message": "token refreshed"})
}

func (a *AuthController) Logout(ctx *gin.Context) {
	// Lê refresh token do cookie para revogar no banco
	refreshToken, _ := ctx.Cookie("refresh_token")
	if refreshToken != "" {
		_ = a.authUsecase.Logout(refreshToken)
	}

	clearAuthCookies(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// setAuthCookies seta access_token e refresh_token como cookies HttpOnly
func setAuthCookies(ctx *gin.Context, accessToken, refreshToken string) {
	secure := os.Getenv("APP_ENV") == "production"

	// Access token — 15 minutos
	ctx.SetCookie(
		"access_token",
		accessToken,
		int((15 * time.Minute).Seconds()),
		"/",
		"",
		secure,
		true, // HttpOnly
	)

	// Refresh token — 7 dias, restrito ao endpoint de refresh
	ctx.SetCookie(
		"refresh_token",
		refreshToken,
		int((7 * 24 * time.Hour).Seconds()),
		"/api/refresh",
		"",
		secure,
		true, // HttpOnly
	)
}

// clearAuthCookies expira os cookies de autenticação
func clearAuthCookies(ctx *gin.Context) {
	ctx.SetCookie("access_token", "", -1, "/", "", false, true)
	ctx.SetCookie("refresh_token", "", -1, "/api/refresh", "", false, true)
}

// Me retorna 200 se o cookie de autenticação for válido
func (a *AuthController) Me(ctx *gin.Context) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := ctx.Get("role")
	ctx.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
}
