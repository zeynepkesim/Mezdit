package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-be/internal/auth"
)

// TestRequireAuth_MissingToken tests auth middleware with missing token
func TestRequireAuth_MissingToken(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with auth middleware
	handler := RequireAuth(testHandler)

	// Make request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Error("Expected success=false")
	}

	errorData := response["error"].(map[string]interface{})
	if errorData["code"] != "missing_token" {
		t.Errorf("Expected error code 'missing_token', got '%s'", errorData["code"])
	}
}

// TestRequireAuth_InvalidTokenFormat tests various invalid token formats
func TestRequireAuth_InvalidTokenFormat(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAuth(testHandler)

	tests := []struct {
		name        string
		authHeader  string
		expectedErr string
	}{
		{
			name:        "Just token without Bearer",
			authHeader:  "some-token-here",
			expectedErr: "invalid_token_format",
		},
		{
			name:        "Wrong prefix",
			authHeader:  "Token some-token-here",
			expectedErr: "invalid_token_format",
		},
		{
			name:        "Missing token after Bearer",
			authHeader:  "Bearer",
			expectedErr: "invalid_token_format",
		},
		{
			name:        "Extra spaces",
			authHeader:  "Bearer  token  extra",
			expectedErr: "invalid_token_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			errorData := response["error"].(map[string]interface{})
			if errorData["code"] != tt.expectedErr {
				t.Errorf("Expected error code '%s', got '%s'", tt.expectedErr, errorData["code"])
			}
		})
	}
}

// TestRequireAuth_InvalidToken tests auth with invalid JWT token
func TestRequireAuth_InvalidToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 401 for invalid token
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	errorData := response["error"].(map[string]interface{})
	if errorData["code"] != "invalid_token" {
		t.Errorf("Expected error code 'invalid_token', got '%s'", errorData["code"])
	}
}

// TestRequireAuth_ValidToken tests auth with valid JWT token
func TestRequireAuth_ValidToken(t *testing.T) {
	// Create a test handler that checks user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r)
		if !ok {
			t.Error("Expected user context to be set")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.UserID != 123 || user.Username != "testuser" {
			t.Errorf("Expected user ID 123 and username 'testuser', got ID %d and username '%s'", user.UserID, user.Username)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := RequireAuth(testHandler)

	// Generate a valid token
	token, err := auth.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", w.Body.String())
	}
}

// TestRequireAuth_OPTIONS tests OPTIONS request handling
func TestRequireAuth_OPTIONS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for OPTIONS request")
	})

	handler := RequireAuth(testHandler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 OK for OPTIONS
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set")
	}
}

// TestOptionalAuth_WithoutToken tests optional auth without token
func TestOptionalAuth_WithoutToken(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// User context should not be set
		_, ok := GetUserFromContext(r)
		if ok {
			t.Error("Expected no user context without token")
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even without token")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestOptionalAuth_WithValidToken tests optional auth with valid token
func TestOptionalAuth_WithValidToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// User context should be set
		user, ok := GetUserFromContext(r)
		if !ok {
			t.Error("Expected user context to be set with valid token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.UserID != 456 || user.Username != "optionaluser" {
			t.Errorf("Expected user ID 456 and username 'optionaluser', got ID %d and username '%s'", user.UserID, user.Username)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth(testHandler)

	// Generate a valid token
	token, err := auth.GenerateToken(456, "optionaluser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestOptionalAuth_WithInvalidToken tests optional auth with invalid token
func TestOptionalAuth_WithInvalidToken(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// User context should not be set for invalid token
		_, ok := GetUserFromContext(r)
		if ok {
			t.Error("Expected no user context with invalid token")
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should still call handler (optional auth)
	if !handlerCalled {
		t.Error("Handler should be called even with invalid token (optional auth)")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestOptionalAuth_OPTIONS tests OPTIONS request handling
func TestOptionalAuth_OPTIONS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for OPTIONS request")
	})

	handler := OptionalAuth(testHandler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 OK for OPTIONS
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set")
	}
}

// TestGetUserFromContext_NoContext tests getting user from request without context
func TestGetUserFromContext_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	user, ok := GetUserFromContext(req)

	if ok {
		t.Error("Expected ok to be false when no user context is set")
	}

	if user != nil {
		t.Error("Expected user to be nil when no context is set")
	}
}

// TestOptionalAuth_MalformedAuthHeader tests various malformed auth headers
func TestOptionalAuth_MalformedAuthHeader(t *testing.T) {
	tests := []string{
		"InvalidFormat",
		"Bearer",
		"Token some-token",
		"Bearer  ",
	}

	for _, authHeader := range tests {
		t.Run(authHeader, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Should proceed without user context
				_, ok := GetUserFromContext(r)
				if ok {
					t.Error("Expected no user context with malformed auth header")
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := OptionalAuth(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", authHeader)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}
