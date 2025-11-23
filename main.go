package main

import (
	"fmt"
	"log"
	"net/http"

	"devops-be/internal/database"
	"devops-be/internal/handlers"
	"devops-be/internal/middleware"
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

	// Vote endpoints
	mux.HandleFunc("/vote", middleware.RequireAuth(handlers.VoteHandler))

	// Serve static images
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

	// Health and readiness endpoints
	mux.HandleFunc("/healthz", handlers.HealthzHandler)
	mux.HandleFunc("/readyz", handlers.ReadyzHandler)

	// Metrics endpoint (on main port for now)
	mux.HandleFunc("/metrics", handlers.MetricsHandler)

	// Wrap with metrics middleware
	handler := handlers.MetricsMiddleware(mux)

	fmt.Println("Server running on http://localhost:8080")
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
	log.Fatal(http.ListenAndServe(":8080", handler))
}