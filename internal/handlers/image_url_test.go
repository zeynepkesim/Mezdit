package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageURLHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/image-url", nil)
			w := httptest.NewRecorder()

			ImageURLHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for %s method, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

func TestImageURLHandler_InvalidJSON(t *testing.T) {
	invalidJSONs := []string{
		"not json",
		"{invalid}",
		"{'object_path': 'test'}",
		"{\"object_path\": }",
		"",
	}

	for _, invalidJSON := range invalidJSONs {
		t.Run(invalidJSON, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/image-url", bytes.NewBufferString(invalidJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ImageURLHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestImageURLHandler_EmptyObjectPath(t *testing.T) {
	requestBody := ImageURLRequest{
		ObjectPath: "",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/image-url", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ImageURLHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for empty object_path, got %d", http.StatusBadRequest, w.Code)
	}

	body := w.Body.String()
	if body != "object_path is required\n" {
		t.Errorf("Expected error message 'object_path is required', got '%s'", body)
	}
}

func TestImageURLHandler_MinIOError(t *testing.T) {
	// When MinIO client is not initialized, we expect an error
	requestBody := ImageURLRequest{
		ObjectPath: "images/test.jpg",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/image-url", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ImageURLHandler(w, req)

	// Should return internal server error since MinIO is not initialized
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d when MinIO fails, got %d", http.StatusInternalServerError, w.Code)
	}

	body := w.Body.String()
	if body != "Failed to generate image URL\n" {
		t.Errorf("Expected error message 'Failed to generate image URL', got '%s'", body)
	}
}

func TestImageURLHandler_ValidRequestFormat(t *testing.T) {
	testCases := []struct {
		name       string
		objectPath string
	}{
		{
			name:       "normal path",
			objectPath: "images/test.jpg",
		},
		{
			name:       "path with UUID",
			objectPath: "images/abc123-def456_test.jpg",
		},
		{
			name:       "path with special chars",
			objectPath: "images/test_file-2024.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := ImageURLRequest{
				ObjectPath: tc.objectPath,
			}

			jsonBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/image-url", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ImageURLHandler(w, req)

			// Will fail with internal server error since MinIO is not initialized,
			// but this verifies the request format is accepted
			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for valid request format, got %d", http.StatusInternalServerError, w.Code)
			}
		})
	}
}

func TestImageURLHandler_ContentType(t *testing.T) {
	// Test that handler rejects requests without proper content-type
	requestBody := ImageURLRequest{
		ObjectPath: "test.jpg",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/image-url", bytes.NewBuffer(jsonBody))
	// Intentionally not setting Content-Type
	w := httptest.NewRecorder()

	ImageURLHandler(w, req)

	// Should still process the request (handler doesn't enforce content-type)
	// Will get MinIO error since client is not initialized
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestImageURLRequest_JSONSerialization(t *testing.T) {
	req := ImageURLRequest{
		ObjectPath: "images/test.jpg",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ImageURLRequest: %v", err)
	}

	var decoded ImageURLRequest
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ImageURLRequest: %v", err)
	}

	if decoded.ObjectPath != req.ObjectPath {
		t.Errorf("Expected object_path %s, got %s", req.ObjectPath, decoded.ObjectPath)
	}
}

func TestImageURLResponse_JSONSerialization(t *testing.T) {
	resp := ImageURLResponse{
		URL: "https://example.com/presigned-url?token=abc123",
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ImageURLResponse: %v", err)
	}

	var decoded ImageURLResponse
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ImageURLResponse: %v", err)
	}

	if decoded.URL != resp.URL {
		t.Errorf("Expected URL %s, got %s", resp.URL, decoded.URL)
	}
}
