package database

import (
	"database/sql"
	"devops-be/internal/models"
	"fmt"
	"log"
)

func LoginUser(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE username = $1`

	var user models.User
	var profilePhoto sql.NullString

	err := DB.QueryRow(query, username).Scan(
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

func CreateUser(name, username, password string) (*models.User, error) {
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

	var user models.User
	var profilePhoto sql.NullString

	err := DB.QueryRow(query, name, username, password).Scan(
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
	_, err = DB.Exec(updateQuery, diceBearURL, user.ID)
	if err != nil {
		log.Printf("Warning: Failed to set profile photo for user %d: %v", user.ID, err)
	} else {
		user.ProfilePhoto = &diceBearURL
	}

	return &user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE username = $1`

	var user models.User
	var profilePhoto sql.NullString

	err := DB.QueryRow(query, username).Scan(
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

func GetUserByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	query := `
		SELECT id, name, username, password, profile_photo, created_at
		FROM users WHERE id = $1`

	var user models.User
	var profilePhoto sql.NullString

	err := DB.QueryRow(query, id).Scan(
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