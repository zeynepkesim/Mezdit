package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
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

type createUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type createPostRequest struct {
	PublisherID    int    `json:"publisherId"`
	PublisherName  string `json:"publisherName"`
	Name           string `json:"name"`
	ContentText    string `json:"contentText"`
	ContentPhotoID *int   `json:"contentPhotoId"`
}

type voteRequest struct {
	PostID  int    `json:"postId"`
	UserID  int    `json:"userId"`
	Upvoted bool   `json:"upvoted"`
}

type editPostRequest struct {
	PostID        int    `json:"postId"`
	UserID        int    `json:"userId"`
	ContentText   string `json:"contentText"`
	PublisherName string `json:"publisherName"`
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
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/posts/edit", editPostHandler)
	http.HandleFunc("/posts/edit-with-image", editPostWithImageHandler)
	http.HandleFunc("/posts/upload", uploadPostHandler)
	http.HandleFunc("/upload", uploadImageHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/users/", userHandler) // GET /users/{username}

	// Serve static images
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("CORS enabled for all origins")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /login")
	fmt.Println("  POST /register")
	fmt.Println("  GET  /posts")
	fmt.Println("  POST /posts")
	fmt.Println("  POST /posts/edit")
	fmt.Println("  POST /posts/edit-with-image")
	fmt.Println("  POST /posts/upload")
	fmt.Println("  POST /upload")
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

// /register endpoint
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var body createUserRequest
	if !decodeJSONBody(w, r, &body) {
		return
	}

	// Trim whitespace
	body.Name = strings.TrimSpace(body.Name)
	body.Username = strings.TrimSpace(body.Username)
	body.Password = strings.TrimSpace(body.Password)

	// Validate input
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if body.Username == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Username is required")
		return
	}
	if body.Password == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	// Attempt to create user
	user, err := CreateUser(body.Name, body.Username, body.Password)
	if err != nil {
		if err.Error() == "username already exists" {
			writeError(w, http.StatusConflict, "username_exists", "Username is already taken")
			return
		}
		log.Printf("Registration error: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "An error occurred during registration")
		return
	}

	writeCreated(w, user)
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
		// Check for userId query parameter to get user-specific view
		userIDStr := r.URL.Query().Get("userId")
		var userID int = 0

		if userIDStr != "" {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil || userID <= 0 {
				writeError(w, http.StatusBadRequest, "validation_error", "Invalid userId parameter")
				return
			}
		}

		posts, err := GetPostsForUser(userID)
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
			// Log the incoming request
			log.Printf("=== POST CREATE REQUEST ===")
			log.Printf("PublisherID: %d", body.PublisherID)
			log.Printf("PublisherName: '%s'", body.PublisherName)
			log.Printf("Name: '%s'", body.Name)
			log.Printf("ContentText: '%s'", body.ContentText)
			log.Printf("ContentPhotoID: %v", body.ContentPhotoID)
			log.Printf("==========================")
			body.ContentText = strings.TrimSpace(body.ContentText)
			body.PublisherName = strings.TrimSpace(body.PublisherName)
			body.Name = strings.TrimSpace(body.Name)

			// If publisherName is empty but name is provided, use name as display name
			effectivePublisherName := body.PublisherName
			if effectivePublisherName == "" && body.Name != "" {
				effectivePublisherName = body.Name
			
			log.Printf("Effective Publisher Name: '%s'", effectivePublisherName)
			}
			
			if body.ContentText == "" {
				writeError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
				return
			}
			
			post, err := AddPost(body.PublisherID, effectivePublisherName, body.ContentText, body.ContentPhotoID)
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
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid request format. Use: {\"publisherId\": number, \"publisherName\": \"string\", \"contentText\": \"string\"}")
		return
	}
	
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET and POST methods are allowed")
}

