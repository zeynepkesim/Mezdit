package main

import (
	"fmt"
	"log"
	"net/http"

	"devops-be/internal/database"
	"devops-be/internal/handlers"
	"devops-be/internal/middleware"
	"devops-be/internal/storage"
	"devops-be/internal/websocket"

	"github.com/joho/godotenv"
)

func main() {
	// Health and readiness endpoints
	http.HandleFunc("/healthz", handlers.HealthzHandler)
	http.HandleFunc("/readyz", handlers.ReadyzHandler)
	// Load .env file (ignore error if file doesn't exist - env vars might be set directly)
	_ = godotenv.Load()

	// Initialize database connection
	database.Init()
	defer database.Close()

	// Initialize MinIO storage (optional in test environment)
	if err := storage.InitMinIO(); err != nil {
		log.Printf("Warning: MinIO initialization failed: %v", err)
		log.Println("Continuing without MinIO - some features may not work")
	}

	// Initialize WebSocket hub
	websocket.InitHub()


	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", handlers.WebSocketHandler)

	// Auth endpoints (public - no authentication required)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/register", handlers.RegisterHandler)

	// Protected auth endpoints
	mux.HandleFunc("/change-password", middleware.RequireAuth(handlers.ChangePasswordHandler))

	// Post endpoints
	mux.HandleFunc("/posts", middleware.OptionalAuth(handlers.PostsHandler)) // GET is public, POST requires auth
	mux.HandleFunc("/posts/edit", middleware.RequireAuth(handlers.EditPostHandler))
	mux.HandleFunc("/posts/edit-with-image", middleware.RequireAuth(handlers.EditPostWithImageHandler))
	mux.HandleFunc("/posts/upload", middleware.RequireAuth(handlers.UploadPostHandler))

	// Image endpoints
	mux.HandleFunc("/upload", middleware.RequireAuth(handlers.UploadImageHandler))
	mux.HandleFunc("/image-url", handlers.ImageURLHandler) // Public endpoint for getting presigned URLs

	// Vote endpoints
	mux.HandleFunc("/vote", middleware.RequireAuth(handlers.VoteHandler))

	// Health and readiness endpoints
	mux.HandleFunc("/healthz", handlers.HealthzHandler)
	mux.HandleFunc("/readyz", handlers.ReadyzHandler)

	// Wrap with metrics middleware
	handler := handlers.MetricsMiddleware(mux)

	// Metrics server (separate port for Week 11 requirement)
	metricsServeMux := http.NewServeMux()
	metricsServeMux.HandleFunc("/metrics", handlers.MetricsHandler)
	go func() {
		fmt.Println("📊 Metrics server running on http://localhost:9090/metrics")
		if err := http.ListenAndServe(":9090", metricsServeMux); err != nil {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()

	fmt.Println("🚀 Server running on http://localhost:8080")
	fmt.Println("WebSocket Hub initialized - Real-time updates enabled")
	fmt.Println("JWT Authentication enabled - Token expiry: 24 hours")
	fmt.Println("CORS enabled for all origins")
	fmt.Println("")
	fmt.Println("Available endpoints:")
	fmt.Println("  🔓 Public endpoints:")
	fmt.Println("    WS   /ws (WebSocket - Real-time updates)")
	fmt.Println("    POST /login (Returns JWT token)")
	fmt.Println("    POST /register (Returns JWT token)")
	fmt.Println("    GET  /posts (Public posts only)")
	fmt.Println("    GET  /images/* (Static file serving)")
	fmt.Println("")
	fmt.Println("  🔐 Protected endpoints (Require Authorization: Bearer <token>):")
	fmt.Println("    POST /change-password")
	fmt.Println("    POST /posts (Create post)")
	fmt.Println("    POST /posts/edit")
	fmt.Println("    POST /posts/edit-with-image")
	fmt.Println("    POST /posts/upload")
	fmt.Println("    POST /upload (Image upload)")
	fmt.Println("    POST /vote")
	fmt.Println("")
	fmt.Println("  📊 Metrics endpoint (separate port):")
	fmt.Println("    GET  http://localhost:9090/metrics")
	log.Fatal(http.ListenAndServe(":8080", handler))
}