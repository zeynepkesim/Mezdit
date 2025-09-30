package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"devops-be/internal/models"
)

// EnableCORS adds CORS headers
func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	EnableCORS(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteOK writes a successful response
func WriteOK(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, models.APISuccess{Success: true, Data: data})
}

// WriteCreated writes a created response
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, models.APISuccess{Success: true, Data: data})
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, models.APIError{Success: false, Error: models.APIErrorBody{Code: code, Message: msg}})
}

// RequireMethod checks if the request method matches the required method
func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	// Handle OPTIONS requests for CORS
	if r.Method == http.MethodOptions {
		EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return false
	}

	if r.Method != method {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return false
	}
	return true
}

// DecodeJSONBody decodes JSON request body
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		WriteError(w, http.StatusUnsupportedMediaType, "invalid_content_type", "Content-Type must be application/json")
		return false
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return false
	}
	return true
}