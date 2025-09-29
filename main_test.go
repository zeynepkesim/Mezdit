package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test setup - skip tests that require database connection
func init() {
	// Skip tests if database is not available
	// This prevents the tests from trying to connect to a real database
}

// HTTP Handler Tests - These test the actual HTTP endpoints

func TestLoginHandler(t *testing.T) {
	// Skip database-dependent tests for now
	t.Skip("Skipping database-dependent tests - requires database setup")

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty username",
			method:         "POST",
			body:           `{"username": "", "password": "password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Username is required",
		},
		{
			name:           "Empty password",
			method:         "POST",
			body:           `{"username": "testuser", "password": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password is required",
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			body:           `{"username": "testuser", "password": "password123"`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request body is not valid JSON",
		},
		{
			name:           "Wrong method - GET",
			method:         "GET",
			body:           `{"username": "testuser", "password": "password123"}`,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "CORS preflight request",
			method:         "OPTIONS",
			body:           "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/login", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(loginHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check CORS headers
			if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Error("Expected CORS header not set")
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response apiError
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if !strings.Contains(response.Error.Message, tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, response.Error.Message)
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	// Skip database-dependent tests for now
	t.Skip("Skipping database-dependent tests - requires database setup")

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty name",
			body:           `{"name": "", "username": "newuser", "password": "password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name is required",
		},
		{
			name:           "Empty username",
			body:           `{"name": "New User", "username": "", "password": "password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Username is required",
		},
		{
			name:           "Empty password",
			body:           `{"name": "New User", "username": "newuser", "password": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/register", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(registerHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response apiError
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if !strings.Contains(response.Error.Message, tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, response.Error.Message)
				}
			}
		})
	}
}

func TestPostsHandler(t *testing.T) {
	// Skip database-dependent tests for now
	t.Skip("Skipping database-dependent tests - requires database setup")

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET posts with invalid userId",
			method:         "GET",
			url:            "/posts?userId=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST post with empty content",
			method:         "POST",
			url:            "/posts",
			body:           `{"publisherId": 1, "publisherName": "Test User", "contentText": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "CORS preflight",
			method:         "OPTIONS",
			url:            "/posts",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(postsHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestVoteHandler(t *testing.T) {
	// Skip database-dependent tests for now
	t.Skip("Skipping database-dependent tests - requires database setup")

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid post ID",
			body:           `{"postId": 0, "userId": 1, "upvoted": true}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Post ID must be a positive integer",
		},
		{
			name:           "Invalid user ID",
			body:           `{"postId": 1, "userId": 0, "upvoted": true}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "User ID must be a positive integer",
		},
		{
			name:           "Invalid JSON",
			body:           `{"postId": 1, "userId": 1, "upvoted": true`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request body is not valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/vote", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(voteHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response apiError
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if !strings.Contains(response.Error.Message, tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, response.Error.Message)
				}
			}
		})
	}
}

// Test middleware functions
func TestEnableCORS(t *testing.T) {
	rr := httptest.NewRecorder()
	enableCORS(rr)

	// Check CORS headers
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin header not set")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Error("Expected Access-Control-Allow-Methods header not set")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Error("Expected Access-Control-Allow-Headers header not set")
	}
}

func TestRequireMethod(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requiredMethod string
		expectedResult bool
		expectedStatus int
	}{
		{
			name:           "Correct method",
			method:         "POST",
			requiredMethod: "POST",
			expectedResult: true,
		},
		{
			name:           "Wrong method",
			method:         "GET",
			requiredMethod: "POST",
			expectedResult: false,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			requiredMethod: "POST",
			expectedResult: false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			result := requireMethod(rr, req, tt.requiredMethod)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			if tt.expectedStatus != 0 && rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestDecodeJSONBody(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedResult bool
		expectedStatus int
	}{
		{
			name:           "Valid JSON",
			contentType:    "application/json",
			body:           `{"username": "test", "password": "pass"}`,
			expectedResult: true,
		},
		{
			name:           "Invalid content type",
			contentType:    "text/plain",
			body:           `{"username": "test", "password": "pass"}`,
			expectedResult: false,
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "Invalid JSON",
			contentType:    "application/json",
			body:           `{"username": "test", "password": "pass"`,
			expectedResult: false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			var dst map[string]interface{}
			result := decodeJSONBody(rr, req, &dst)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			if tt.expectedStatus != 0 && rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
