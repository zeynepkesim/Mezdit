package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type apiSuccess struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiError struct {
	Success bool         `json:"success"`
	Error   apiErrorBody `json:"error"`
}

// CORS middleware
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, apiSuccess{Success: true, Data: data})
}

func writeCreated(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, apiSuccess{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, apiError{Success: false, Error: apiErrorBody{Code: code, Message: msg}})
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	// Handle OPTIONS requests for CORS
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return false
	}
	
	if r.Method != method {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return false
	}
	return true
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "invalid_content_type", "Content-Type must be application/json")
		return false
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return false
	}
	return true
}

// Request bodies

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createPostRequest struct {
	PublisherID    int    `json:"publisherId"`
	ContentText    string `json:"contentText"`
	ContentPhotoID *int   `json:"contentPhotoId"`
}

type voteRequest struct {
	PostID  int    `json:"postId"`
	UserID  int    `json:"userId"`
	Upvoted bool   `json:"upvoted"`
}

// Simplified request for testing (backward compatibility)
type createPostSimpleRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type voteSimpleRequest struct {
	ID       int    `json:"id"`
	Like     bool   `json:"like"`
	Username string `json:"username"`
}

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/users/", userHandler) // GET /users/{username}

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("CORS enabled for all origins")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /login")
	fmt.Println("  GET  /posts")
	fmt.Println("  POST /posts")
	fmt.Println("  POST /vote")
	fmt.Println("  GET  /users/{username}")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// /login endpoint
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var body loginRequest
	if !decodeJSONBody(w, r, &body) {
		return
	}

	// Trim whitespace
	body.Username = strings.TrimSpace(body.Username)
	body.Password = strings.TrimSpace(body.Password)

	// Validate input
	if body.Username == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Username is required")
		return
	}
	if body.Password == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	// Attempt login
	user, err := LoginUser(body.Username, body.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			writeError(w, http.StatusUnauthorized, "authentication_failed", "Invalid username or password")
			return
		}
		log.Printf("Login error: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "An error occurred during authentication")
		return
	}

	writeOK(w, user)
}

// /users/{username} endpoint
func userHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	// Extract username from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	username := strings.TrimSpace(path)
	
	if username == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Username is required in URL path")
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		if err.Error() == "user not found" {
			writeError(w, http.StatusNotFound, "user_not_found", "User not found")
			return
		}
		log.Printf("Error getting user: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve user information")
		return
	}

	writeOK(w, user)
}

// /posts endpoint
func postsHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		posts, err := GetPosts()
		if err != nil {
			log.Printf("Error getting posts: %v", err)
			writeError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve posts")
			return
		}
		writeOK(w, posts)
		return
	}
	
	if r.Method == http.MethodPost {
		// Try new format first
		var body createPostRequest
		
		// Read body and try to decode as new format
		if decodeJSONBody(w, r, &body) && body.PublisherID > 0 {
			body.ContentText = strings.TrimSpace(body.ContentText)
			if body.ContentText == "" {
				writeError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
				return
			}
			
			post, err := AddPost(body.PublisherID, body.ContentText, body.ContentPhotoID)
			if err != nil {
				if strings.Contains(err.Error(), "publisher not found") {
					writeError(w, http.StatusBadRequest, "invalid_publisher", "Publisher user does not exist")
					return
				}
				log.Printf("Error creating post: %v", err)
				writeError(w, http.StatusInternalServerError, "server_error", "Failed to create post")
				return
			}
			writeCreated(w, post)
			return
		}

		// If new format fails, try legacy format (this is a fallback)
		// Note: Since we already consumed the body above, we need to handle this differently
		// For now, we'll require the new format
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid request format. Use: {\"publisherId\": number, \"contentText\": \"string\"}")
		return
	}
	
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET and POST methods are allowed")
}

// /vote endpoint
func voteHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var body voteRequest
	if !decodeJSONBody(w, r, &body) {
		return
	}
	
	// Validate input
	if body.PostID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Post ID must be a positive integer")
		return
	}
	if body.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "User ID must be a positive integer")
		return
	}

	post, err := VotePost(body.PostID, body.UserID, body.Upvoted)
	if err != nil {
		if err.Error() == "post not found" {
			writeError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "user not found" {
			writeError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		log.Printf("Error voting on post: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to process vote")
		return
	}
	writeOK(w, post)
}
