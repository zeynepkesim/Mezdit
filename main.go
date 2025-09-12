package main

import (
	"encoding/json"
	"fmt"
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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
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

type createPostRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type voteRequest struct {
	ID       int    `json:"id"`
	Like     bool   `json:"like"`
	Username string `json:"username"`
}

func main() {
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/vote", voteHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// /posts endpoint
func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeOK(w, GetPosts())
		return
	}
	if r.Method == http.MethodPost {
		var body createPostRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		body.Author = strings.TrimSpace(body.Author)
		body.Content = strings.TrimSpace(body.Content)
		if body.Author == "" || body.Content == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "author and content fields cannot be empty")
			return
		}
		post := AddPost(body.Author, body.Content)
		writeCreated(w, post)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
}

// /vote endpoint
func voteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var body voteRequest
	if !decodeJSONBody(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Username) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "username field is required")
		return
	}
	if body.ID <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "id must be a positive integer")
		return
	}
	// body.Like already bool

	post := VotePost(body.ID, body.Like, body.Username)
	if post == nil {
		writeError(w, http.StatusNotFound, "not_found", "Post not found")
		return
	}
	writeOK(w, post)
}
