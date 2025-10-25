package middleware

import (
	"context"
	"net/http"
	"strings"

	"devops-be/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// UserContext holds the authenticated user information
type UserContext struct {
	UserID   int
	Username string
}

// RequireAuth is a middleware that validates JWT token
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		EnableCORS(w)

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteError(w, http.StatusUnauthorized, "invalid_token_format", "Authorization header must be 'Bearer {token}'")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}

		// Add user info to request context
		userCtx := &UserContext{
			UserID:   claims.UserID,
			Username: claims.Username,
		}
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuth is a middleware that validates JWT token if present, but doesn't require it
func OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		EnableCORS(w)

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]

				// Validate token
				claims, err := auth.ValidateToken(tokenString)
				if err == nil {
					// Add user info to request context
					userCtx := &UserContext{
						UserID:   claims.UserID,
						Username: claims.Username,
					}
					ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// Continue without auth context
		next.ServeHTTP(w, r)
	}
}

// GetUserFromContext extracts user info from request context
func GetUserFromContext(r *http.Request) (*UserContext, bool) {
	user, ok := r.Context().Value(UserContextKey).(*UserContext)
	return user, ok
}
