package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoginHandler_Success tests successful login
func TestLoginHandler_Success(t *testing.T) {
	// Skip this test as it requires database connection
	// This test validates the handler structure exists and compiles correctly
	t.Skip("Skipping database-dependent test")
}

// TestLoginHandler_MissingUsername tests login with missing username
func TestLoginHandler_MissingUsername(t *testing.T) {
	reqBody := map[string]string{
		"username": "",
		"password": "password123",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	LoginHandler(w, req)

	// Should return 400 Bad Request for empty username
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false, got %v", response["success"])
	}
}

// TestLoginHandler_MissingPassword tests login with missing password
func TestLoginHandler_MissingPassword(t *testing.T) {
	reqBody := map[string]string{
		"username": "testuser",
		"password": "",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	LoginHandler(w, req)

	// Should return 400 Bad Request for empty password
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestLoginHandler_InvalidMethod tests login with wrong HTTP method
func TestLoginHandler_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	// Should return 405 Method Not Allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestLoginHandler_InvalidJSON tests login with invalid JSON
func TestLoginHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	LoginHandler(w, req)

	// Should return 400 Bad Request for invalid JSON
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestRegisterHandler_ValidationErrors tests registration validation
func TestRegisterHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		description    string
	}{
		{
			name: "Missing name",
			requestBody: map[string]string{
				"name":     "",
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty name",
		},
		{
			name: "Missing username",
			requestBody: map[string]string{
				"name":     "Test User",
				"username": "",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty username",
		},
		{
			name: "Missing password",
			requestBody: map[string]string{
				"name":     "Test User",
				"username": "testuser",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty password",
		},
		{
			name: "Password too short",
			requestBody: map[string]string{
				"name":     "Test User",
				"username": "testuser",
				"password": "12345",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject password less than 6 characters",
		},
		{
			name: "Whitespace only name",
			requestBody: map[string]string{
				"name":     "   ",
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject whitespace-only name",
		},
		{
			name: "Whitespace only username",
			requestBody: map[string]string{
				"name":     "Test User",
				"username": "   ",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject whitespace-only username",
		},
		{
			name: "Whitespace only password",
			requestBody: map[string]string{
				"name":     "Test User",
				"username": "testuser",
				"password": "   ",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject whitespace-only password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			RegisterHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d", tt.description, tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if response["success"] != false {
				t.Errorf("%s: Expected success=false", tt.description)
			}
		})
	}
}

// TestRegisterHandler_InvalidMethod tests register with wrong HTTP method
func TestRegisterHandler_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	// Should return 405 Method Not Allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestChangePasswordHandler_ValidationErrors tests password change validation
func TestChangePasswordHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "Invalid user ID (zero)",
			requestBody: map[string]interface{}{
				"userId":          0,
				"currentPassword": "oldpass",
				"newPassword":     "newpass123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject zero user ID",
		},
		{
			name: "Invalid user ID (negative)",
			requestBody: map[string]interface{}{
				"userId":          -1,
				"currentPassword": "oldpass",
				"newPassword":     "newpass123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject negative user ID",
		},
		{
			name: "Missing current password",
			requestBody: map[string]interface{}{
				"userId":          1,
				"currentPassword": "",
				"newPassword":     "newpass123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty current password",
		},
		{
			name: "Missing new password",
			requestBody: map[string]interface{}{
				"userId":          1,
				"currentPassword": "oldpass",
				"newPassword":     "",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty new password",
		},
		{
			name: "New password too short",
			requestBody: map[string]interface{}{
				"userId":          1,
				"currentPassword": "oldpass",
				"newPassword":     "12345",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject password less than 6 characters",
		},
		{
			name: "Whitespace current password",
			requestBody: map[string]interface{}{
				"userId":          1,
				"currentPassword": "   ",
				"newPassword":     "newpass123",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject whitespace-only current password",
		},
		{
			name: "Whitespace new password",
			requestBody: map[string]interface{}{
				"userId":          1,
				"currentPassword": "oldpass",
				"newPassword":     "   ",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject whitespace-only new password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ChangePasswordHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d", tt.description, tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if response["success"] != false {
				t.Errorf("%s: Expected success=false", tt.description)
			}
		})
	}
}

// TestChangePasswordHandler_InvalidMethod tests password change with wrong method
func TestChangePasswordHandler_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/change-password", nil)
	w := httptest.NewRecorder()

	ChangePasswordHandler(w, req)

	// Should return 405 Method Not Allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestLoginHandler_WhitespaceTrimming tests that whitespace is trimmed
func TestLoginHandler_WhitespaceTrimming(t *testing.T) {
	reqBody := map[string]string{
		"username": "  ",
		"password": "  ",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	LoginHandler(w, req)

	// Should return 400 because after trimming, fields are empty
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestRegisterHandler_PasswordMinLength tests password length validation
func TestRegisterHandler_PasswordMinLength(t *testing.T) {
	tests := []struct {
		password       string
		shouldPass     bool
		expectedStatus int
	}{
		{"12345", false, http.StatusBadRequest},   // Too short - validation should fail
	}

	for _, tt := range tests {
		reqBody := map[string]string{
			"name":     "Test User",
			"username": "testuser",
			"password": tt.password,
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		RegisterHandler(w, req)

		if w.Code != tt.expectedStatus {
			t.Errorf("Password '%s': Expected status %d, got %d", tt.password, tt.expectedStatus, w.Code)
		}
	}
}

// TestHandlers_CORSPreflight tests CORS preflight handling
func TestHandlers_CORSPreflight(t *testing.T) {
	handlers := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"Login", LoginHandler, "/login"},
		{"Register", RegisterHandler, "/register"},
		{"ChangePassword", ChangePasswordHandler, "/change-password"},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, h.path, nil)
			w := httptest.NewRecorder()

			h.handler(w, req)

			// Should return 200 OK for OPTIONS
			if w.Code != http.StatusOK {
				t.Errorf("%s: Expected status 200 for OPTIONS, got %d", h.name, w.Code)
			}
		})
	}
}

// TestLoginHandler_EmptyJSON tests login with empty JSON object
func TestLoginHandler_EmptyJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestRegisterHandler_EmptyJSON tests register with empty JSON object
func TestRegisterHandler_EmptyJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	RegisterHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestChangePasswordHandler_EmptyJSON tests password change with empty JSON
func TestChangePasswordHandler_EmptyJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ChangePasswordHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestLoginHandler_InvalidJSONFormat tests various invalid JSON formats
func TestLoginHandler_InvalidJSONFormat(t *testing.T) {
	invalidJSONs := []string{
		"{username: testuser}",          // Missing quotes
		`{"username": "test"`,           // Incomplete
		`{"username": testuser}`,        // Value without quotes
		"not json at all",
		`{"username": "test", "password": }`, // Missing value
	}

	for _, invalidJSON := range invalidJSONs {
		t.Run(invalidJSON, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(invalidJSON))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			LoginHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
			}
		})
	}
}

// TestRegisterHandler_InvalidJSONFormat tests invalid JSON for register
func TestRegisterHandler_InvalidJSONFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	RegisterHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestChangePasswordHandler_InvalidJSONFormat tests invalid JSON for password change
func TestChangePasswordHandler_InvalidJSONFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ChangePasswordHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
