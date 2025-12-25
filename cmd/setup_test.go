package main

import (
	"api/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupRouter_HealthRouteExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database := db.SetupTestDB()
	defer database.Close()

	r := SetupRouter(database)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
}
