package handlers

import (
	"net/http"
	"devops-be/internal/database"
)

// HealthzHandler returns 200 OK if the app is running (no external deps)
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ReadyzHandler returns 200 OK if the app is ready (DB reachable)
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if database.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("db not ready"))
		return
	}
	if err := database.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("db not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
