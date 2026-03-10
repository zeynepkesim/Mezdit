package models

import (
	"encoding/json"
	"testing"
	"time"
)

// TestUser_TableName tests User TableName method
func TestUser_TableName(t *testing.T) {
	user := User{}
	tableName := user.TableName()

	if tableName != "users" {
		t.Errorf("Expected table name 'users', got '%s'", tableName)
	}
}

// TestImage_TableName tests Image TableName method
func TestImage_TableName(t *testing.T) {
	image := Image{}
	tableName := image.TableName()

	if tableName != "images" {
		t.Errorf("Expected table name 'images', got '%s'", tableName)
	}
}

// TestPost_TableName tests Post TableName method
func TestPost_TableName(t *testing.T) {
	post := Post{}
	tableName := post.TableName()

	if tableName != "posts" {
		t.Errorf("Expected table name 'posts', got '%s'", tableName)
	}
}

// TestUserVote_TableName tests UserVote TableName method
func TestUserVote_TableName(t *testing.T) {
	vote := UserVote{}
	tableName := vote.TableName()

	if tableName != "user_votes" {
		t.Errorf("Expected table name 'user_votes', got '%s'", tableName)
	}
}

// TestUser_Creation tests User struct creation
func TestUser_Creation(t *testing.T) {
	now := time.Now()
	photo := "https://example.com/photo.jpg"

	user := User{
		ID:           1,
		Name:         "John Doe",
		Username:     "johndoe",
		Password:     "hashedpassword",
		ProfilePhoto: &photo,
		CreatedAt:    now,
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}

	if user.Username != "johndoe" {
		t.Errorf("Expected username 'johndoe', got '%s'", user.Username)
	}

	if *user.ProfilePhoto != photo {
		t.Errorf("Expected profile photo '%s', got '%s'", photo, *user.ProfilePhoto)
	}
}

// TestUser_JSONMarshaling tests that password is not included in JSON
func TestUser_JSONMarshaling(t *testing.T) {
	user := User{
		ID:       1,
		Name:     "John Doe",
		Username: "johndoe",
		Password: "secretpassword",
	}

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user to JSON: %v", err)
	}

	jsonString := string(jsonBytes)

	// Password should not appear in JSON
	if contains(jsonString, "secretpassword") {
		t.Error("Password should not be included in JSON output")
	}

	// Other fields should appear
	if !contains(jsonString, "johndoe") {
		t.Error("Username should be included in JSON output")
	}

	if !contains(jsonString, "John Doe") {
		t.Error("Name should be included in JSON output")
	}
}

// TestPost_Creation tests Post struct creation
func TestPost_Creation(t *testing.T) {
	now := time.Now()
	isLiked := true
	photoID := 5
	photoPath := "images/test.jpg"
	publisherPhoto := "https://example.com/avatar.jpg"

	post := Post{
		ID:               1,
		PublisherID:      2,
		PublisherName:    "John Doe",
		PublisherPhoto:   &publisherPhoto,
		Header:           "Test Post",
		ContentText:      "This is a test post",
		ContentPhotoID:   &photoID,
		ContentPhotoPath: &photoPath,
		IsPrivate:        false,
		VoteCount:        10,
		IsLiked:          &isLiked,
		CreatedAt:        now,
	}

	if post.ID != 1 {
		t.Errorf("Expected ID 1, got %d", post.ID)
	}

	if post.PublisherID != 2 {
		t.Errorf("Expected publisher ID 2, got %d", post.PublisherID)
	}

	if post.Header != "Test Post" {
		t.Errorf("Expected header 'Test Post', got '%s'", post.Header)
	}

	if post.IsPrivate != false {
		t.Error("Expected IsPrivate to be false")
	}

	if post.VoteCount != 10 {
		t.Errorf("Expected vote count 10, got %d", post.VoteCount)
	}

	if *post.IsLiked != true {
		t.Error("Expected IsLiked to be true")
	}

	if *post.ContentPhotoID != photoID {
		t.Errorf("Expected content photo ID %d, got %d", photoID, *post.ContentPhotoID)
	}
}

// TestImage_Creation tests Image struct creation
func TestImage_Creation(t *testing.T) {
	now := time.Now()
	postID := 10

	image := Image{
		ID:        1,
		PostID:    &postID,
		Path:      "images/photo.jpg",
		CreatedAt: now,
	}

	if image.ID != 1 {
		t.Errorf("Expected ID 1, got %d", image.ID)
	}

	if *image.PostID != postID {
		t.Errorf("Expected post ID %d, got %d", postID, *image.PostID)
	}

	if image.Path != "images/photo.jpg" {
		t.Errorf("Expected path 'images/photo.jpg', got '%s'", image.Path)
	}
}

