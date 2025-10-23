package main

import (
	"fmt"
	"log"
	"net/http"

	"devops-be/internal/database"
	"devops-be/internal/handlers"
)

func main() {
	// Initialize database connection
	database.Init()
	defer database.Close()

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
	fmt.Println("CORS enabled for all origins")
	fmt.Println("Available endpoints:")
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}