package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Password     string    `json:"-"` // Don't expose password in JSON
	ProfilePhoto *string   `json:"profilePhoto"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Image struct {
	ID        int       `json:"id"`
	PostID    *int      `json:"postId"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
}

type Post struct {
	ID               int       `json:"id"`
	PublisherID      int       `json:"publisherId"`
	PublisherName    string    `json:"publisherName"`    // From posts table
	PublisherPhoto   *string   `json:"publisherPhoto"`   // JOIN from users table
	ContentText      string    `json:"contentText"`
	ContentPhotoID   *int      `json:"contentPhotoId"`
	ContentPhotoPath *string   `json:"contentPhotoPath"` // JOIN from images table
	VoteCount        int       `json:"voteCount"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Vote struct {
	PostID   int  `json:"postId"`
	UserID   int  `json:"userId"`
	Upvoted  bool `json:"upvoted"`
}

var db *sql.DB

func initDB() {
	// PostgreSQL bağlantı string'i - environment variable'dan al
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default local PostgreSQL bağlantısı (şifresiz postgres user)
		databaseURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	}

	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("Failed to parse database config:", err)
	}

	db = stdlib.OpenDB(*config)
	
	// Test connection
	if err := db.Ping(); err != nil {
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

// Auth operations
func LoginUser(username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE username = $1`

	var user User
	var profilePhoto sql.NullString
	
	err := db.QueryRow(query, username).Scan(
		&user.ID, &user.Name, &user.Username, &user.Password, 
		&profilePhoto, &user.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Login attempt: User '%s' not found", username)
			return nil, fmt.Errorf("invalid username or password")
		}
		log.Printf("Database error during login for user '%s': %v", username, err)
		return nil, fmt.Errorf("database error occurred")
	}

	log.Printf("Login attempt: User '%s' found, checking password", username)
	log.Printf("Debug: Provided password='%s', Stored password='%s'", password, user.Password)
	
	// Simple password check (in real app, use bcrypt hash comparison)
	if user.Password != password {
		log.Printf("Login failed: Password mismatch for user '%s'", username)
		return nil, fmt.Errorf("invalid username or password")
	}

	log.Printf("Login successful for user '%s'", username)
	
	if profilePhoto.Valid {
		user.ProfilePhoto = &profilePhoto.String
	}

	return &user, nil
}

