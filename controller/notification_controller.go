package controller

import (
	"api/notifications"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	//hub *notifications.Hub concreto
	//Mesmo padrão recorrente no projeto. Defina uma interface
	hub *notifications.Hub
}

func NewNotificationController(hub *notifications.Hub) *NotificationController {
	return &NotificationController{hub: hub}
}

func (nc *NotificationController) Stream(c *gin.Context) {
	userID := c.GetInt("user_id")
	//userID == 0 é uma validação frágil
	//Depende de uma convenção implícita de que 0 nunca é um ID válido.
	// Se o banco usar IDs sequenciais começando em 1 isso funciona,
	// mas é uma suposição não documentada. O helper tipado sugerido anteriormente seria mais explícito
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	events := nc.hub.Subscribe(c.Request.Context(), userID)
	notifications.Stream(c, events)
}

// Sem headers SSE explícitos no controller
//Os headers Content-Type: text/event-stream,
//  Cache-Control: no-cache e Connection: keep-alive
// provavelmente estão sendo setados dentro de
// notifications.Stream. Isso está OK se for
// sempre SSE, mas vale confirmar —
// se notifications.Stream não setar esses headers, clientes SSE vão ter comportamento indefinido.

// Sem limite de conexões por usuário
//Um usuário poderia abrir múltiplas conexões SSE simultâneas. Considere limitar no hub.Subscribe:
// No Hub: se já existe uma subscription para esse userID, fechar a anterior
// ou retornar erro se exceder N conexões

//Sem log de conexão/desconexão
//Para debugging em produção, útil saber quando usuários conectam e desconectam do stream:
//goslog.Info("SSE client connected", "user_id", userID)
//defer slog.Info("SSE client disconnected", "user_id", userID)
