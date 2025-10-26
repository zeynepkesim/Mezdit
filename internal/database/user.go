package database

import (
	"database/sql"
	"devops-be/internal/models"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
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

	log.Printf("Login attempt: User '%s' found, verifying password with bcrypt", username)

	// Compare password with bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
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

	// Hash the password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, fmt.Errorf("failed to create user")
	}

	query := `
		INSERT INTO users (name, username, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, username, password, profile_photo, created_at`

	var user models.User
	var profilePhoto sql.NullString

	err = DB.QueryRow(query, name, username, string(hashedPassword)).Scan(
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

func ChangePassword(userID int, currentPassword string, newPassword string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}

	if currentPassword == "" {
		return fmt.Errorf("current password is required")
	}

	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters")
	}

	if currentPassword == newPassword {
		return fmt.Errorf("new password must be different from current password")
	}

	// Get user to verify current password
	user, err := GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password with bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		log.Printf("Password change failed: Current password mismatch for user %d", userID)
		return fmt.Errorf("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing new password for user %d: %v", userID, err)
		return fmt.Errorf("failed to update password")
	}

	// Update password with hashed version
	query := `UPDATE users SET password = $1 WHERE id = $2`
	_, err = DB.Exec(query, string(hashedPassword), userID)
	if err != nil {
		log.Printf("Failed to update password for user %d: %v", userID, err)
		return fmt.Errorf("failed to update password")
	}

	log.Printf("Password changed successfully for user %d", userID)
	return nil
}