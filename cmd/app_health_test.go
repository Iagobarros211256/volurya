package main

import (
	"api/db"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Health endpoint should return 200 and indicate that the app is running
func TestHealthEndpoint(t *testing.T) {
	database := db.SetupTestDB()
	defer database.Close()

	r := SetupRouter(database)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Assert status code
	if resp.Code != http.StatusOK {
		t.Fatalf(
			"GET /health: expected status %d, got %d",
			http.StatusOK,
			resp.Code,
		)
	}

	// Assert response body (minimal semantic check)
	body := resp.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatalf("GET /health: expected non-empty response body")
	}
}
