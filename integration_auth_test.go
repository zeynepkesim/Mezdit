package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"devops-be/internal/auth"
	"devops-be/internal/database"
	"devops-be/internal/handlers"
	"devops-be/internal/middleware"
)

func TestMain(m *testing.M) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	// Initialize database connection for tests
	database.Init()

	code := m.Run()

	// Cleanup
	database.Close()
	os.Exit(code)
}

// TestLoginHandler tests the login endpoint
func TestLoginHandler(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database not available, skipping integration test")
	}

	tests := []struct {
		name           string
		method         string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "Valid login request format",
			method:         http.MethodPost,
			body:           `{"username":"testuser","password":"password123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized, // Will fail auth but format is valid
		},
		{
			name:           "Empty username",
			method:         http.MethodPost,
			body:           `{"username":"","password":"password123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty password",
			method:         http.MethodPost,
			body:           `{"username":"testuser","password":""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `{"username":"test"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong content type",
			method:         http.MethodPost,
			body:           `{"username":"test","password":"pass"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/login", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			handlers.LoginHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// TestRegisterHandler tests the register endpoint
func TestRegisterHandler(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database not available, skipping integration test")
	}

	tests := []struct {
		name           string
		method         string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "Empty name",
			method:         http.MethodPost,
			body:           `{"name":"","username":"testuser","password":"password123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty username",
			method:         http.MethodPost,
			body:           `{"name":"Test User","username":"","password":"password123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty password",
			method:         http.MethodPost,
			body:           `{"name":"Test User","username":"testuser","password":""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Short password (less than 6 characters)",
			method:         http.MethodPost,
			body:           `{"name":"Test User","username":"testuser_short_pass_unique_xyz_123","password":"short"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `{"name":"Test"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong content type",
			method:         http.MethodPost,
			body:           `{"name":"Test","username":"test","password":"password123"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/register", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			handlers.RegisterHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// TestChangePasswordHandler tests the change password endpoint
func TestChangePasswordHandler(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database not available, skipping integration test")
	}

	tests := []struct {
		name           string
		method         string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "Invalid user ID (zero)",
			method:         http.MethodPost,
			body:           `{"userId":0,"currentPassword":"oldpass","newPassword":"newpass123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid user ID (negative)",
			method:         http.MethodPost,
			body:           `{"userId":-1,"currentPassword":"oldpass","newPassword":"newpass123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty current password",
			method:         http.MethodPost,
			body:           `{"userId":1,"currentPassword":"","newPassword":"newpass123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty new password",
			method:         http.MethodPost,
			body:           `{"userId":1,"currentPassword":"oldpass","newPassword":""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Short new password (less than 6 characters)",
			method:         http.MethodPost,
			body:           `{"userId":1,"currentPassword":"oldpass","newPassword":"short"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Same password",
			method:         http.MethodPost,
			body:           `{"userId":1,"currentPassword":"password123","newPassword":"password123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `{"userId":1`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong content type",
			method:         http.MethodPost,
			body:           `{"userId":1,"currentPassword":"old","newPassword":"new123456"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/change-password", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			handlers.ChangePasswordHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// TestJWTTokenGeneration tests JWT token generation
func TestJWTTokenGeneration(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		wantErr  bool
	}{
		{
			name:     "Valid token generation",
			userID:   1,
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "Valid token with different user",
			userID:   999,
			username: "anotheruser",
			wantErr:  false,
		},
		{
			name:     "Valid token with empty username",
			userID:   1,
			username: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.userID, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Error("Expected token to be generated, got empty string")
				}

				// Verify token can be validated
				claims, err := auth.ValidateToken(token)
				if err != nil {
					t.Errorf("Failed to validate generated token: %v", err)
				}

				if claims.UserID != tt.userID {
					t.Errorf("Expected UserID %d, got %d", tt.userID, claims.UserID)
				}

				if claims.Username != tt.username {
					t.Errorf("Expected Username %s, got %s", tt.username, claims.Username)
				}
			}
		})
	}
}

// TestJWTTokenValidation tests JWT token validation
func TestJWTTokenValidation(t *testing.T) {
	// Generate a valid token first
	validToken, _ := auth.GenerateToken(1, "testuser")

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Invalid token format",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Malformed token",
			token:   "notavalidtoken",
			wantErr: true,
		},
		{
			name:    "Token with wrong signature",
			token:   validToken[:len(validToken)-5] + "wrong",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims == nil {
				t.Error("Expected claims to be returned for valid token")
			}
		})
	}
}

// TestRequireAuthMiddleware tests the authentication middleware
func TestRequireAuthMiddleware(t *testing.T) {
	// Create a test handler that will be wrapped
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Generate a valid token
	validToken, _ := auth.GenerateToken(1, "testuser")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Invalid token format (no Bearer)",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Empty Bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			// Apply middleware
			handler := middleware.RequireAuth(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

// TestOptionalAuthMiddleware tests the optional authentication middleware
func TestOptionalAuthMiddleware(t *testing.T) {
	// Create a test handler that checks for user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := middleware.GetUserFromContext(r)
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	})

	// Generate a valid token
	validToken, _ := auth.GenerateToken(1, "testuser")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid token - authenticated",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "authenticated",
		},
		{
			name:           "No token - anonymous",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "anonymous",
		},
		{
			name:           "Invalid token - anonymous",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusOK,
			expectedBody:   "anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			// Apply middleware
			handler := middleware.OptionalAuth(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

// TestLoginResponseWithToken tests that successful login returns a token
func TestLoginResponseWithToken(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database not available, skipping integration test")
	}

	// This test requires a real user in the database
	testUsername := "test_auth_user"
	testPassword := "password123"

	// Try to create user (may already exist)
	database.CreateUser("Test Auth User", testUsername, testPassword)

	reqBody := map[string]string{
		"username": testUsername,
		"password": testPassword,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handlers.LoginHandler(rr, req)

	// If successful, should return token
	if rr.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check that response contains success=true
		if success, ok := response["success"].(bool); !ok || !success {
			t.Error("Expected success to be true")
		}

		// Check that response contains data with token
		if data, ok := response["data"].(map[string]interface{}); ok {
			if token, hasToken := data["token"].(string); !hasToken || token == "" {
				t.Error("Expected response to contain non-empty token")
			}
		} else {
			t.Error("Expected response to contain data object")
		}
	}
}
