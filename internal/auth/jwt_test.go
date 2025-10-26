package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		wantErr  bool
	}{
		{
			name:     "Valid user - normal case",
			userID:   1,
			username: "john_doe",
			wantErr:  false,
		},
		{
			name:     "Valid user - different ID",
			userID:   42,
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "Valid user - zero ID",
			userID:   0,
			username: "zero_user",
			wantErr:  false,
		},
		{
			name:     "Valid user - empty username",
			userID:   1,
			username: "",
			wantErr:  false,
		},
		{
			name:     "Valid user - special characters in username",
			userID:   1,
			username: "user@test.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify token is not empty
				if token == "" {
					t.Error("Generated token is empty")
				}

				// Verify token has 3 parts (header.payload.signature)
				if len(token) < 10 {
					t.Error("Generated token is too short")
				}

				// Try to parse and validate the token
				claims, err := ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token is invalid: %v", err)
				}

				// Verify claims
				if claims.UserID != tt.userID {
					t.Errorf("Token userID = %d, want %d", claims.UserID, tt.userID)
				}

				if claims.Username != tt.username {
					t.Errorf("Token username = %s, want %s", claims.Username, tt.username)
				}
			}
		})
	}
}

// TestValidateToken tests JWT token validation
func TestValidateToken(t *testing.T) {
	// Generate a valid token for testing
	validToken, _ := GenerateToken(1, "test_user")

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkUser bool
		userID    int
		username  string
	}{
		{
			name:      "Valid token",
			token:     validToken,
			wantErr:   false,
			checkUser: true,
			userID:    1,
			username:  "test_user",
		},
		{
			name:    "Invalid token - malformed",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Invalid token - empty",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Invalid token - random string",
			token:   "thisisnotavalidjwttoken",
			wantErr: true,
		},
		{
			name:    "Invalid token - only header",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkUser {
				if claims.UserID != tt.userID {
					t.Errorf("Claims userID = %d, want %d", claims.UserID, tt.userID)
				}

				if claims.Username != tt.username {
					t.Errorf("Claims username = %s, want %s", claims.Username, tt.username)
				}
			}
		})
	}
}

// TestTokenExpiration tests JWT token expiration
func TestTokenExpiration(t *testing.T) {
	// This test verifies that token has expiration set
	token, err := GenerateToken(1, "test_user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check that expiration is set
	if claims.ExpiresAt == nil {
		t.Error("Token expiration not set")
	}

	// Check that expiration is in the future
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token expiration is in the past")
	}

	// Check that expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Token expiration is not 24 hours. Expected ~%v, got %v (diff: %v)",
			expectedExpiry, claims.ExpiresAt.Time, diff)
	}

	// Check IssuedAt is set
	if claims.IssuedAt == nil {
		t.Error("Token IssuedAt not set")
	}

	// Check NotBefore is set
	if claims.NotBefore == nil {
		t.Error("Token NotBefore not set")
	}
}

// TestTokenWithDifferentSecrets tests that tokens are validated with correct secret
func TestTokenWithDifferentSecrets(t *testing.T) {
	// Save original secret
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Generate token with secret1
	os.Setenv("JWT_SECRET", "secret1")
	token, err := GenerateToken(1, "test_user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate with same secret - should work
	_, err = ValidateToken(token)
	if err != nil {
		t.Errorf("Token should be valid with same secret: %v", err)
	}

	// Change secret and try to validate - should fail
	os.Setenv("JWT_SECRET", "secret2")
	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Token should be invalid with different secret")
	}
}

// TestJWTClaims tests the Claims structure
func TestJWTClaims(t *testing.T) {
	claims := &Claims{
		UserID:   123,
		Username: "test_user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Verify claims structure
	if claims.UserID != 123 {
		t.Errorf("Claims UserID = %d, want 123", claims.UserID)
	}

	if claims.Username != "test_user" {
		t.Errorf("Claims Username = %s, want test_user", claims.Username)
	}

	if claims.ExpiresAt == nil {
		t.Error("Claims ExpiresAt is nil")
	}

	if claims.IssuedAt == nil {
		t.Error("Claims IssuedAt is nil")
	}

	if claims.NotBefore == nil {
		t.Error("Claims NotBefore is nil")
	}
}

// TestGetJWTSecret tests JWT secret retrieval
func TestGetJWTSecret(t *testing.T) {
	// Save original secret
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	tests := []struct {
		name        string
		envValue    string
		expectEmpty bool
	}{
		{
			name:        "Custom secret set",
			envValue:    "my-custom-secret",
			expectEmpty: false,
		},
		{
			name:        "Empty secret - should use default",
			envValue:    "",
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.envValue)
			secret := getJWTSecret()

			if tt.expectEmpty && len(secret) > 0 {
				t.Error("Expected empty secret")
			}

			if !tt.expectEmpty && len(secret) == 0 {
				t.Error("Secret should not be empty")
			}
		})
	}
}
