package handlers

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
	"devops-be/internal/database"
)

var (
	totalRequests uint64
	totalErrors uint64
	totalLatency int64 // in microseconds
)

// MetricsHandler exposes Prometheus-style metrics
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP app_requests_total Total HTTP requests\n")
	fmt.Fprintf(w, "# TYPE app_requests_total counter\n")
	fmt.Fprintf(w, "app_requests_total %d\n", atomic.LoadUint64(&totalRequests))

	fmt.Fprintf(w, "# HELP app_errors_total Total HTTP errors\n")
	fmt.Fprintf(w, "# TYPE app_errors_total counter\n")
	fmt.Fprintf(w, "app_errors_total %d\n", atomic.LoadUint64(&totalErrors))

	fmt.Fprintf(w, "# HELP app_request_latency_microseconds Total request latency in microseconds\n")
	fmt.Fprintf(w, "# TYPE app_request_latency_microseconds counter\n")
	fmt.Fprintf(w, "app_request_latency_microseconds %d\n", atomic.LoadInt64(&totalLatency))

	// Custom metric: total posts
	posts, err := database.GetPosts()
	if err == nil {
		fmt.Fprintf(w, "# HELP app_posts_total Total number of posts\n")
		fmt.Fprintf(w, "# TYPE app_posts_total gauge\n")
		fmt.Fprintf(w, "app_posts_total %d\n", len(posts))
	}

	// Custom metric: total users
	users, err := database.GetAllUsers()
	if err == nil {
		fmt.Fprintf(w, "# HELP app_users_total Total number of users\n")
		fmt.Fprintf(w, "# TYPE app_users_total gauge\n")
		fmt.Fprintf(w, "app_users_total %d\n", len(users))
	}
}

// MetricsMiddleware tracks requests, errors, and latency
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		atomic.AddUint64(&totalRequests, 1)
		ww := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		if ww.status >= 400 {
			atomic.AddUint64(&totalErrors, 1)
		}
		atomic.AddInt64(&totalLatency, int64(time.Since(start).Microseconds()))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
