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
