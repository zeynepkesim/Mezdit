package handlers

import (
	"devops-be/internal/storage"
	"encoding/json"
	"net/http"
	"time"
)

type ImageURLRequest struct {
	ObjectPath string `json:"object_path"`
}

type ImageURLResponse struct {
	URL string `json:"url"`
}

// ImageURLHandler generates a presigned URL for accessing an image from MinIO
func ImageURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImageURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ObjectPath == "" {
		http.Error(w, "object_path is required", http.StatusBadRequest)
		return
	}

	// Generate presigned URL valid for 24 hours
	url, err := storage.GetImageURL(req.ObjectPath, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate image URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ImageURLResponse{URL: url})
}
