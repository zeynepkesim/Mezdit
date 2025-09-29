package main

import (
	"fmt"
	"testing"
	"time"
)

// Mock data for testing
var mockUsers = []User{
	{
		ID:           1,
		Name:         "Test User",
		Username:     "testuser",
		Password:     "password123",
		ProfilePhoto: stringPtr("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1"),
		CreatedAt:    time.Now(),
	},
	{
		ID:           2,
		Name:         "Admin User",
		Username:     "admin",
		Password:     "admin123",
		ProfilePhoto: stringPtr("https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=2"),
		CreatedAt:    time.Now(),
	},
}

var mockPosts = []Post{
	{
		ID:            1,
		PublisherID:   1,
		PublisherName: "Test Publisher",
		ContentText:   "Bu bir test postudur",
		VoteCount:     5,
		CreatedAt:     time.Now(),
	},
	{
		ID:            2,
		PublisherID:   1,
		PublisherName: "Anonymous",
		ContentText:   "Anonim post içeriği",
		VoteCount:     0,
		CreatedAt:     time.Now(),
	},
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// Mock helper function
func mockGetUserByIDHelper(id int) *User {
	for _, user := range mockUsers {
		if user.ID == id {
			return &user
		}
	}
	return nil
}

// Mock AddPost function
func mockAddPost(publisherID int, publisherName string, contentText string, contentPhotoID *int) (*Post, error) {
	// Validation
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
	user := mockGetUserByIDHelper(publisherID)
	if user == nil {
		return nil, fmt.Errorf("publisher not found")
	}

	// Create new post
	newPost := Post{
		ID:             len(mockPosts) + 1,
		PublisherID:    publisherID,
		PublisherName:  publisherName,
		PublisherPhoto: user.ProfilePhoto,
		ContentText:    contentText,
		ContentPhotoID: contentPhotoID,
		VoteCount:      0,
		CreatedAt:      time.Now(),
	}

	return &newPost, nil
}

// Post creation test
func TestAddPostWithMockData(t *testing.T) {
	tests := []struct {
		name           string
		publisherID    int
		publisherName  string
		contentText    string
		contentPhotoID *int
		wantError      bool
		expectedError  string
	}{
		{
			name:          "Valid post creation",
			publisherID:   1, // testuser exists in mock data
			publisherName: "Test Publisher",
			contentText:   "Bu bir test postudur.",
			wantError:     false,
		},
		{
			name:          "Valid post with empty publisher name",
			publisherID:   1,
			publisherName: "",
			contentText:   "Anonymous post",
			wantError:     false,
		},
		{
			name:          "Invalid publisher ID - zero",
			publisherID:   0,
			publisherName: "Test",
			contentText:   "Content",
			wantError:     true,
			expectedError: "invalid publisher id",
		},
		{
			name:          "Invalid publisher ID - negative",
			publisherID:   -1,
			publisherName: "Test",
			contentText:   "Content",
			wantError:     true,
			expectedError: "invalid publisher id",
		},
		{
			name:          "Empty content text",
			publisherID:   1,
			publisherName: "Test",
			contentText:   "",
			wantError:     true,
			expectedError: "content text cannot be empty",
		},
		{
			name:          "Non-existent publisher",
			publisherID:   999,
			publisherName: "Test",
			contentText:   "Content",
			wantError:     true,
			expectedError: "publisher not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := mockAddPost(tt.publisherID, tt.publisherName, tt.contentText, tt.contentPhotoID)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if post == nil {
				t.Error("Expected post but got nil")
				return
			}

			// Verify post properties
			if post.PublisherID != tt.publisherID {
				t.Errorf("Expected publisherID %d, got %d", tt.publisherID, post.PublisherID)
			}

			expectedPublisherName := tt.publisherName
			if expectedPublisherName == "" {
				expectedPublisherName = "Anonymous"
			}
			if post.PublisherName != expectedPublisherName {
				t.Errorf("Expected publisherName '%s', got '%s'", expectedPublisherName, post.PublisherName)
			}

			if post.ContentText != tt.contentText {
				t.Errorf("Expected contentText '%s', got '%s'", tt.contentText, post.ContentText)
			}

			if post.ID <= 0 {
				t.Error("Expected positive post ID")
			}

			if post.VoteCount != 0 {
				t.Errorf("Expected vote count 0, got %d", post.VoteCount)
			}

			if post.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}

			// Verify publisher photo comes from user
			expectedUser := mockGetUserByIDHelper(tt.publisherID)
			if expectedUser != nil && expectedUser.ProfilePhoto != nil {
				if post.PublisherPhoto == nil {
					t.Error("Expected publisher photo to be set from user")
				} else if *post.PublisherPhoto != *expectedUser.ProfilePhoto {
					t.Errorf("Expected publisher photo '%s', got '%s'", *expectedUser.ProfilePhoto, *post.PublisherPhoto)
				}
			}
		})
	}
}

// Additional tests you could add to post_test.go:

// Test user validation
func TestCreateUserValidation(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		password  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid user creation",
			username:  "newuser",
			password:  "password123",
			wantError: false,
		},
		{
			name:      "Empty username",
			username:  "",
			password:  "password123",
			wantError: true,
			errorMsg:  "username is required",
		},
		{
			name:      "Empty password",
			username:  "newuser",
			password:  "",
			wantError: true,
			errorMsg:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test the validation logic of CreateUser function
			// You'd need to mock the database calls or use a test database
		})
	}
}

// Test vote logic
func TestVotePostLogic(t *testing.T) {
	tests := []struct {
		name          string
		postID        int
		userID        int
		upvoted       bool
		previousVote  *bool // nil = no previous vote, true = upvote, false = downvote
		expectedDelta int   // Expected vote count change
		wantError     bool
	}{
		{
			name:          "New upvote",
			postID:        1,
			userID:        1,
			upvoted:       true,
			previousVote:  nil,
			expectedDelta: 1,
			wantError:     false,
		},
		{
			name:          "New downvote",
			postID:        1,
			userID:        1,
			upvoted:       false,
			previousVote:  nil,
			expectedDelta: -1,
			wantError:     false,
		},
		{
			name:          "Toggle upvote off",
			postID:        1,
			userID:        1,
			upvoted:       true,
			previousVote:  boolPtr(true),
			expectedDelta: -1,
			wantError:     false,
		},
		{
			name:          "Change from upvote to downvote",
			postID:        1,
			userID:        1,
			upvoted:       false,
			previousVote:  boolPtr(true),
			expectedDelta: -2,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test the vote calculation logic
			// You'd need to implement the vote logic testing
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
