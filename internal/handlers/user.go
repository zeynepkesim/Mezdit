package handlers

import (
	"log"
	"net/http"
	"strings"

	"devops-be/internal/database"
	"devops-be/internal/middleware"
)

// UserHandler handles GET /users/{username}
func UserHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if !middleware.RequireMethod(w, r, http.MethodGet) {
		return
	}

	// Extract username from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	username := strings.TrimSpace(path)

	if username == "" {
		middleware.WriteError(w, http.StatusBadRequest, "validation_error", "Username is required in URL path")
		return
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.WriteError(w, http.StatusNotFound, "user_not_found", "User not found")
			return
		}
		log.Printf("Error getting user: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve user information")
		return
	}

	middleware.WriteOK(w, user)
}