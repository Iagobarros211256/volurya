package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Event struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type client struct {
	userID int
	ch     chan Event
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int]map[*client]struct{}
	closed  bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[*client]struct{}),
	}
}

func (h *Hub) Subscribe(ctx context.Context, userID int) <-chan Event {
	c := &client{
		userID: userID,
		ch:     make(chan Event, 8),
	}

	h.mu.Lock()
	if h.closed {
		close(c.ch)
		h.mu.Unlock()
		return c.ch
	}

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*client]struct{})
	}
	h.clients[userID][c] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.unsubscribe(c)
	}()

	return c.ch
}

func (h *Hub) Publish(userID int, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return
	}

	for c := range h.clients[userID] {
		select {
		case c.ch <- event:
		default:
			slog.Warn("dropping notification for slow SSE client", "user_id", userID, "event_type", event.Type)
		}
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	h.closed = true
	for _, userClients := range h.clients {
		for c := range userClients {
			close(c.ch)
		}
	}
	h.clients = make(map[int]map[*client]struct{})
}

func (h *Hub) unsubscribe(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients := h.clients[c.userID]
	if userClients == nil {
		return
	}

	delete(userClients, c)
	close(c.ch)

	if len(userClients) == 0 {
		delete(h.clients, c.userID)
	}
}

func Stream(c *gin.Context, events <-chan Event) {
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	writeEvent(writer, "connected", Event{
		Type:    "connected",
		Message: "Conectado às notificações em tempo real",
	})
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			writeEvent(writer, event.Type, event)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(writer, ": ping\n\n")
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func writeEvent(writer http.ResponseWriter, eventName string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal SSE event", "error", err)
		return
	}

	fmt.Fprintf(writer, "event: %s\n", eventName)
	fmt.Fprintf(writer, "data: %s\n\n", data)
}
