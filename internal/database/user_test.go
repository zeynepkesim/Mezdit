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
			name:    "Too low cost (3)",
			cost:    3,
			wantErr: true,
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
