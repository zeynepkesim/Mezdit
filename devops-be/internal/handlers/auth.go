package handlers

import (
	"log"
	"net/http"
	"strings"

	"devops-be/internal/auth"
	"devops-be/internal/database"
	"devops-be/internal/middleware"
	"devops-be/internal/models"
	"devops-be/internal/security"
)

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var body models.LoginRequest
	if !middleware.DecodeJSONBody(w, r, &body) {
		return
	}

	// Sanitize inputs to prevent XSS and injection attacks
	body.Username = security.SanitizeUsername(body.Username)
	body.Password = strings.TrimSpace(body.Password) // Don't sanitize password, just trim

	// Validate input
	if body.Username == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Username is required")
		return
	}
	if body.Password == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}
	if !security.ValidateLength(body.Username, security.MaxUsernameLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Username is too long")
		return
	}
	if !security.ValidateLength(body.Password, security.MaxPasswordLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Password is too long")
		return
	}

	// Attempt login
	user, err := database.LoginUser(body.Username, body.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			middleware.WriteError(w, http.StatusUnauthorized, "authentication_failed", "Invalid username or password")
			return
		}
		log.Printf("Login error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "An error occurred during authentication")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to generate authentication token")
		return
	}

	// Return user data with token
	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}
	middleware.WriteOK(w, response)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var body models.CreateUserRequest
	if !middleware.DecodeJSONBody(w, r, &body) {
		return
	}

	// Sanitize inputs to prevent XSS attacks
	body.Name = security.SanitizeInput(body.Name)
	body.Username = security.SanitizeUsername(body.Username)
	body.Password = strings.TrimSpace(body.Password) // Don't sanitize password, just trim

	// Validate input
	if body.Name == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if body.Username == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Username is required")
		return
	}
	if body.Password == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}
	if len(body.Password) < 6 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Password must be at least 6 characters")
		return
	}
	// Validate input lengths
	if !security.ValidateLength(body.Name, security.MaxNameLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Name is too long (max 20 characters)")
		return
	}
	if !security.ValidateLength(body.Username, security.MaxUsernameLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Username is too long (max 50 characters)")
		return
	}
	if !security.ValidateLength(body.Password, security.MaxPasswordLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Password is too long (max 32 characters)")
		return
	}

	// Attempt to create user
	user, err := database.CreateUser(body.Name, body.Username, body.Password)
	if err != nil {
		if err.Error() == "username already exists" {
			middleware.WriteError(w, http.StatusConflict, "username_exists", "Username is already taken")
			return
		}
		log.Printf("Registration error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "An error occurred during registration")
		return
	}

	// Generate JWT token for the new user
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to generate authentication token")
		return
	}

	// Return user data with token
	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}
	middleware.WriteCreated(w, response)
}

// ChangePasswordHandler handles password change
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var body models.ChangePasswordRequest
	if !middleware.DecodeJSONBody(w, r, &body) {
		return
	}

	// Trim whitespace (passwords are not sanitized, only trimmed)
	body.CurrentPassword = strings.TrimSpace(body.CurrentPassword)
	body.NewPassword = strings.TrimSpace(body.NewPassword)

	// Validate input
	if body.UserID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "User ID must be a positive integer")
		return
	}
	if body.CurrentPassword == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Current password is required")
		return
	}
	if body.NewPassword == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "New password is required")
		return
	}
	if len(body.NewPassword) < 6 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "New password must be at least 6 characters")
		return
	}
	if !security.ValidateLength(body.CurrentPassword, security.MaxPasswordLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Current password is too long")
		return
	}
	if !security.ValidateLength(body.NewPassword, security.MaxPasswordLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "New password is too long")
		return
	}

	// Attempt to change password
	err := database.ChangePassword(body.UserID, body.CurrentPassword, body.NewPassword)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.WriteError(w, http.StatusNotFound, "user_not_found", "User not found")
			return
		}
		if err.Error() == "current password is incorrect" {
			middleware.WriteError(w, http.StatusUnauthorized, "incorrect_password", "Current password is incorrect")
			return
		}
		if err.Error() == "new password must be different from current password" {
			middleware.WriteError(w, http.StatusBadRequest, "validation_error", "New password must be different from current password")
			return
		}
		if err.Error() == "new password must be at least 6 characters" {
			middleware.WriteError(w, http.StatusBadRequest, "validation_error", "New password must be at least 6 characters")
			return
		}
		log.Printf("Change password error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to change password")
		return
	}

	// Return success response
	response := map[string]string{
		"message": "Password changed successfully",
	}
	middleware.WriteOK(w, response)
}
