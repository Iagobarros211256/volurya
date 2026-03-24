package controller

import (
	"api/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderUsecase *usecase.OrderUsecase
}

func NewOrderController(ou *usecase.OrderUsecase) *OrderController {
	return &OrderController{orderUsecase: ou}
}

// CreateOrder cria uma ordem e retorna o link de pagamento do PagSeguro
func (oc *OrderController) CreateOrder(c *gin.Context) {
	userID := c.GetInt("user_id") // vindo do middleware JWT

	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	paymentURL, err := oc.orderUsecase.CreateOrder(userID, req.ProductID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Ordem criada com sucesso",
		"payment_url": paymentURL,
	})
}
