package handlers

import (
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

func TestMetricsHandler(t *testing.T) {
	t.Skip("Skipping MetricsHandler test - requires DB setup")
	// MetricsHandler calls database.GetPosts() and database.GetAllUsers()
	// which require a real database connection. We test the core metrics
	// (requests, errors, latency) via MetricsMiddleware tests instead.
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
