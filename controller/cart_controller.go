package controller

import (
	"api/notifications"
	"api/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
	cc.publish(userID, "cart_item_added", "Produto adicionado ao carrinho")
}

// UpdateItem atualiza a quantidade de um item
func (cc *CartController) UpdateItem(c *gin.Context) {
	userID := c.GetInt("user_id")

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

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

	c.Status(http.StatusNoContent)
	cc.publish(userID, "cart_item_removed", "Produto removido do carrinho")
}

// Checkout finaliza o carrinho
func (cc *CartController) Checkout(c *gin.Context) {
	userID := c.GetInt("user_id")

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
