package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUserHandler_OPTIONS tests CORS preflight request
func TestUserHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/users/testuser", nil)
	w := httptest.NewRecorder()

	UserHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestUserHandler_InvalidMethod tests invalid HTTP methods
func TestUserHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/users/testuser", nil)
			w := httptest.NewRecorder()

			UserHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

// TestUserHandler_EmptyUsername tests request with empty username
func TestUserHandler_EmptyUsername(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()

	UserHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUserHandler_WhitespaceUsername tests request with whitespace username
func TestUserHandler_WhitespaceUsername(t *testing.T) {
	// Use URL encoding for spaces
	req := httptest.NewRequest(http.MethodGet, "/users/%20%20%20", nil)
	w := httptest.NewRecorder()

	UserHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for whitespace username, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUserHandler_ValidUsername tests request with valid username format
func TestUserHandler_ValidUsername(t *testing.T) {
	t.Skip("Skipping database-dependent test")

	// This will fail at database level but tests validation logic
	req := httptest.NewRequest(http.MethodGet, "/users/validuser", nil)
	w := httptest.NewRecorder()

	UserHandler(w, req)

	// Should get 404 or 500 (not 400) since validation passed
	if w.Code == http.StatusBadRequest {
		t.Errorf("Should not get BadRequest for valid username format")
	}
}
