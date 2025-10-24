package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Init() {
	// PostgreSQL connection string - get from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default local PostgreSQL connection (passwordless postgres user)
		databaseURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	}

	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("Failed to parse database config:", err)
	}

	DB = stdlib.OpenDB(*config)

	// Test connection
	if err := DB.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to PostgreSQL database")
	createTables()
}

func createTables() {
	// Tables are created via setup.sql file
	// This function can be used for additional setup if needed
	fmt.Println("Database tables should be created via setup.sql")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}