// TestUserVote_Creation tests UserVote struct creation
func TestUserVote_Creation(t *testing.T) {
	vote := UserVote{
		PostID:  5,
		UserID:  10,
		Upvoted: true,
	}

	if vote.PostID != 5 {
		t.Errorf("Expected post ID 5, got %d", vote.PostID)
	}

	if vote.UserID != 10 {
		t.Errorf("Expected user ID 10, got %d", vote.UserID)
	}

	if vote.Upvoted != true {
		t.Error("Expected upvoted to be true")
	}
}

// TestVote_Creation tests legacy Vote struct
func TestVote_Creation(t *testing.T) {
	vote := Vote{
		PostID:  5,
		UserID:  10,
		Upvoted: false,
	}

	if vote.PostID != 5 {
		t.Errorf("Expected post ID 5, got %d", vote.PostID)
	}

	if vote.UserID != 10 {
		t.Errorf("Expected user ID 10, got %d", vote.UserID)
	}

	if vote.Upvoted != false {
		t.Error("Expected upvoted to be false")
	}
}

// TestLoginRequest_JSONUnmarshal tests LoginRequest JSON unmarshaling
func TestLoginRequest_JSONUnmarshal(t *testing.T) {
	jsonString := `{"username": "testuser", "password": "testpass"}`

	var req LoginRequest
	err := json.Unmarshal([]byte(jsonString), &req)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if req.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", req.Username)
	}

	if req.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", req.Password)
	}
}

// TestCreateUserRequest_JSONUnmarshal tests CreateUserRequest JSON unmarshaling
func TestCreateUserRequest_JSONUnmarshal(t *testing.T) {
	jsonString := `{"name": "John Doe", "username": "johndoe", "password": "secret123"}`

	var req CreateUserRequest
	err := json.Unmarshal([]byte(jsonString), &req)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if req.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", req.Name)
	}

	if req.Username != "johndoe" {
		t.Errorf("Expected username 'johndoe', got '%s'", req.Username)
	}

	if req.Password != "secret123" {
		t.Errorf("Expected password 'secret123', got '%s'", req.Password)
	}
}

// TestVoteRequest_JSONUnmarshal tests VoteRequest JSON unmarshaling
func TestVoteRequest_JSONUnmarshal(t *testing.T) {
	jsonString := `{"postId": 5, "userId": 10, "upvoted": true}`

	var req VoteRequest
	err := json.Unmarshal([]byte(jsonString), &req)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if req.PostID != 5 {
		t.Errorf("Expected post ID 5, got %d", req.PostID)
	}

	if req.UserID != 10 {
		t.Errorf("Expected user ID 10, got %d", req.UserID)
	}

	if req.Upvoted != true {
		t.Error("Expected upvoted to be true")
	}
}

// TestAPISuccess_JSONMarshal tests APISuccess JSON marshaling
func TestAPISuccess_JSONMarshal(t *testing.T) {
	response := APISuccess{
		Success: true,
		Data: map[string]string{
			"message": "Operation successful",
		},
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	jsonString := string(jsonBytes)

	if !contains(jsonString, "\"success\":true") {
		t.Error("Expected success field in JSON")
	}

	if !contains(jsonString, "Operation successful") {
		t.Error("Expected data in JSON")
	}
}

// TestAPIError_JSONMarshal tests APIError JSON marshaling
func TestAPIError_JSONMarshal(t *testing.T) {
	response := APIError{
		Success: false,
		Error: APIErrorBody{
			Code:    "test_error",
			Message: "This is a test error",
		},
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	jsonString := string(jsonBytes)

	if !contains(jsonString, "\"success\":false") {
		t.Error("Expected success field to be false in JSON")
	}

	if !contains(jsonString, "test_error") {
		t.Error("Expected error code in JSON")
	}

	if !contains(jsonString, "This is a test error") {
		t.Error("Expected error message in JSON")
	}
}

// TestCreatePostRequest_WithPointer tests CreatePostRequest with pointer fields
func TestCreatePostRequest_WithPointer(t *testing.T) {
	photoID := 123
	req := CreatePostRequest{
		PublisherID:    1,
		PublisherName:  "Test User",
		Header:         "Test Header",
		ContentText:    "Test content",
		ContentPhotoID: &photoID,
		IsPrivate:      true,
	}

	if *req.ContentPhotoID != 123 {
		t.Errorf("Expected content photo ID 123, got %d", *req.ContentPhotoID)
	}

	if !req.IsPrivate {
		t.Error("Expected IsPrivate to be true")
	}
}

// TestChangePasswordRequest_Creation tests ChangePasswordRequest struct
func TestChangePasswordRequest_Creation(t *testing.T) {
	req := ChangePasswordRequest{
		UserID:          5,
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	}

	if req.UserID != 5 {
		t.Errorf("Expected user ID 5, got %d", req.UserID)
	}

	if req.CurrentPassword != "oldpass" {
		t.Errorf("Expected current password 'oldpass', got '%s'", req.CurrentPassword)
	}

	if req.NewPassword != "newpass" {
		t.Errorf("Expected new password 'newpass', got '%s'", req.NewPassword)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
