package controller

import (
	"api/notifications"
	"api/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Dependências concretas em vez de interfaces
// Impede testes unitários sem instanciar dependências reais. Defina interfaces locais.
type CartController struct {
	cartUsecase  *usecase.CartUsecase
	orderUsecase *usecase.OrderUsecase
	hub          *notifications.Hub
}

func NewCartController(cartUsecase *usecase.CartUsecase, orderUsecase *usecase.OrderUsecase, hub *notifications.Hub) *CartController {
	return &CartController{
		cartUsecase:  cartUsecase,
		orderUsecase: orderUsecase,
		hub:          hub,
	}
}

// GetCart retorna o carrinho do usuário
func (cc *CartController) GetCart(c *gin.Context) {
	userID := c.GetInt("user_id")

	cart, err := cc.cartUsecase.GetCart(userID)
	if err != nil {
		//Erros internos vazando para o cliente. Erros de banco, queries SQL, stack traces podem vazar. Mapeie erros conhecidos
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddItem adiciona um produto ao carrinho
func (cc *CartController) AddItem(c *gin.Context) {
	userID := c.GetInt("user_id")

	var body struct {
		ProductID int `json:"product_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := cc.cartUsecase.AddItem(userID, body.ProductID, body.Quantity)
	if err != nil {
		// Erros internos vazando para o cliente
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// publish chamado após c.JSON / c.Status já escritos.
	//
	//Isso não causa erro visível mas é uma inversão de responsabilidade
	// — a notificação deveria ser disparada antes ou de forma desacoplada.
	// Mais importante: se publish bloquear ou der erro,
	// o handler já respondeu e não há como tratar.
	// Considere disparar em goroutine ou garantir que é sempre non-blocking
	c.JSON(http.StatusCreated, item)
	cc.publish(userID, "cart_item_added", "Produto adicionado ao carrinho")
}

// UpdateItem atualiza a quantidade de um item
func (cc *CartController) UpdateItem(c *gin.Context) {
	// c.GetInt("user_id") retorna 0 silenciosamente.
	//Se o middleware não rodou ou falhou em setar user_id, retorna 0 sem erro.
	// Usuário com ID 0 pode acessar ou modificar dados de outros. Use o helper tipado sugerido anteriormente:
	userID := c.GetInt("user_id")

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	//Structs anônimas nos handlers poderiam ser tipos nomeados
	var body struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := cc.cartUsecase.UpdateItem(userID, itemID, body.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
	cc.publish(userID, "cart_item_updated", "Carrinho atualizado")
}

// RemoveItem remove um item do carrinho
func (cc *CartController) RemoveItem(c *gin.Context) {
	userID := c.GetInt("user_id")

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := cc.cartUsecase.RemoveItem(userID, itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//RemoveItem notifica depois de c.Status(204).204 No Content não tem body,
	// então está correto, mas a notificação ainda ocorre após a resposta.
	// O mesmo ponto do publish em goroutine se aplica aqui.
	c.Status(http.StatusNoContent)
	cc.publish(userID, "cart_item_removed", "Produto removido do carrinho")
}

// Checkout finaliza o carrinho
func (cc *CartController) Checkout(c *gin.Context) {
	userID := c.GetInt("user_id")
	//O cartUsecase recebendo um orderUsecase como parâmetro de
	// método indica acoplamento circular entre usecases.
	// O correto é injetar orderUsecase no cartUsecase via construtor,
	// ou criar um CheckoutUsecase dedicado que orquestra os dois.
	paymentURLs, err := cc.cartUsecase.Checkout(userID, cc.orderUsecase)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "checkout realizado com sucesso",
		"payment_urls": paymentURLs,
	})
	cc.publish(userID, "checkout_created", "Checkout realizado com sucesso")
}

func (cc *CartController) publish(userID int, eventType string, message string) {
	if cc.hub == nil {
		return
	}

	cc.hub.Publish(userID, notifications.Event{
		Type:    eventType,
		Message: message,
	})
}

// Status codes inconsistentes nos erros
//AddItem, UpdateItem, RemoveItem e Checkout todos retornam 400
// para qualquer erro do usecase —
// incluindo produto não encontrado (404),
// estoque insuficiente (422), e erros internos (500).
// O mapeamento correto melhora a experiência do cliente da API.
