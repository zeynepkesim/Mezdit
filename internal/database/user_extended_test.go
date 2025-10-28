package database

import (
	"testing"
)

// TestGetUserByUsername_WhitespaceValidation tests whitespace validation
func TestGetUserByUsername_WhitespaceValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"Whitespace only", "   "},
		{"Tab character", "\t"},
		{"Newline", "\n"},
		{"Mixed whitespace", " \t\n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if DB == nil {
				t.Skip("Skipping database-dependent test")
			}
			
			// These should query database but return not found
			_, err := GetUserByUsername(tt.username)
			
			// Should get an error (user not found or database error)
			if err == nil {
				t.Error("Expected error for whitespace username")
			}
		})
	}
}

// TestGetUserByID_LargeID tests with large positive ID
func TestGetUserByID_LargeID(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping database-dependent test")
	}
	
	// Very large ID that doesn't exist
	_, err := GetUserByID(999999999)
	
	// Should return not found or database error, not panic
	if err == nil {
		t.Error("Expected error for non-existent large ID")
	}
}

// TestChangePassword_EmptyPasswords tests with empty strings
func TestChangePassword_EmptyPasswords(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		current     string
		newPassword string
		wantError   string
	}{
		{
			name:        "Both empty",
			userID:      1,
			current:     "",
			newPassword: "",
			wantError:   "current password is required",
		},
		{
			name:        "Empty current password",
			userID:      1,
			current:     "",
			newPassword: "newpass123",
			wantError:   "current password is required",
		},
		{
			name:        "Empty new password",
			userID:      1,
			current:     "currentpass",
			newPassword: "",
			wantError:   "new password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ChangePassword(tt.userID, tt.current, tt.newPassword)
			
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

// TestLoginUser_InvalidCredentials tests login validation edge cases
func TestLoginUser_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		password  string
		wantError string
	}{
		{
			name:      "Username with only spaces",
			username:  "   ",
			password:  "password123",
			wantError: "username and password are required",
		},
		{
			name:      "Password with only tabs",
			username:  "testuser",
			password:  "\t\t\t",
			wantError: "username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if DB == nil {
				t.Skip("Skipping database-dependent test")
			}
			
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

// TestCreateUser_UsernameEdgeCases tests username validation edge cases
func TestCreateUser_UsernameEdgeCases(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping database-dependent test")
	}

	tests := []struct {
		name     string
		userName string
		username string
		password string
	}{
		{
			name:     "Very long username",
			userName: "Test User",
			username: generateLongString(1000),
			password: "password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateUser(tt.userName, tt.username, tt.password)
			
			// Should handle gracefully (error or success, no panic)
			_ = err // We don't check specific error, just that it doesn't panic
		})
	}
}

// generateLongString helper function to generate a string of specified length
func generateLongString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = 'a'
	}
	return string(bytes)
}
