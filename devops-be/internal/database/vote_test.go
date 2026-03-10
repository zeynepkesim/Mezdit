package database

import (
	"testing"
)

// TestVotePost_InvalidPostID tests VotePost with invalid post IDs
func TestVotePost_InvalidPostID(t *testing.T) {
	tests := []struct {
		postID int
		name   string
	}{
		{0, "zero post ID"},
		{-1, "negative post ID"},
		{-100, "large negative post ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VotePost(tt.postID, 1, true)

			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}

			if err.Error() != "invalid post id" {
				t.Errorf("Expected 'invalid post id' error, got '%s'", err.Error())
			}
		})
	}
}

// TestVotePost_InvalidUserID tests VotePost with invalid user IDs
func TestVotePost_InvalidUserID(t *testing.T) {
	tests := []struct {
		userID int
		name   string
	}{
		{0, "zero user ID"},
		{-1, "negative user ID"},
		{-100, "large negative user ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VotePost(1, tt.userID, true)

			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}

			if err.Error() != "invalid user id" {
				t.Errorf("Expected 'invalid user id' error, got '%s'", err.Error())
			}
		})
	}
}

// TestVotePost_BothInvalid tests VotePost with both invalid IDs
func TestVotePost_BothInvalid(t *testing.T) {
	_, err := VotePost(0, 0, true)

	if err == nil {
		t.Error("Expected error for both invalid IDs")
	}

	// Should catch post ID first
	if err.Error() != "invalid post id" {
		t.Errorf("Expected 'invalid post id' error, got '%s'", err.Error())
	}
}

// TestVotePost_UpvoteAndDownvote tests both upvote and downvote logic paths
func TestVotePost_UpvoteAndDownvote(t *testing.T) {
	t.Skip("Skipping database-dependent test")
}
