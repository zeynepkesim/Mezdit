package handlers

import (
	"log"
	"net/http"

	"devops-be/internal/database"
	"devops-be/internal/middleware"
	"devops-be/internal/models"
)

// VoteHandler handles POST /vote
func VoteHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var body models.VoteRequest
	if !middleware.DecodeJSONBody(w, r, &body) {
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

	post, err := database.VotePost(body.PostID, body.UserID, body.Upvoted)
	if err != nil {
		if err.Error() == "post not found" {
			middleware.WriteError(w, http.StatusNotFound, "post_not_found", "Post not found")
			return
		}
		if err.Error() == "user not found" {
			middleware.WriteError(w, http.StatusBadRequest, "user_not_found", "User not found")
			return
		}
		log.Printf("Error voting on post: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to process vote")
		return
	}
	middleware.WriteOK(w, post)
}