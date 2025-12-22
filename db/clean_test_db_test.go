package db

import (
	"testing"
)

func TestCleanTestDB(t *testing.T) {
	database := SetupTestDB()
	defer database.Close()

	if err := CleanTestDB(database); err != nil {
		t.Fatalf("clean failed: %v", err)
	}
}
