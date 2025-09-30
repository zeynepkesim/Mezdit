package handlers

import (
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"devops-be/internal/database"
	"devops-be/internal/middleware"
)

// UploadImageHandler handles POST /upload - standalone image upload
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid_form", "Failed to parse multipart form")
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "missing_image", "Image file is required")
		return
	}
	defer file.Close()

	// Save image
	image, err := database.SaveImage(file, header.Filename)
	if err != nil {
		log.Printf("Error saving image: %v", err)
		if strings.Contains(err.Error(), "invalid file type") {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
			return
		}
		middleware.WriteError(w, http.StatusInternalServerError, "upload_failed", "Failed to upload image")
		return
	}

	middleware.WriteOK(w, image)
}

// UploadPostHandler handles POST /posts/upload - create post with image
func UploadPostHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid_form", "Failed to parse multipart form")
		return
	}

	// Get form data
	publisherIDStr := r.FormValue("publisherId")
	publisherName := strings.TrimSpace(r.FormValue("publisherName"))
	contentText := strings.TrimSpace(r.FormValue("contentText"))

	// Validate input
	if publisherIDStr == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Publisher ID is required")
		return
	}
	if contentText == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	publisherID, err := strconv.Atoi(publisherIDStr)
	if err != nil || publisherID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid publisher ID")
		return
	}

	// Check for uploaded image
	var imageID *int
	file, header, err := r.FormFile("image")
	if err == nil {
		// Image uploaded, save it
		defer file.Close()
		image, err := database.SaveImage(file, header.Filename)
		if err != nil {
			log.Printf("Error saving image: %v", err)
			if strings.Contains(err.Error(), "invalid file type") {
				middleware.WriteError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
				return
			}
			middleware.WriteError(w, http.StatusInternalServerError, "upload_failed", "Failed to upload image")
			return
		}
		imageID = &image.ID
	}

	// Create post
	post, err := database.AddPost(publisherID, publisherName, contentText, imageID)
	if err != nil {
		if strings.Contains(err.Error(), "publisher not found") {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_publisher", "Publisher user does not exist")
			return
		}
		log.Printf("Error creating post: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to create post")
		return
	}

	// Update image post_id if image was uploaded
	if imageID != nil {
		err = database.UpdateImagePostID(*imageID, post.ID)
		if err != nil {
			log.Printf("Warning: Failed to update image post_id: %v", err)
		}
	}

	middleware.WriteCreated(w, post)
}

// EditPostWithImageHandler handles POST /posts/edit-with-image
func EditPostWithImageHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid_form_data", "Invalid multipart form data")
		return
	}

	// Extract form values
	postIDStr := r.FormValue("postId")
	userIDStr := r.FormValue("userId")
	contentText := strings.TrimSpace(r.FormValue("contentText"))
	publisherName := strings.TrimSpace(r.FormValue("publisherName"))

	// Validate required fields
	if postIDStr == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Post ID is required")
		return
	}
	if userIDStr == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "User ID is required")
		return
	}
	if contentText == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	// Convert IDs to integers
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid post ID")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid user ID")
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
		middleware.WriteError(w, http.StatusBadRequest, "invalid_image", "Error processing image file")
		return
	}
	// If err == http.ErrMissingFile, that's OK - no image provided

	// Update post with image
	var post interface{}
	if imageFilename != "" {
		post, err = database.EditPostWithImage(postID, userID, contentText, publisherName, imageFile, imageFilename)
	} else {
		post, err = database.EditPost(postID, userID, contentText, publisherName)
	}

	if err != nil {
		if err.Error() == "post not found" {
			middleware.WriteError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "you are not the owner of this post" {
			middleware.WriteError(w, http.StatusForbidden, "not_post_owner", "You are not the owner of this post")
			return
		}
		if err.Error() == "user not found" {
			middleware.WriteError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		if strings.Contains(err.Error(), "invalid file type") {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
			return
		}
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to update post")
		return
	}

	middleware.WriteOK(w, post)
}