// User operations
func CreateUser(name, username, password string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Check if username already exists
	existingUser, _ := GetUserByUsername(username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	query := `
		INSERT INTO users (name, username, password) 
		VALUES ($1, $2, $3) 
		RETURNING id, name, username, password, profile_photo, created_at`

	var user User
	var profilePhoto sql.NullString
	
	err := db.QueryRow(query, name, username, password).Scan(
		&user.ID, &user.Name, &user.Username, &user.Password, 
		&profilePhoto, &user.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	// Generate DiceBear profile photo URL
	diceBearURL := fmt.Sprintf("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=%d", user.ID)
	
	// Update user with profile photo
	updateQuery := `UPDATE users SET profile_photo = $1 WHERE id = $2`
	_, err = db.Exec(updateQuery, diceBearURL, user.ID)
	if err != nil {
		log.Printf("Warning: Failed to set profile photo for user %d: %v", user.ID, err)
	} else {
		user.ProfilePhoto = &diceBearURL
	}

	return &user, nil
}

func GetUserByUsername(username string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE username = $1`

	var user User
	var profilePhoto sql.NullString
	
	err := db.QueryRow(query, username).Scan(
		&user.ID, &user.Name, &user.Username, &user.Password, 
		&profilePhoto, &user.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error occurred")
	}

	if profilePhoto.Valid {
		user.ProfilePhoto = &profilePhoto.String
	}

	return &user, nil
}

func GetUserByID(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE id = $1`

	var user User
	var profilePhoto sql.NullString
	
	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Username, &user.Password, 
		&profilePhoto, &user.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error occurred")
	}

	if profilePhoto.Valid {
		user.ProfilePhoto = &profilePhoto.String
	}

	return &user, nil
}

// Post operations
func AddPost(publisherID int, publisherName string, contentText string, contentPhotoID *int) (*Post, error) {
	if publisherID <= 0 {
		return nil, fmt.Errorf("invalid publisher id")
	}
	
	// If publisher name is empty, set to Anonymous
	if publisherName == "" {
		publisherName = "Anonymous"
	}
	
	if contentText == "" {
		return nil, fmt.Errorf("content text cannot be empty")
	}

	// Check if publisher exists
	user, err := GetUserByID(publisherID)
	if err != nil {
		return nil, fmt.Errorf("publisher not found")
	}

	query := `
		INSERT INTO posts (publisher_id, publisher_name, content_text, content_photo_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, publisher_id, publisher_name, content_text, content_photo_id, vote_count, created_at`

	var post Post
	var contentPhotoIDNull sql.NullInt32

	err = db.QueryRow(query, publisherID, publisherName, contentText, contentPhotoID).Scan(
		&post.ID, &post.PublisherID, &post.PublisherName, &post.ContentText, 
		&contentPhotoIDNull, &post.VoteCount, &post.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create post")
	}

	if contentPhotoIDNull.Valid {
		contentPhotoIDInt := int(contentPhotoIDNull.Int32)
		post.ContentPhotoID = &contentPhotoIDInt
	}

	// Get publisher photo from user
	if user.ProfilePhoto != nil {
		post.PublisherPhoto = user.ProfilePhoto
	}

	return &post, nil
}

func VotePost(postID int, userID int, upvoted bool) (*Post, error) {
	if postID <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}
	
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction")
	}
	defer tx.Rollback()

	// Check if post exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", postID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error occurred")
	}
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	// Check if user exists
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error occurred")
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user has already voted
	var previousVote sql.NullBool
	err = tx.QueryRow(
		"SELECT upvoted FROM user_votes WHERE post_id = $1 AND user_id = $2",
		postID, userID,
	).Scan(&previousVote)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("database error occurred")
	}

	// Calculate vote count change
	var voteCountDelta int = 0

	if previousVote.Valid {
		// User has voted before
		if previousVote.Bool == upvoted {
			// Same vote, no change needed
		} else {
			// Different vote, cancel previous and apply new
			if previousVote.Bool {
				voteCountDelta = -2 // Remove upvote, add downvote
			} else {
				voteCountDelta = 2 // Remove downvote, add upvote
			}
		}
		
		// Update existing vote
		_, err = tx.Exec(
			"UPDATE user_votes SET upvoted = $1 WHERE post_id = $2 AND user_id = $3",
			upvoted, postID, userID,
		)
	} else {
		// New vote
		if upvoted {
			voteCountDelta = 1
		} else {
			voteCountDelta = -1
		}
		
		// Insert new vote
		_, err = tx.Exec(
			"INSERT INTO user_votes (post_id, user_id, upvoted) VALUES ($1, $2, $3)",
			postID, userID, upvoted,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to record vote")
	}

	// Update post vote count
	if voteCountDelta != 0 {
		_, err = tx.Exec(
			"UPDATE posts SET vote_count = vote_count + $1 WHERE id = $2",
			voteCountDelta, postID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update vote count")
		}
	}

	// Get updated post with JOIN to get publisher photo and content photo path
	var post Post
	var contentPhotoIDNull sql.NullInt32
	var contentPhotoPath sql.NullString
	var publisherPhoto sql.NullString

	query := `
		SELECT p.id, p.publisher_id, p.publisher_name, p.content_text, 
			   p.content_photo_id, p.vote_count, p.created_at,
			   u.profile_photo as publisher_photo,
			   i.path as content_photo_path
		FROM posts p
		JOIN users u ON p.publisher_id = u.id
		LEFT JOIN images i ON p.content_photo_id = i.id
		WHERE p.id = $1`

	err = tx.QueryRow(query, postID).Scan(
		&post.ID, &post.PublisherID, &post.PublisherName, &post.ContentText,
		&contentPhotoIDNull, &post.VoteCount, &post.CreatedAt,
		&publisherPhoto, &contentPhotoPath,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated post")
	}

	if publisherPhoto.Valid {
		post.PublisherPhoto = &publisherPhoto.String
	}
	if contentPhotoIDNull.Valid {
		contentPhotoIDInt := int(contentPhotoIDNull.Int32)
		post.ContentPhotoID = &contentPhotoIDInt
	}
	if contentPhotoPath.Valid {
		post.ContentPhotoPath = &contentPhotoPath.String
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to save changes")
	}

	return &post, nil
}

func GetPosts() ([]Post, error) {
	query := `
		SELECT p.id, p.publisher_id, p.publisher_name, p.content_text, 
			   p.content_photo_id, p.vote_count, p.created_at,
			   u.profile_photo as publisher_photo,
			   i.path as content_photo_path
		FROM posts p
		JOIN users u ON p.publisher_id = u.id
		LEFT JOIN images i ON p.content_photo_id = i.id
		ORDER BY p.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts")
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var publisherPhoto sql.NullString
		var contentPhotoIDNull sql.NullInt32
		var contentPhotoPath sql.NullString

		err := rows.Scan(
			&post.ID, &post.PublisherID, &post.PublisherName, &post.ContentText,
			&contentPhotoIDNull, &post.VoteCount, &post.CreatedAt,
			&publisherPhoto, &contentPhotoPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to process post data")
		}

		if publisherPhoto.Valid {
			post.PublisherPhoto = &publisherPhoto.String
		}
		if contentPhotoIDNull.Valid {
			contentPhotoIDInt := int(contentPhotoIDNull.Int32)
			post.ContentPhotoID = &contentPhotoIDInt
		}
		if contentPhotoPath.Valid {
			post.ContentPhotoPath = &contentPhotoPath.String
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while reading posts")
	}

	if posts == nil {
		return []Post{}, nil
	}

	return posts, nil
}
