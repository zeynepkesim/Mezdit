package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"devops-be/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *sql.DB      // Legacy database connection (for gradual migration)
var GormDB *gorm.DB // GORM database connection (new, with auto-migration)

func Init() {
	// PostgreSQL connection string - get from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default local PostgreSQL connection (passwordless postgres user)
		databaseURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	}

	// Initialize legacy database connection (for backwards compatibility)
	if err := initLegacyDB(databaseURL); err != nil {
		log.Printf("Warning: Failed to initialize legacy database: %v", err)
		log.Println("Database-dependent tests will be skipped")
		return
	}

	// Initialize GORM connection with auto-migration
	if err := initGormDB(databaseURL); err != nil {
		log.Printf("Warning: Failed to initialize GORM database: %v", err)
		log.Println("Database-dependent tests will be skipped")
		return
	}

	fmt.Println("✓ Connected to PostgreSQL database")
	fmt.Println("✓ GORM auto-migration completed")
}

// initLegacyDB initializes the legacy sql.DB connection
func initLegacyDB(databaseURL string) error {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	DB = stdlib.OpenDB(*config)

	// Test connection
	if err := DB.Ping(); err != nil {
		DB = nil // Set to nil if connection failed
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}

// initGormDB initializes GORM connection and runs auto-migration
func initGormDB(databaseURL string) error {
	var err error

	// GORM configuration
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Show SQL queries in development
	}

	// Connect with GORM
	GormDB, err = gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database with GORM: %w", err)
	}

	// Run auto-migration
	// This will create tables if they don't exist, and add new columns if needed
	err = GormDB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Image{},
		&models.UserVote{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	fmt.Println("  → Auto-migrated tables: users, posts, images, user_votes")
	return nil
}

func createTables() {
	// Tables are now created automatically via GORM AutoMigrate
	// setup.sql can still be used for seed data (sample users, posts, etc.)
	fmt.Println("Note: Tables are auto-created by GORM. Use setup.sql only for seed data.")
}

func Close() {
	if DB != nil {
		DB.Close()
	}

	if GormDB != nil {
		sqlDB, err := GormDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
