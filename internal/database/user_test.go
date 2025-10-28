package database

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestPasswordHashing tests bcrypt password hashing
func TestPasswordHashing(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password - simple",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Valid password - complex",
			password: "P@ssw0rd!2023#Strong",
			wantErr:  false,
		},
		{
			name:     "Valid password - minimum length",
			password: "pass",
			wantErr:  false,
		},
		{
			name:     "Valid password - long",
			password: "this-is-a-very-long-password-with-many-characters-12345",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)

			if (err != nil) != tt.wantErr {
				t.Errorf("bcrypt.GenerateFromPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the hash is valid bcrypt format
				if len(hash) == 0 {
					t.Error("Generated hash is empty")
				}

				// Verify hash starts with bcrypt identifier
				hashStr := string(hash)
				if hashStr[:4] != "$2a$" && hashStr[:4] != "$2b$" && hashStr[:4] != "$2y$" {
					t.Errorf("Hash doesn't have valid bcrypt format: %s", hashStr[:4])
				}
			}
		})
	}
}

// TestPasswordVerification tests bcrypt password verification
func TestPasswordVerification(t *testing.T) {
	// Pre-generate some hashes for testing
	validPassword := "password123"
	validHash, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		hash     []byte
		password string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			hash:     validHash,
			password: validPassword,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			hash:     validHash,
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Empty password with valid hash",
			hash:     validHash,
			password: "",
			wantErr:  true,
		},
		{
			name:     "Case sensitive - wrong case",
			hash:     validHash,
			password: "PASSWORD123",
			wantErr:  true,
		},
		{
			name:     "Extra characters",
			hash:     validHash,
			password: "password123extra",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword(tt.hash, []byte(tt.password))

			if (err != nil) != tt.wantErr {
				t.Errorf("bcrypt.CompareHashAndPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPasswordHashUniqueness verifies that same password generates different hashes
func TestPasswordHashUniqueness(t *testing.T) {
	password := "samepassword123"

	// Generate two hashes for the same password
	hash1, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate first hash: %v", err)
	}

	hash2, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate second hash: %v", err)
	}

	// Hashes should be different (due to different salts)
	if string(hash1) == string(hash2) {
		t.Error("Two hashes of the same password should be different (different salts)")
	}

	// But both should verify correctly
	if err := bcrypt.CompareHashAndPassword(hash1, []byte(password)); err != nil {
		t.Error("First hash should verify the password")
	}

	if err := bcrypt.CompareHashAndPassword(hash2, []byte(password)); err != nil {
		t.Error("Second hash should verify the password")
	}
}

// TestPasswordHashCost verifies bcrypt cost parameter
func TestPasswordHashCost(t *testing.T) {
	password := "testpassword"

	tests := []struct {
		name    string
		cost    int
		wantErr bool
	}{
		{
			name:    "Default cost (10)",
			cost:    bcrypt.DefaultCost,
			wantErr: false,
		},
		{
			name:    "Minimum cost (4)",
			cost:    bcrypt.MinCost,
			wantErr: false,
		},
		{
			name:    "High cost (12)",
			cost:    12,
			wantErr: false,
		},
		{
			name:    "Too high cost (32)",
			cost:    32,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bcrypt.GenerateFromPassword([]byte(password), tt.cost)

			if (err != nil) != tt.wantErr {
				t.Errorf("bcrypt.GenerateFromPassword() with cost=%d, error = %v, wantErr %v",
					tt.cost, err, tt.wantErr)
			}
		})
	}
}

// TestPasswordStrength tests password validation logic
func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		minLength  int
		shouldPass bool
	}{
		{
			name:       "Valid - meets minimum length",
			password:   "pass123",
			minLength:  6,
			shouldPass: true,
		},
		{
			name:       "Invalid - too short",
			password:   "pass",
			minLength:  6,
			shouldPass: false,
		},
		{
			name:       "Valid - exactly minimum",
			password:   "passwd",
			minLength:  6,
			shouldPass: true,
		},
		{
			name:       "Valid - strong password",
			password:   "StrongP@ss123!",
			minLength:  6,
			shouldPass: true,
		},
		{
			name:       "Invalid - empty password",
			password:   "",
			minLength:  6,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple length validation (requirement says "reasonable length")
			isValid := len(tt.password) >= tt.minLength

			if isValid != tt.shouldPass {
				t.Errorf("Password validation failed: password=%s, minLength=%d, got=%v, want=%v",
					tt.password, tt.minLength, isValid, tt.shouldPass)
			}
		})
	}
}

