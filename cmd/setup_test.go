package main

import (
	"api/db"
	"testing"
)

func TestSetupRouter(t *testing.T) {
	database := db.SetupTestDB()
	defer database.Close()

	r := SetupRouter(database)

	if r == nil {
		t.Fatal("router is nil")
	}
}
