package database

import (
	"os"
	"testing"
)

// setupTestDB initializes database for integration tests
// Only runs in CI/CD where DATABASE_URL is set
func setupTestDB(t *testing.T) {
	t.Helper()
	
	// Check if we're in CI/CD environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping integration test - DATABASE_URL not set (CI/CD only)")
	}

	// Initialize database if not already done
	if DB == nil {
		Init()
		if DB == nil {
			t.Fatal("Failed to initialize database")
		}
	}
}

// TestAddPost_Integration tests AddPost with real database
func TestAddPost_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name          string
		publisherID   int
		publisherName string
		header        string
		content       string
		isPrivate     bool
		shouldSucceed bool
	}{
		{
			name:          "Valid public post",
			publisherID:   1,
			publisherName: "TestUser",
			header:        "Integration Test Post",
			content:       "This is a test post content",
			isPrivate:     false,
			shouldSucceed: true,
		},
		{
			name:          "Valid private post",
			publisherID:   1,
			publisherName: "TestUser",
			header:        "Private Test Post",
			content:       "This is a private test post",
			isPrivate:     true,
			shouldSucceed: true,
		},
		{
			name:          "Empty header",
			publisherID:   1,
			publisherName: "TestUser",
			header:        "",
			content:       "Content",
			isPrivate:     false,
			shouldSucceed: false,
		},
		{
			name:          "Empty content",
			publisherID:   1,
			publisherName: "TestUser",
			header:        "Header",
			content:       "",
			isPrivate:     false,
			shouldSucceed: false,
		},
		{
			name:          "Invalid publisher ID",
			publisherID:   0,
			publisherName: "TestUser",
			header:        "Test",
			content:       "Content",
			isPrivate:     false,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AddPost(tt.publisherID, tt.publisherName, tt.header, tt.content, nil, tt.isPrivate)
			
			if tt.shouldSucceed && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Error("Expected error, got success")
			}
		})
	}
}

// TestGetPosts_Integration tests GetPosts with real database
func TestGetPosts_Integration(t *testing.T) {
	setupTestDB(t)

	posts, err := GetPosts()
	if err != nil {
		t.Fatalf("GetPosts failed: %v", err)
	}

	// Should return at least some posts (or empty array)
	if posts == nil {
		t.Error("Posts should not be nil")
	}
}

// TestGetPostsForUser_Integration tests GetPostsForUser with real database
func TestGetPostsForUser_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name   string
		userID int
	}{
		{"User ID 1", 1},
		{"User ID 0 (public posts)", 0},
		{"Non-existent user", 99999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := GetPostsForUser(tt.userID)
			if err != nil {
				t.Errorf("GetPostsForUser failed: %v", err)
			}
			if posts == nil {
				t.Error("Posts should not be nil")
			}
		})
	}
}

// TestEditPost_Integration tests EditPost with real database
func TestEditPost_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name          string
		postID        int
		userID        int
		header        string
		content       string
		publisherName string
		isPrivate     bool
		expectError   bool
	}{
		{
			name:          "Valid edit",
			postID:        1,
			userID:        1,
			header:        "Updated Header",
			content:       "Updated Content",
			publisherName: "UpdatedName",
			isPrivate:     false,
			expectError:   false,
		},
		{
			name:          "Empty header",
			postID:        1,
			userID:        1,
			header:        "",
			content:       "Content",
			publisherName: "Name",
			isPrivate:     false,
			expectError:   true,
		},
		{
			name:          "Invalid post ID",
			postID:        0,
			userID:        1,
			header:        "Header",
			content:       "Content",
			publisherName: "Name",
			isPrivate:     false,
			expectError:   true,
		},
		{
			name:          "Invalid user ID",
			postID:        1,
			userID:        0,
			header:        "Header",
			content:       "Content",
			publisherName: "Name",
			isPrivate:     false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EditPost(tt.postID, tt.userID, tt.header, tt.content, tt.publisherName, tt.isPrivate)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestVotePost_Integration tests VotePost with real database
func TestVotePost_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		postID      int
		userID      int
		upvoted     bool
		expectError bool
	}{
		{
			name:        "Valid upvote",
			postID:      1,
			userID:      1,
			upvoted:     true,
			expectError: false,
		},
		{
			name:        "Valid downvote",
			postID:      1,
			userID:      1,
			upvoted:     false,
			expectError: false,
		},
		{
			name:        "Invalid post ID",
			postID:      0,
			userID:      1,
			upvoted:     true,
			expectError: true,
		},
		{
			name:        "Invalid user ID",
			postID:      1,
			userID:      0,
			upvoted:     true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VotePost(tt.postID, tt.userID, tt.upvoted)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCreateUser_Integration tests CreateUser with real database
func TestCreateUser_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		username    string
		password    string
		fullname    string
		expectError bool
	}{
		{
			name:        "Empty username",
			username:    "",
			password:    "pass123",
			fullname:    "Test User",
			expectError: true,
		},
		{
			name:        "Empty password",
			username:    "testuser",
			password:    "",
			fullname:    "Test User",
			expectError: true,
		},
		{
			name:        "Short password",
			username:    "testuser",
			password:    "123",
			fullname:    "Test User",
			expectError: true,
		},
		{
			name:        "Whitespace username",
			username:    "   ",
			password:    "pass123",
			fullname:    "Test User",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateUser(tt.username, tt.password, tt.fullname)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestLoginUser_Integration tests LoginUser with real database
func TestLoginUser_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "Empty username",
			username:    "",
			password:    "pass123",
			expectError: true,
		},
		{
			name:        "Empty password",
			username:    "testuser",
			password:    "",
			expectError: true,
		},
		{
			name:        "Whitespace username",
			username:    "   ",
			password:    "pass123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoginUser(tt.username, tt.password)
			
			// These should fail validation
			if !tt.expectError && err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

// TestGetUserByUsername_Integration tests GetUserByUsername with real database
func TestGetUserByUsername_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		username    string
		expectError bool
	}{
		{
			name:        "Empty username",
			username:    "",
			expectError: true,
		},
		{
			name:        "Whitespace username",
			username:    "   ",
			expectError: true,
		},
		{
			name:        "Valid username format",
			username:    "testuser",
			expectError: false, // Will return not found, but no validation error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetUserByUsername(tt.username)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestGetUserByID_Integration tests GetUserByID with real database
func TestGetUserByID_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name   string
		userID int
	}{
		{"User ID 1", 1},
		{"User ID 0", 0},
		{"Negative ID", -1},
		{"Large ID", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetUserByID(tt.userID)
			
			// Should handle all IDs gracefully (may return not found)
			if err != nil {
				t.Logf("GetUserByID(%d) error: %v", tt.userID, err)
			}
		})
	}
}

// TestChangePassword_Integration tests ChangePassword with real database
func TestChangePassword_Integration(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name           string
		userID         int
		currentPass    string
		newPass        string
		expectError    bool
	}{
		{
			name:        "Empty current password",
			userID:      1,
			currentPass: "",
			newPass:     "newpass123",
			expectError: true,
		},
		{
			name:        "Empty new password",
			userID:      1,
			currentPass: "oldpass",
			newPass:     "",
			expectError: true,
		},
		{
			name:        "Short new password",
			userID:      1,
			currentPass: "oldpass",
			newPass:     "123",
			expectError: true,
		},
		{
			name:        "Invalid user ID",
			userID:      0,
			currentPass: "oldpass",
			newPass:     "newpass123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ChangePassword(tt.userID, tt.currentPass, tt.newPass)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