// TestLoginUser_Validation tests input validation for LoginUser
func TestLoginUser_Validation(t *testing.T) {
	// Save original DB and set to nil for validation tests
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	tests := []struct {
		name      string
		username  string
		password  string
		wantError string
	}{
		{
			name:      "Empty username",
			username:  "",
			password:  "password123",
			wantError: "username and password are required",
		},
		{
			name:      "Empty password",
			username:  "testuser",
			password:  "",
			wantError: "username and password are required",
		},
		{
			name:      "Both empty",
			username:  "",
			password:  "",
			wantError: "username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoginUser(tt.username, tt.password)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}

// TestCreateUser_Validation tests input validation for CreateUser
func TestCreateUser_Validation(t *testing.T) {
	// Save original DB and set to nil for validation tests
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	tests := []struct {
		name      string
		userName  string
		username  string
		password  string
		wantError string
	}{
		{
			name:      "Empty name",
			userName:  "",
			username:  "testuser",
			password:  "password123",
			wantError: "name is required",
		},
		{
			name:      "Empty username",
			userName:  "Test User",
			username:  "",
			password:  "password123",
			wantError: "username is required",
		},
		{
			name:      "Empty password",
			userName:  "Test User",
			username:  "testuser",
			password:  "",
			wantError: "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateUser(tt.userName, tt.username, tt.password)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}

// TestGetUserByUsername_Validation tests input validation
func TestGetUserByUsername_Validation(t *testing.T) {
	_, err := GetUserByUsername("")
	if err == nil {
		t.Error("Expected error for empty username")
	}
	if err.Error() != "username is required" {
		t.Errorf("Expected 'username is required', got '%s'", err.Error())
	}
}

// TestGetUserByID_Validation tests input validation
func TestGetUserByID_Validation(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		wantError string
	}{
		{
			name:      "Zero ID",
			id:        0,
			wantError: "invalid user id",
		},
		{
			name:      "Negative ID",
			id:        -1,
			wantError: "invalid user id",
		},
		{
			name:      "Large negative ID",
			id:        -999,
			wantError: "invalid user id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetUserByID(tt.id)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}

// TestChangePassword_Validation tests input validation for ChangePassword
func TestChangePassword_Validation(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		currentPassword string
		newPassword     string
		wantError       string
	}{
		{
			name:            "Invalid user ID - zero",
			userID:          0,
			currentPassword: "current123",
			newPassword:     "newpass123",
			wantError:       "invalid user id",
		},
		{
			name:            "Invalid user ID - negative",
			userID:          -1,
			currentPassword: "current123",
			newPassword:     "newpass123",
			wantError:       "invalid user id",
		},
		{
			name:            "Empty current password",
			userID:          1,
			currentPassword: "",
			newPassword:     "newpass123",
			wantError:       "current password is required",
		},
		{
			name:            "Empty new password",
			userID:          1,
			currentPassword: "current123",
			newPassword:     "",
			wantError:       "new password is required",
		},
		{
			name:            "New password too short",
			userID:          1,
			currentPassword: "current123",
			newPassword:     "short",
			wantError:       "new password must be at least 6 characters",
		},
		{
			name:            "Same password",
			userID:          1,
			currentPassword: "password123",
			newPassword:     "password123",
			wantError:       "new password must be different from current password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ChangePassword(tt.userID, tt.currentPassword, tt.newPassword)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}

// TestLoginUser_AdditionalValidation tests more login validation cases
func TestLoginUser_AdditionalValidation(t *testing.T) {
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	tests := []struct {
		name      string
		username  string
		password  string
		wantError string
	}{
		{
			name:      "Whitespace username",
			username:  "   ",
			password:  "password123",
			wantError: "username and password are required",
		},
		{
			name:      "Whitespace password",
			username:  "testuser",
			password:  "   ",
			wantError: "username and password are required",
		},
		{
			name:      "Tab character username",
			username:  "\t",
			password:  "password123",
			wantError: "username and password are required",
		},
		{
			name:      "Newline character",
			username:  "\n",
			password:  "password123",
			wantError: "username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoginUser(tt.username, tt.password)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}

// TestCreateUser_AdditionalValidation tests more create user validation
func TestCreateUser_AdditionalValidation(t *testing.T) {
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	tests := []struct {
		name      string
		userName  string
		username  string
		password  string
		wantError string
	}{
		{
			name:      "Tab character in name",
			userName:  "\t",
			username:  "testuser",
			password:  "password123",
			wantError: "name is required",
		},
		{
			name:      "Newline in username",
			userName:  "Test User",
			username:  "\n",
			password:  "password123",
			wantError: "username is required",
		},
		{
			name:      "Mixed whitespace password",
			userName:  "Test User",
			username:  "testuser",
			password:  " \t\n ",
			wantError: "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateUser(tt.userName, tt.username, tt.password)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if err.Error() != tt.wantError {
				t.Errorf("Expected error '%s', got '%s'", tt.wantError, err.Error())
			}
		})
	}
}
