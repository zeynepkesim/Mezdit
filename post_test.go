package main

import (
	"devops-be/internal/models"
	"fmt"
	"testing"
	"time"
)

// Mock data for testing
var mockUsers = []models.User{
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

var mockPosts = []models.Post{
	{
		ID:            1,
		PublisherID:   1, // to get fail, change this
		PublisherName: "Test Publisher",
		ContentText:   "This is a test post",
		VoteCount:     5,
		CreatedAt:     time.Now(),
	},
	{
		ID:            2,
		PublisherID:   1,
		PublisherName: "Anonymous",
		ContentText:   "Anonymous post content",
		VoteCount:     0,
		CreatedAt:     time.Now(),
	},
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// Mock helper function
func mockGetUserByIDHelper(id int) *models.User {
	for _, user := range mockUsers {
		if user.ID == id {
			return &user
		}
	}
	return nil
}

// Mock AddPost function
func mockAddPost(publisherID int, publisherName string, contentText string, contentPhotoID *int) (*models.Post, error) {
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
	newPost := models.Post{
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

// /////////////////////// Post creation test
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

// ////////////////////// Test vote logic
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

func intPtr(i int) *int {
	return &i
}

// Mock EditPost function
func mockEditPost(postID int, userID int, newContentText string, newPublisherName string, contentPhotoID *int) (*models.Post, error) {
	// Validation
	if postID <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if newContentText == "" {
		return nil, fmt.Errorf("content text cannot be empty")
	}

	// If publisherName is empty, set to Anonymous
	if newPublisherName == "" {
		newPublisherName = "Anonymous"
	}

	// Find the post
	var targetPost *models.Post
	for i := range mockPosts {
		if mockPosts[i].ID == postID {
			targetPost = &mockPosts[i]
			break
		}
	}

	if targetPost == nil {
		return nil, fmt.Errorf("post not found")
	}

	// Check if user exists
	user := mockGetUserByIDHelper(userID)
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Security check: verify that the user is the owner of the post
	if targetPost.PublisherID != userID {
		return nil, fmt.Errorf("you are not the owner of this post")
	}

	// Update the post
	targetPost.ContentText = newContentText
	targetPost.PublisherName = newPublisherName
	if contentPhotoID != nil {
		targetPost.ContentPhotoID = contentPhotoID
	}

	return targetPost, nil
}

// ////////////////////////// Test edit post functionality
func TestEditPostWithMockData(t *testing.T) {
	tests := []struct {
		name             string
		postID           int
		userID           int
		newContentText   string
		newPublisherName string
		contentPhotoID   *int
		wantError        bool
		expectedError    string
	}{
		{
			name:             "Valid post edit",
			postID:           1,
			userID:           1, // testuser is the owner of post 1
			newContentText:   "Updated content",
			newPublisherName: "Updated Name",
			wantError:        false,
		},
		{
			name:             "Valid post edit with photo",
			postID:           1,
			userID:           1,
			newContentText:   "Photo update content",
			newPublisherName: "Photo Update",
			contentPhotoID:   intPtr(123),
			wantError:        false,
		},
		{
			name:             "Edit with empty publisher name (should become Anonymous)",
			postID:           1,
			userID:           1,
			newContentText:   "Anonymous update",
			newPublisherName: "",
			wantError:        false,
		},
		{
			name:             "Invalid post ID - zero",
			postID:           0,
			userID:           1,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "invalid post id",
		},
		{
			name:             "Invalid post ID - negative",
			postID:           -1,
			userID:           1,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "invalid post id",
		},
		{
			name:             "Invalid user ID - zero",
			postID:           1,
			userID:           0,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "invalid user id",
		},
		{
			name:             "Invalid user ID - negative",
			postID:           1,
			userID:           -1,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "invalid user id",
		},
		{
			name:             "Empty content text",
			postID:           1,
			userID:           1,
			newContentText:   "",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "content text cannot be empty",
		},
		{
			name:             "Non-existent post",
			postID:           999,
			userID:           1,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "post not found",
		},
		{
			name:             "Non-existent user",
			postID:           1,
			userID:           999,
			newContentText:   "Content",
			newPublisherName: "Name",
			wantError:        true,
			expectedError:    "user not found",
		},
		{
			name:             "User is not the post owner",
			postID:           1,
			userID:           2, // admin user trying to edit testuser's post
			newContentText:   "Unauthorized update",
			newPublisherName: "Hacker",
			wantError:        true,
			expectedError:    "you are not the owner of this post",
		},
	}
	//test pipeliennow.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := mockEditPost(tt.postID, tt.userID, tt.newContentText, tt.newPublisherName, tt.contentPhotoID)

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
			if post.ID != tt.postID {
				t.Errorf("Expected postID %d, got %d", tt.postID, post.ID)
			}

			if post.ContentText != tt.newContentText {
				t.Errorf("Expected contentText '%s', got '%s'", tt.newContentText, post.ContentText)
			}

			expectedPublisherName := tt.newPublisherName
			if expectedPublisherName == "" {
				expectedPublisherName = "Anonymous"
			}
			if post.PublisherName != expectedPublisherName {
				t.Errorf("Expected publisherName '%s', got '%s'", expectedPublisherName, post.PublisherName)
			}

			if tt.contentPhotoID != nil {
				if post.ContentPhotoID == nil {
					t.Error("Expected contentPhotoID to be set")
				} else if *post.ContentPhotoID != *tt.contentPhotoID {
					t.Errorf("Expected contentPhotoID %d, got %d", *tt.contentPhotoID, *post.ContentPhotoID)
				}
			}

			// Verify that the publisher ID hasn't changed
			if post.PublisherID != tt.userID {
				t.Errorf("Expected publisherID %d, got %d", tt.userID, post.PublisherID)
			}
		})
	}
}
