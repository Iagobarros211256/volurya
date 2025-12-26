package db

import "testing"

func TestSetupTestDB_Smoke(t *testing.T) {
	db := SetupTestDB()
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("database not reachable: %v", err)
	}
}
