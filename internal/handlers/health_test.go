package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-be/internal/database"
)

func TestHealthzHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Health check returns OK - GET",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Health check returns OK - POST",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Health check returns OK - HEAD",
			method:         http.MethodHead,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			w := httptest.NewRecorder()

			HealthzHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method != http.MethodHead && w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestReadyzHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "DB nil - returns service unavailable",
			setupDB: func() {
				database.DB = nil
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "db not ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original DB
			originalDB := database.DB
			defer func() {
				database.DB = originalDB
			}()

			tt.setupDB()

			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			w := httptest.NewRecorder()

			ReadyzHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
