package controller

import (
	"api/notifications"
	"api/usecase"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Dependências concretas. Padrão recorrente no projeto todo — substitua por interfaces
type OrderController struct {
	orderUsecase *usecase.OrderUsecase
	hub          *notifications.Hub
}

func NewOrderController(ou *usecase.OrderUsecase, hub *notifications.Hub) *OrderController {
	return &OrderController{orderUsecase: ou, hub: hub}
}

// CreateOrder cria uma ordem e retorna o link de pagamento do PagSeguro
func (oc *OrderController) CreateOrder(c *gin.Context) {
	//Mesmo problema recorrente — sem validação, ID 0 passa silenciosamente para o usecase
	userID := c.GetInt("user_id") // vindo do middleware JWT
	//Mesmo padrão do CartController — nomeie para facilitar reuso e testes
	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		//A API mistura português e inglês nas respostas.
		//  Padronize — APIs REST geralmente usam inglês para facilitar integração com clientes internacionais.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	paymentURL, err := oc.orderUsecase.CreateOrder(userID, req.ProductID, req.Quantity)
	if err != nil {
		if errors.Is(err, usecase.ErrInsufficientStock) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, usecase.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//Erro interno vazando no 500. Erros de banco, queries SQL,
		//  mensagens internas chegam ao cliente. Logue internamente e retorne mensagem genérica
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		//A API mistura português e inglês nas respostas.
		// Padronize — APIs REST geralmente usam inglês para facilitar integração com clientes internacionais.
		"message":     "Ordem criada com sucesso",
		"payment_url": paymentURL,
	})

	if oc.hub != nil {
		oc.hub.Publish(userID, notifications.Event{
			Type:    "order_created",
			Message: "Ordem criada com sucesso",
		})
	}
}

// Notificação após resposta já escrita
//c.JSON(http.StatusCreated, gin.H{...})  // resposta enviada

//if oc.hub != nil {                       // notificação depois
//    oc.hub.Publish(...)
//}
// Mesmo problema do CartController.
// Se Publish bloquear,
// o handler fica preso após já ter respondido. Use goroutine .

//ErrInsufficientStock mapeado para 409 Conflict
// 409 é tecnicamente defensável mas semanticamente
// estranho para estoque insuficiente.
// O mais comum na indústria é 422 Unprocessable Entity — a
// requisição está bem formada, mas não pode ser processada dado o estado atual do recurso.

// Comentário desatualizado
// CreateOrder cria uma ordem e retorna o link de pagamento do PagSeguro
//O projeto usa Stripe, não PagSeguro. Indica que houve uma troca de provedor de pagamento e o comentário não foi atualizado — pode confundir quem mantém o código
