package controller

import (
	"api/notifications"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	hub *notifications.Hub
}

func NewNotificationController(hub *notifications.Hub) *NotificationController {
	return &NotificationController{hub: hub}
}

func (nc *NotificationController) Stream(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	events := nc.hub.Subscribe(c.Request.Context(), userID)
	notifications.Stream(c, events)
}
