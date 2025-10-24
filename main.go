package main

import (
	"fmt"
	"log"
	"net/http"

	"devops-be/internal/database"
	"devops-be/internal/handlers"
	"devops-be/internal/websocket"
)

func main() {
	// Initialize database connection
	database.Init()
	defer database.Close()

	// Initialize WebSocket hub
	websocket.InitHub()

	// WebSocket endpoint
	http.HandleFunc("/ws", handlers.WebSocketHandler)

	// Auth endpoints
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/change-password", handlers.ChangePasswordHandler)

	// Post endpoints
	http.HandleFunc("/posts", handlers.PostsHandler)
	http.HandleFunc("/posts/edit", handlers.EditPostHandler)
	http.HandleFunc("/posts/edit-with-image", handlers.EditPostWithImageHandler)
	http.HandleFunc("/posts/upload", handlers.UploadPostHandler)

	// Image endpoints
	http.HandleFunc("/upload", handlers.UploadImageHandler)

	// Vote endpoints
	http.HandleFunc("/vote", handlers.VoteHandler)

	// Serve static images
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("WebSocket Hub initialized - Real-time updates enabled")
	fmt.Println("CORS enabled for all origins")
	fmt.Println("")
	fmt.Println("Available endpoints:")
	fmt.Println("  WS   /ws (WebSocket - Real-time updates)")
	fmt.Println("  POST /login")
	fmt.Println("  POST /register")
	fmt.Println("  POST /change-password")
	fmt.Println("  GET  /posts")
	fmt.Println("  POST /posts")
	fmt.Println("  POST /posts/edit")
	fmt.Println("  POST /posts/edit-with-image")
	fmt.Println("  POST /posts/upload")
	fmt.Println("  POST /upload")
	fmt.Println("  POST /vote")
	fmt.Println("  GET  /images/* (Static file serving)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}