// /posts/edit endpoint
func editPostHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var body editPostRequest
	if !decodeJSONBody(w, r, &body) {
		return
	}

	// Trim whitespace
	body.ContentText = strings.TrimSpace(body.ContentText)
	body.PublisherName = strings.TrimSpace(body.PublisherName)

	// Validate input
	if body.PostID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Post ID must be a positive integer")
		return
	}
	if body.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "User ID must be a positive integer")
		return
	}
	if body.ContentText == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	post, err := EditPost(body.PostID, body.UserID, body.ContentText, body.PublisherName)
	if err != nil {
		if err.Error() == "post not found" {
			writeError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "user not found" {
			writeError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		if err.Error() == "you are not the owner of this post" {
			writeError(w, http.StatusForbidden, "not_post_owner", "You are not the owner of this post")
			return
		}
		log.Printf("Error editing post: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to edit post")
		return
	}

	writeOK(w, post)
}

// /upload endpoint - standalone image upload
func uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_form", "Failed to parse multipart form")
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing_image", "Image file is required")
		return
	}
	defer file.Close()

	// Save image
	image, err := SaveImage(file, header.Filename)
	if err != nil {
		log.Printf("Error saving image: %v", err)
		if strings.Contains(err.Error(), "invalid file type") {
			writeError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "upload_failed", "Failed to upload image")
		return
	}

	writeOK(w, image)
}

// /posts/upload endpoint - create post with image
func uploadPostHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_form", "Failed to parse multipart form")
		return
	}

	// Get form data
	publisherIDStr := r.FormValue("publisherId")
	publisherName := strings.TrimSpace(r.FormValue("publisherName"))
	contentText := strings.TrimSpace(r.FormValue("contentText"))

	// Validate input
	if publisherIDStr == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Publisher ID is required")
		return
	}
	if contentText == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	publisherID, err := strconv.Atoi(publisherIDStr)
	if err != nil || publisherID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid publisher ID")
		return
	}

	// Check for uploaded image
	var imageID *int
	file, header, err := r.FormFile("image")
	if err == nil {
		// Image uploaded, save it
		defer file.Close()
		image, err := SaveImage(file, header.Filename)
		if err != nil {
			log.Printf("Error saving image: %v", err)
			if strings.Contains(err.Error(), "invalid file type") {
				writeError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "upload_failed", "Failed to upload image")
			return
		}
		imageID = &image.ID
	}

	// Create post
	post, err := AddPost(publisherID, publisherName, contentText, imageID)
	if err != nil {
		if strings.Contains(err.Error(), "publisher not found") {
			writeError(w, http.StatusBadRequest, "invalid_publisher", "Publisher user does not exist")
			return
		}
		log.Printf("Error creating post: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create post")
		return
	}

	// Update image post_id if image was uploaded
	if imageID != nil {
		err = UpdateImagePostID(*imageID, post.ID)
		if err != nil {
			log.Printf("Warning: Failed to update image post_id: %v", err)
		}
	}

	writeCreated(w, post)
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

func editPostWithImageHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_form_data", "Invalid multipart form data")
		return
	}

	// Extract form values
	postIDStr := r.FormValue("postId")
	userIDStr := r.FormValue("userId")
	contentText := strings.TrimSpace(r.FormValue("contentText"))
	publisherName := strings.TrimSpace(r.FormValue("publisherName"))

	// Validate required fields
	if postIDStr == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Post ID is required")
		return
	}
	if userIDStr == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "User ID is required")
		return
	}
	if contentText == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	// Convert IDs to integers
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid post ID")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid user ID")
		return
	}

	// Handle image file (optional)
	var imageFile multipart.File
	var imageFilename string

	file, fileHeader, err := r.FormFile("image")
	if err == nil {
		// Image was provided
		imageFile = file
		imageFilename = fileHeader.Filename
		defer file.Close()
	} else if err != http.ErrMissingFile {
		// Error other than missing file
		writeError(w, http.StatusBadRequest, "invalid_image", "Error processing image file")
		return
	}
	// If err == http.ErrMissingFile, that's OK - no image provided

	// Update post with image
	post, err := EditPostWithImage(postID, userID, contentText, publisherName, imageFile, imageFilename)
	if err != nil {
		if err.Error() == "post not found" {
			writeError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "you are not the owner of this post" {
			writeError(w, http.StatusForbidden, "not_post_owner", "You are not the owner of this post")
			return
		}
		if err.Error() == "user not found" {
			writeError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		if strings.Contains(err.Error(), "invalid file type") {
			writeError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update post")
		return
	}

	writeOK(w, post)
}
