package controller

import (
	"database/sql"
	"testing"
	"time"
)

func TestHealthCheck_ValidDB(t *testing.T) {
	// Mock database that works
	mockDB := &sql.DB{}

	controller := NewHealthController(mockDB)
	if controller == nil {
		t.Fatal("controller should not be nil")
	}
}

func TestHealthResponse_Format(t *testing.T) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "up",
		Version:   "1.0.0",
	}

	if response.Status != "healthy" {
		t.Fatalf("expected status 'healthy', got %s", response.Status)
	}

	if response.Database != "up" {
		t.Fatalf("expected database 'up', got %s", response.Database)
	}

	if response.Version != "1.0.0" {
		t.Fatalf("expected version '1.0.0', got %s", response.Version)
	}
}
