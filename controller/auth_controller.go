package controller

import (
	"api/usecase"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthController(uc *usecase.AuthUsecase) *AuthController {
	return &AuthController{
		authUsecase: uc,
	}
}

func (a *AuthController) Login(ctx *gin.Context) {

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	token, err := a.authUsecase.Login(body.Email, body.Password)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	ctx.JSON(200, gin.H{
		"access_token": token,
	})
}

func (a *AuthController) Signup(ctx *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	token, err := a.authUsecase.Signup(body.Email, body.Password)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(201, gin.H{
		"message":      "user created",
		"access_token": token,
	})

}
