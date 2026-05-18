package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	db *sql.DB
}

func NewHealthController(db *sql.DB) *HealthController {
	return &HealthController{db: db}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
	Version   string `json:"version"`
}

// Health godoc
// @Summary Health check
// @Description Check API and database health
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthController) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dbStatus := "down"
	if err := h.db.PingContext(ctx); err == nil {
		dbStatus = "up"
	}

	statusCode := http.StatusOK
	status := "healthy"

	if dbStatus != "up" {
		statusCode = http.StatusServiceUnavailable
		status = "unhealthy"
	}

	c.JSON(statusCode, HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  dbStatus,
		Version:   "1.0.0",
	})
}
