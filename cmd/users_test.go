package main

import (
	"api/db"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUser_ShouldPersistUserAndReturn201(t *testing.T) {
	database := db.SetupTestDB()
	defer database.Close()

	if err := db.CleanTestDB(database); err != nil {
		t.Fatalf("failed to clean db: %v", err)
	}

	if err := db.EnsureTablesExist(database); err != nil {
		t.Fatalf("failed to ensure tables: %v", err)
	}

	r := SetupRouter(database)

	payload := map[string]string{
		"name":  "Iago",
		"email": "iago@test.com",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}

	if resp.Body.Len() == 0 {
		t.Fatal("expected response body, got empty")
	}

	var count int
	err = database.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email = $1",
		"iago@test.com",
	).Scan(&count)

	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}
}
