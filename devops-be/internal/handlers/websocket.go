package handlers

import (
	"log"
	"net/http"

	ws "devops-be/internal/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (for development)
		// In production, you should validate the origin
		return true
	},
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=== WebSocket Handler Called ===")
	log.Printf("Remote Address: %s", r.RemoteAddr)
	log.Printf("Request URL: %s", r.URL.String())
	log.Printf("Request Method: %s", r.Method)

	// Log all headers
	log.Println("All Request Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("  %s: %s", name, value)
		}
	}

	// Check specific upgrade headers
	upgrade := r.Header.Get("Upgrade")
	connection := r.Header.Get("Connection")
	log.Printf("Upgrade header: '%s'", upgrade)
	log.Printf("Connection header: '%s'", connection)

	if upgrade == "" || connection == "" {
		log.Println("ERROR: Missing Upgrade or Connection headers!")
		http.Error(w, "WebSocket upgrade headers missing", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}

	log.Println("✓ WebSocket upgrade successful!")

	// Create new client
	client := &ws.Client{
		Hub:  ws.GlobalHub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// Register client with the hub
	log.Println("Registering client with hub...")
	ws.GlobalHub.Register <- client
	log.Println("Client registered!")

	// Start client's goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Printf("New WebSocket client connected. Total clients in hub: ???") // You might need to add a counter in your hub
}
