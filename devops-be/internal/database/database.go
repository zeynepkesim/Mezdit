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
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode to prevent [error] logs in tests
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
		&models.Image{},
		&models.Post{},
		&models.UserVote{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	if err := addForeignKeyConstraints(); err != nil {
		log.Printf("Warning: Failed to add foreign key constraints: %v", err)
	}

	fmt.Println("  → Auto-migrated tables: users, posts, images, user_votes")
	fmt.Println("  → Added foreign key constraints")

	// Seed database with sample data if tables are empty
	if err := seedDatabase(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	return nil
}

// addForeignKeyConstraints adds foreign key constraints matching setup.sql
func addForeignKeyConstraints() error {
	// Add foreign key from posts to users
	if err := GormDB.Exec(`
		ALTER TABLE posts
		DROP CONSTRAINT IF EXISTS fk_posts_publisher;

		ALTER TABLE posts
		ADD CONSTRAINT fk_posts_publisher
		FOREIGN KEY (publisher_id) REFERENCES users(id);
	`).Error; err != nil {
		return fmt.Errorf("failed to add posts->users FK: %w", err)
	}

	// Add foreign key from posts to images
	if err := GormDB.Exec(`
		ALTER TABLE posts
		DROP CONSTRAINT IF EXISTS fk_posts_content_photo;

		ALTER TABLE posts
		ADD CONSTRAINT fk_posts_content_photo
		FOREIGN KEY (content_photo_id) REFERENCES images(id);
	`).Error; err != nil {
		return fmt.Errorf("failed to add posts->images FK: %w", err)
	}

	// Add foreign key from images to posts
	if err := GormDB.Exec(`
		ALTER TABLE images
		DROP CONSTRAINT IF EXISTS fk_images_post;

		ALTER TABLE images
		ADD CONSTRAINT fk_images_post
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to add images->posts FK: %w", err)
	}

	// Add foreign keys for user_votes
	if err := GormDB.Exec(`
		ALTER TABLE user_votes
		DROP CONSTRAINT IF EXISTS fk_user_votes_post;

		ALTER TABLE user_votes
		ADD CONSTRAINT fk_user_votes_post
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to add user_votes->posts FK: %w", err)
	}

	if err := GormDB.Exec(`
		ALTER TABLE user_votes
		DROP CONSTRAINT IF EXISTS fk_user_votes_user;

		ALTER TABLE user_votes
		ADD CONSTRAINT fk_user_votes_user
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to add user_votes->users FK: %w", err)
	}

	return nil
}

// seedDatabase populates the database with sample data from setup.sql
func seedDatabase() error {
	// Check if users table is empty
	var userCount int64
	if err := GormDB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Only seed if database is empty
	if userCount > 0 {
		fmt.Println("  → Database already contains data, skipping seed")
		return nil
	}

	fmt.Println("  → Seeding database with sample data...")

	// Sample users with bcrypt hashed passwords (password: "password123")
	users := []models.User{
		{
			Name:         "John Doe",
			Username:     "john_doe",
			Password:     "$2a$10$YyMu9W4FgeVKLsKzRofXIesu4vSVF1Rk8yjfzrF.q4fZNxMsVZSoG",
			ProfilePhoto: stringPtr("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1"),
		},
		{
			Name:         "Jane Smith",
			Username:     "jane_smith",
			Password:     "$2a$10$l1IHmn2gJobbgZtYQ6Qmi.DtsReGVD.udsk4Im3zc1/EIDP4rMI1O",
			ProfilePhoto: stringPtr("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=2"),
		},
		{
			Name:         "Bob Wilson",
			Username:     "bob_wilson",
			Password:     "$2a$10$azP54b5aw6iLZZhm4kHQ/uLuSIyvNNiegCYa5atFjOojEqhFfLI/W",
			ProfilePhoto: stringPtr("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=3"),
		},
	}

	// Insert users
	for _, user := range users {
		if err := GormDB.Create(&user).Error; err != nil {
			log.Printf("Warning: Failed to create user %s: %v", user.Username, err)
		}
	}

	// Sample posts
	posts := []models.Post{
		{
			PublisherID:   1,
			PublisherName: "John Doe",
			Header:        "First Post",
			ContentText:   "Hello world! This is my first post.",
			IsPrivate:     false,
			VoteCount:     5,
		},
		{
			PublisherID:   2,
			PublisherName: "Jane Smith",
			Header:        "Welcome",
			ContentText:   "Great to be here!",
			IsPrivate:     false,
			VoteCount:     3,
		},
		{
			PublisherID:   1,
			PublisherName: "John Doe",
			Header:        "Update",
			ContentText:   "Just had a great day learning DevOps!",
			IsPrivate:     false,
			VoteCount:     8,
		},
		{
			PublisherID:   1,
			PublisherName: "John Doe",
			Header:        "Private Note",
			ContentText:   "This is my private thought.",
			IsPrivate:     true,
			VoteCount:     0,
		},
	}

	// Insert posts
	for _, post := range posts {
		if err := GormDB.Create(&post).Error; err != nil {
			log.Printf("Warning: Failed to create post: %v", err)
		}
	}

	// Sample votes
	votes := []models.UserVote{
		{PostID: 1, UserID: 2, Upvoted: true},
		{PostID: 1, UserID: 3, Upvoted: true},
		{PostID: 2, UserID: 1, Upvoted: true},
		{PostID: 3, UserID: 2, Upvoted: true},
		{PostID: 3, UserID: 3, Upvoted: false},
	}

	// Insert votes
	for _, vote := range votes {
		if err := GormDB.Create(&vote).Error; err != nil {
			log.Printf("Warning: Failed to create vote: %v", err)
		}
	}

	fmt.Println("  → Seeded 3 users, 4 posts, and 5 votes")
	fmt.Println("  → Sample login: username='john_doe', password='password123'")

	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
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
