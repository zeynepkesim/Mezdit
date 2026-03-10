package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestVoteHandler_OPTIONS tests CORS preflight request
func TestVoteHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/vote", nil)
	w := httptest.NewRecorder()

	VoteHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestVoteHandler_InvalidMethod tests invalid HTTP methods
func TestVoteHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/vote", nil)
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

// TestVoteHandler_InvalidPostID tests validation for post ID
func TestVoteHandler_InvalidPostID(t *testing.T) {
	tests := []struct {
		name   string
		postID int
		userID int
	}{
		{"Zero post ID", 0, 1},
		{"Negative post ID", -1, 1},
		{"Large negative post ID", -999, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"post_id": tt.postID,
				"user_id": tt.userID,
				"upvoted": true,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// TestVoteHandler_InvalidUserID tests validation for user ID
func TestVoteHandler_InvalidUserID(t *testing.T) {
	tests := []struct {
		name   string
		postID int
		userID int
	}{
		{"Zero user ID", 1, 0},
		{"Negative user ID", 1, -1},
		{"Large negative user ID", 1, -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"post_id": tt.postID,
				"user_id": tt.userID,
				"upvoted": true,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// TestVoteHandler_InvalidJSON tests with invalid JSON body
func TestVoteHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	VoteHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestVoteHandler_MissingFields tests with missing required fields
func TestVoteHandler_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"Missing post_id", map[string]interface{}{"user_id": 1, "upvoted": true}},
		{"Missing user_id", map[string]interface{}{"post_id": 1, "upvoted": true}},
		{"Missing upvoted", map[string]interface{}{"post_id": 1, "user_id": 1}},
		{"Empty body", map[string]interface{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for %s, got %d", http.StatusBadRequest, tt.name, w.Code)
			}
		})
	}
}

// TestVoteHandler_UpvoteAndDownvote tests both upvote and downvote values
func TestVoteHandler_UpvoteAndDownvote(t *testing.T) {
	tests := []struct {
		name    string
		upvoted bool
	}{
		{"Upvote", true},
		{"Downvote", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"post_id": 1,
				"user_id": 1,
				"upvoted": tt.upvoted,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			// Should pass validation (might fail at DB level)
			if w.Code == http.StatusBadRequest {
				// Check if it's validation error (not acceptable)
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				if errorType, ok := response["error_type"].(string); ok && errorType == "validation_error" {
					t.Errorf("Valid request should not fail validation")
				}
			}
		})
	}
}

// TestVoteHandler_EmptyJSON tests vote with empty JSON
func TestVoteHandler_EmptyJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	VoteHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestVoteHandler_LargeIDs tests with very large ID values
func TestVoteHandler_LargeIDs(t *testing.T) {
	tests := []struct {
		name   string
		postID int
		userID int
	}{
		{"Large post ID", 999999999, 1},
		{"Large user ID", 1, 999999999},
		{"Both large", 999999999, 999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"post_id": tt.postID,
				"user_id": tt.userID,
				"upvoted": true,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			VoteHandler(w, req)

			// Should pass validation (large positive numbers are valid)
			if w.Code == http.StatusBadRequest {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				if errorData, ok := response["error"].(map[string]interface{}); ok {
					if errorData["code"] == "validation_error" {
						t.Errorf("%s: Large positive IDs should pass validation", tt.name)
					}
				}
			}
		})
	}
}

// TestVoteHandler_MissingContentType tests without Content-Type header
func TestVoteHandler_MissingContentType(t *testing.T) {
	body := map[string]interface{}{
		"post_id": 1,
		"user_id": 1,
		"upvoted": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewReader(bodyBytes))
	// Not setting Content-Type header
	w := httptest.NewRecorder()

	VoteHandler(w, req)

	// Should still work or fail gracefully
	if w.Code == http.StatusOK {
		t.Log("Handler accepts requests without Content-Type")
	}
}
