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

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}

	// Create new client
	client := &ws.Client{
		Hub:  ws.GlobalHub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// Register client with the hub
	ws.GlobalHub.Register <- client

	// Start client's goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Println("New WebSocket client connected")
}
