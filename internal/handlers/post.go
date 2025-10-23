package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"devops-be/internal/database"
	"devops-be/internal/middleware"
	"devops-be/internal/models"
)

// PostsHandler handles GET /posts and POST /posts
func PostsHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
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
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid userId parameter")
				return
			}
		}

		posts, err := database.GetPostsForUser(userID)
		if err != nil {
			log.Printf("Error getting posts: %v", err)
			middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve posts")
			return
		}
		middleware.WriteOK(w, posts)
		return
	}

	if r.Method == http.MethodPost {
		// Try new format first
		var body models.CreatePostRequest

		// Read body and try to decode as new format
		if middleware.DecodeJSONBody(w, r, &body) && body.PublisherID > 0 {
			// Log the incoming request
			log.Printf("=== POST CREATE REQUEST ===")
			log.Printf("PublisherID: %d", body.PublisherID)
			log.Printf("PublisherName: '%s'", body.PublisherName)
			log.Printf("Name: '%s'", body.Name)
			log.Printf("Header: '%s'", body.Header)
			log.Printf("ContentText: '%s'", body.ContentText)
			log.Printf("ContentPhotoID: %v", body.ContentPhotoID)
			log.Printf("IsPrivate: %v", body.IsPrivate)
			log.Printf("==========================")
			body.Header = strings.TrimSpace(body.Header)
			body.ContentText = strings.TrimSpace(body.ContentText)
			body.PublisherName = strings.TrimSpace(body.PublisherName)
			body.Name = strings.TrimSpace(body.Name)

			// If publisherName is empty but name is provided, use name as display name
			effectivePublisherName := body.PublisherName
			if effectivePublisherName == "" && body.Name != "" {
				effectivePublisherName = body.Name

				log.Printf("Effective Publisher Name: '%s'", effectivePublisherName)
			}

			if body.Header == "" {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Header cannot be empty")
				return
			}

			if body.ContentText == "" {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
				return
			}

			post, err := database.AddPost(body.PublisherID, effectivePublisherName, body.Header, body.ContentText, body.ContentPhotoID, body.IsPrivate)
			if err != nil {
				if strings.Contains(err.Error(), "publisher not found") {
					middleware.WriteError(w, http.StatusBadRequest, "invalid_publisher", "Publisher user does not exist")
					return
				}
				log.Printf("Error creating post: %v", err)
				middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to create post")
				return
			}
			middleware.WriteCreated(w, post)
			return
		}

		// If new format fails, try legacy format (this is a fallback)
		// Note: Since we already consumed the body above, we need to handle this differently
		// For now, we'll require the new format
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid request format. Use: {\"publisherId\": number, \"publisherName\": \"string\", \"header\": \"string\", \"contentText\": \"string\"}")
		return
	}

	middleware.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET and POST methods are allowed")
}

// EditPostHandler handles POST /posts/edit
func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var body models.EditPostRequest
	if !middleware.DecodeJSONBody(w, r, &body) {
		return
	}

	// Trim whitespace
	body.Header = strings.TrimSpace(body.Header)
	body.ContentText = strings.TrimSpace(body.ContentText)
	body.PublisherName = strings.TrimSpace(body.PublisherName)

	// Validate input
	if body.PostID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Post ID must be a positive integer")
		return
	}
	if body.UserID <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "User ID must be a positive integer")
		return
	}
	if body.Header == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Header cannot be empty")
		return
	}
	if body.ContentText == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content text cannot be empty")
		return
	}

	post, err := database.EditPost(body.PostID, body.UserID, body.Header, body.ContentText, body.PublisherName, body.IsPrivate)
	if err != nil {
		if err.Error() == "post not found" {
			middleware.WriteError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "user not found" {
			middleware.WriteError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		if err.Error() == "you are not the owner of this post" {
			middleware.WriteError(w, http.StatusForbidden, "not_post_owner", "You are not the owner of this post")
			return
		}
		log.Printf("Error editing post: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to edit post")
		return
	}

	middleware.WriteOK(w, post)
}