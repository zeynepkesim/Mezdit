package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"devops-be/internal/database"
	"devops-be/internal/middleware"
	"devops-be/internal/models"
	"devops-be/internal/security"
	ws "devops-be/internal/websocket"
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

		// Get optional authenticated user from context (OptionalAuth middleware may set this)
		authUser, authOK := middleware.GetUserFromContext(r)

		if userIDStr != "" {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil || userID <= 0 {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Invalid userId parameter")
				return
			}

			// If a userId is provided in query params, require that the request is authenticated
			// and that the authenticated user's ID matches the requested userId. This prevents
			// clients from fetching private posts of other users by spoofing the userId query param.
			if !authOK || authUser.UserID != userID {
				middleware.WriteError(w, http.StatusForbidden, "forbidden", "Cannot request posts for another user")
				return
			}
		} else if authOK {
			// If no explicit userId provided but request is authenticated, show posts for the authenticated user
			userID = authUser.UserID
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
		// POST requires authentication - check for valid token
		_, ok := middleware.GetUserFromContext(r)
		if !ok {
			// No user in context - token validation failed or missing
			middleware.WriteError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
			return
		}

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

			// Sanitize inputs to prevent XSS attacks
			body.Header = security.SanitizeText(body.Header)
			body.ContentText = security.SanitizeText(body.ContentText)
			body.PublisherName = security.SanitizeInput(body.PublisherName)
			body.Name = security.SanitizeInput(body.Name)

			// Validate input lengths to prevent DoS
			if !security.ValidateLength(body.Header, security.MaxPostHeaderLength) {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Header is too long (max 500 characters)")
				return
			}
			if !security.ValidateLength(body.ContentText, security.MaxPostContentLength) {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content is too long (max 5000 characters)")
				return
			}
			if !security.ValidateLength(body.PublisherName, security.MaxNameLength) {
				middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Publisher name is too long (max 255 characters)")
				return
			}

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

			// Broadcast new post to all connected clients
			if ws.GlobalHub != nil {
				ws.GlobalHub.BroadcastMessage(ws.MessageTypePostCreated, post)
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

	// Sanitize inputs to prevent XSS attacks
	body.Header = security.SanitizeText(body.Header)
	body.ContentText = security.SanitizeText(body.ContentText)
	body.PublisherName = security.SanitizeInput(body.PublisherName)

	// Validate input lengths
	if !security.ValidateLength(body.Header, security.MaxPostHeaderLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Header is too long (max 500 characters)")
		return
	}
	if !security.ValidateLength(body.ContentText, security.MaxPostContentLength) {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Content is too long (max 5000 characters)")
		return
	}

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

	// Read current privacy state to determine if privacy changed
	var oldIsPrivate bool
	err := database.DB.QueryRow("SELECT is_private FROM posts WHERE id = $1", body.PostID).Scan(&oldIsPrivate)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.WriteError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		log.Printf("Error checking existing post privacy: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to edit post")
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

	// Broadcast appropriate message: if privacy changed, send POST_PRIVACY_CHANGED
	if ws.GlobalHub != nil {
		if oldIsPrivate != post.IsPrivate {
			ws.GlobalHub.BroadcastMessage(ws.MessageTypePostPrivacy, post)
		} else {
			ws.GlobalHub.BroadcastMessage(ws.MessageTypePostUpdated, post)
		}
	}

	middleware.WriteOK(w, post)
}
