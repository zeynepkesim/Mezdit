package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUploadImageHandler_OPTIONS tests CORS preflight
func TestUploadImageHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/upload", nil)
	w := httptest.NewRecorder()

	UploadImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestUploadImageHandler_InvalidMethod tests invalid HTTP methods
func TestUploadImageHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/upload", nil)
			w := httptest.NewRecorder()

			UploadImageHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

// TestUploadImageHandler_MissingImage tests POST without image
func TestUploadImageHandler_MissingImage(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	UploadImageHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUploadPostHandler_OPTIONS tests CORS preflight
func TestUploadPostHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/posts/upload", nil)
	w := httptest.NewRecorder()

	UploadPostHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestUploadPostHandler_InvalidMethod tests invalid HTTP methods
func TestUploadPostHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/posts/upload", nil)
			w := httptest.NewRecorder()

			UploadPostHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

// TestUploadPostHandler_MissingFields tests POST with missing fields
func TestUploadPostHandler_MissingFields(t *testing.T) {
	tests := []struct {
		name          string
		publisherID   string
		publisherName string
		header        string
		contentText   string
		expectedError string
	}{
		{
			name:          "Missing publisher ID",
			publisherID:   "",
			publisherName: "Test User",
			header:        "Test Header",
			contentText:   "Test content",
			expectedError: "Publisher ID is required",
		},
		{
			name:          "Missing header",
			publisherID:   "1",
			publisherName: "Test User",
			header:        "",
			contentText:   "Test content",
			expectedError: "Header cannot be empty",
		},
		{
			name:          "Missing content text",
			publisherID:   "1",
			publisherName: "Test User",
			header:        "Test Header",
			contentText:   "",
			expectedError: "Content text cannot be empty",
		},
		{
			name:          "Invalid publisher ID - non-numeric",
			publisherID:   "abc",
			publisherName: "Test User",
			header:        "Test Header",
			contentText:   "Test content",
			expectedError: "Invalid publisher ID",
		},
		{
			name:          "Invalid publisher ID - negative",
			publisherID:   "-1",
			publisherName: "Test User",
			header:        "Test Header",
			contentText:   "Test content",
			expectedError: "Invalid publisher ID",
		},
		{
			name:          "Invalid publisher ID - zero",
			publisherID:   "0",
			publisherName: "Test User",
			header:        "Test Header",
			contentText:   "Test content",
			expectedError: "Invalid publisher ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			writer.WriteField("publisherId", tt.publisherID)
			writer.WriteField("publisherName", tt.publisherName)
			writer.WriteField("header", tt.header)
			writer.WriteField("contentText", tt.contentText)

			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/posts/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			UploadPostHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: Expected status %d, got %d", tt.name, http.StatusBadRequest, w.Code)
			}
		})
	}
}

// TestEditPostWithImageHandler_OPTIONS tests CORS preflight
func TestEditPostWithImageHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/posts/edit/image", nil)
	w := httptest.NewRecorder()

	EditPostWithImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestEditPostWithImageHandler_InvalidMethod tests invalid HTTP methods
func TestEditPostWithImageHandler_InvalidMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/posts/edit/image", nil)
			w := httptest.NewRecorder()

			EditPostWithImageHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}
