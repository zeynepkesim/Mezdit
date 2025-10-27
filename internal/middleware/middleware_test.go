package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestEnableCORS tests CORS header setting
func TestEnableCORS(t *testing.T) {
	w := httptest.NewRecorder()
	EnableCORS(w)

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin header to be *")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}

	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header to be set")
	}
}

// TestWriteJSON tests JSON response writing
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	WriteJSON(w, http.StatusOK, data)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set")
	}

	// Check JSON body
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	if response["message"] != "test" {
		t.Errorf("Expected message 'test', got '%s'", response["message"])
	}
}

// TestWriteOK tests successful response writing
func TestWriteOK(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	WriteOK(w, data)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("Expected success=true")
	}

	if response["data"] == nil {
		t.Error("Expected data field to be present")
	}
}

// TestWriteCreated tests created response writing
func TestWriteCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	WriteCreated(w, data)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("Expected success=true")
	}

	if response["data"] == nil {
		t.Error("Expected data field to be present")
	}
}

// TestWriteError tests error response writing
func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, http.StatusBadRequest, "test_error", "This is a test error")

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Error("Expected success=false")
	}

	errorData := response["error"].(map[string]interface{})
	if errorData["code"] != "test_error" {
		t.Errorf("Expected error code 'test_error', got '%s'", errorData["code"])
	}

	if errorData["message"] != "This is a test error" {
		t.Errorf("Expected error message 'This is a test error', got '%s'", errorData["message"])
	}
}

// TestRequireMethod tests HTTP method validation
func TestRequireMethod(t *testing.T) {
	tests := []struct {
		method         string
		requiredMethod string
		shouldPass     bool
		expectedStatus int
	}{
		{http.MethodPost, http.MethodPost, true, 0},
		{http.MethodGet, http.MethodPost, false, http.StatusMethodNotAllowed},
		{http.MethodPut, http.MethodPost, false, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.MethodDelete, true, 0},
		{http.MethodOptions, http.MethodPost, false, http.StatusOK}, // OPTIONS returns false but 200 OK
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			result := RequireMethod(w, req, tt.requiredMethod)

			if result != tt.shouldPass {
				t.Errorf("Method %s (required %s): Expected result %v, got %v", tt.method, tt.requiredMethod, tt.shouldPass, result)
			}

			if !tt.shouldPass && tt.expectedStatus > 0 && w.Code != tt.expectedStatus {
				t.Errorf("Method %s: Expected status %d, got %d", tt.method, tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestDecodeJSONBody tests JSON body decoding
func TestDecodeJSONBody(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name           string
		contentType    string
		body           string
		shouldPass     bool
		expectedStatus int
	}{
		{
			name:           "Valid JSON",
			contentType:    "application/json",
			body:           `{"name": "test", "value": 123}`,
			shouldPass:     true,
			expectedStatus: 0,
		},
		{
			name:           "Valid JSON with charset",
			contentType:    "application/json; charset=utf-8",
			body:           `{"name": "test", "value": 123}`,
			shouldPass:     true,
			expectedStatus: 0,
		},
		{
			name:           "Invalid content type",
			contentType:    "text/plain",
			body:           `{"name": "test", "value": 123}`,
			shouldPass:     false,
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "Invalid JSON",
			contentType:    "application/json",
			body:           `{"name": "test", "value": }`,
			shouldPass:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown fields",
			contentType:    "application/json",
			body:           `{"name": "test", "value": 123, "unknown": "field"}`,
			shouldPass:     false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			var dst TestStruct
			result := DecodeJSONBody(w, req, &dst)

			if result != tt.shouldPass {
				t.Errorf("%s: Expected result %v, got %v", tt.name, tt.shouldPass, result)
			}

			if !tt.shouldPass && w.Code != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			if tt.shouldPass {
				if dst.Name != "test" || dst.Value != 123 {
					t.Errorf("%s: Failed to decode JSON correctly", tt.name)
				}
			}
		})
	}
}

// TestDecodeJSONBody_EmptyBody tests decoding empty body
func TestDecodeJSONBody_EmptyBody(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	var dst TestStruct
	result := DecodeJSONBody(w, req, &dst)

	if result {
		t.Error("Expected DecodeJSONBody to fail on empty body")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestWriteJSON_DifferentStatusCodes tests various HTTP status codes
func TestWriteJSON_DifferentStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			w := httptest.NewRecorder()
			data := map[string]string{"status": http.StatusText(code)}

			WriteJSON(w, code, data)

			if w.Code != code {
				t.Errorf("Expected status %d, got %d", code, w.Code)
			}
		})
	}
}

// TestRequireMethod_OPTIONS tests OPTIONS handling for CORS
func TestRequireMethod_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	result := RequireMethod(w, req, http.MethodPost)

	// OPTIONS should return false (don't proceed) but set 200 OK
	if result {
		t.Error("Expected RequireMethod to return false for OPTIONS")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers are set
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set for OPTIONS")
	}
}
