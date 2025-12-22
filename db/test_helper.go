package db

import "database/sql"

func SetupCleanTestDB() *sql.DB {
	db := SetupTestDB()

	if err := CleanTestDB(db); err != nil {
		panic(err)
	}

	return db
}
