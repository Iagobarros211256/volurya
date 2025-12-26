package db

import "testing"

func TestCleanTestDB_ShouldRemoveAllData(t *testing.T) {
	database := SetupTestDB()
	defer database.Close()

	if err := EnsureTablesExist(database); err != nil {
		t.Fatalf("failed to ensure tables: %v", err)
	}

	// Insere dados fake
	_, err := database.Exec(
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3)`,
		"test@test.com",
		"hashed",
		"user",
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Sanity check
	var before int
	_ = database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&before)
	if before == 0 {
		t.Fatal("expected data before clean")
	}

	// Act
	if err := CleanTestDB(database); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	// Assert
	var after int
	err = database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&after)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if after != 0 {
		t.Fatalf("expected 0 users after clean, got %d", after)
	}
}
