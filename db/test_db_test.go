package db

import "testing"

func TestSetupTestDB(t *testing.T) {
	db := SetupTestDB()
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("expected db to be reachable, got error: %v", err)
	}
}
