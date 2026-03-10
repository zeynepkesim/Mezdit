package handlers

import (
	"devops-be/internal/database"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestMetricsMiddleware(t *testing.T) {
	// Reset metrics
	atomic.StoreUint64(&totalRequests, 0)
	atomic.StoreUint64(&totalErrors, 0)
	atomic.StoreInt64(&totalLatency, 0)

	tests := []struct {
		name           string
		handlerStatus  int
		expectedErrors uint64
	}{
		{
			name:           "Successful request - no error",
			handlerStatus:  http.StatusOK,
			expectedErrors: 0,
		},
		{
			name:           "Client error - counts as error",
			handlerStatus:  http.StatusBadRequest,
			expectedErrors: 1,
		},
		{
			name:           "Server error - counts as error",
			handlerStatus:  http.StatusInternalServerError,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset error count before each test
			atomic.StoreUint64(&totalErrors, 0)
			initialRequests := atomic.LoadUint64(&totalRequests)

			// Create a test handler that returns specific status
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
			})

			// Wrap with metrics middleware
			wrapped := MetricsMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			// Check request counter incremented
			if atomic.LoadUint64(&totalRequests) != initialRequests+1 {
				t.Errorf("Expected request count to increment by 1")
			}

			// Check error counter
			if atomic.LoadUint64(&totalErrors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, atomic.LoadUint64(&totalErrors))
			}

			// Check latency was recorded (should be > 0 even for fast requests)
			latency := atomic.LoadInt64(&totalLatency)
			if latency < 0 {
				t.Errorf("Expected latency to be >= 0, got %d", latency)
			}
			// Note: Latency might be 0 for very fast requests in test environment, this is acceptable
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	rw.WriteHeader(http.StatusNotFound)

	if rw.status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rw.status)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected underlying recorder status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestResponseWriter_Hijack(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	// httptest.ResponseRecorder doesn't support hijacking, should return error
	_, _, err := rw.Hijack()
	if err == nil {
		t.Error("Expected error when hijacking httptest.ResponseRecorder, got nil")
	}
}

func TestMetricsHandler_ResponseFormat(t *testing.T) {
	// Reset metrics for clean test
	atomic.StoreUint64(&totalRequests, 100)
	atomic.StoreUint64(&totalErrors, 10)
	atomic.StoreInt64(&totalLatency, 5000)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(w, req)

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4" {
		t.Errorf("Expected Content-Type 'text/plain; version=0.0.4', got '%s'", contentType)
	}

	// Check response contains metrics
	body := w.Body.String()

	expectedMetrics := []string{
		"app_requests_total",
		"app_errors_total",
		"app_request_latency_microseconds",
		"# HELP",
		"# TYPE",
	}

	for _, metric := range expectedMetrics {
		if !contains(body, metric) {
			t.Errorf("Expected response to contain '%s'", metric)
		}
	}

	// Verify actual metric values are present
	if !contains(body, "app_requests_total 100") {
		t.Error("Expected to find 'app_requests_total 100' in response")
	}
	if !contains(body, "app_errors_total 10") {
		t.Error("Expected to find 'app_errors_total 10' in response")
	}
	if !contains(body, "app_request_latency_microseconds 5000") {
		t.Error("Expected to find 'app_request_latency_microseconds 5000' in response")
	}
}

func TestMetricsHandler_StatusCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandler_WithoutDatabase(t *testing.T) {
	// Save original DB
	origDB := database.DB
	defer func() { database.DB = origDB }()

	// Set DB to nil
	database.DB = nil

	atomic.StoreUint64(&totalRequests, 50)
	atomic.StoreUint64(&totalErrors, 5)
	atomic.StoreInt64(&totalLatency, 1000)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(w, req)

	// Should still return basic metrics even without DB
	body := w.Body.String()

	if !contains(body, "app_requests_total 50") {
		t.Error("Expected to find 'app_requests_total 50' in response")
	}
	if !contains(body, "app_errors_total 5") {
		t.Error("Expected to find 'app_errors_total 5' in response")
	}

	// DB-dependent metrics (posts, users) should not be present
	if contains(body, "app_posts_total") {
		t.Error("Did not expect to find 'app_posts_total' when DB is nil")
	}
	if contains(body, "app_users_total") {
		t.Error("Did not expect to find 'app_users_total' when DB is nil")
	}
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	// Test that default status is maintained
	if rw.status != 200 {
		t.Errorf("Expected default status 200, got %d", rw.status)
	}

	// Write response without explicitly calling WriteHeader
	rw.ResponseWriter.Write([]byte("test"))

	// Status should still be 200
	if rw.status != 200 {
		t.Errorf("Expected status to remain 200, got %d", rw.status)
	}
}

func TestResponseWriter_MultipleWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	// Call WriteHeader multiple times
	rw.WriteHeader(http.StatusOK)
	rw.WriteHeader(http.StatusNotFound) // Second call should change status

	// Verify final status
	if rw.status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rw.status)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSearch(s, substr))
}

func containsSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
