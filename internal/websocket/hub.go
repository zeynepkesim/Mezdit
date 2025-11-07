package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	MessageTypePostCreated MessageType = "POST_CREATED"
	MessageTypePostUpdated MessageType = "POST_UPDATED"
	MessageTypePostVoted   MessageType = "POST_VOTED"
	MessageTypePostPrivacy MessageType = "POST_PRIVACY_CHANGED"
)

// Message represents a WebSocket message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Global hub instance
var GlobalHub *Hub

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("Client disconnected. Total clients: %d", len(h.Clients))
			}
			h.mu.Unlock()

		case message, ok := <-h.Broadcast:
			if !ok {
				// Channel closed - stop the hub run loop
				log.Println("Broadcast channel closed, stopping hub")
				return
			}

			h.mu.RLock()
			messageJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				h.mu.RUnlock()
				continue
			}

			log.Printf("Broadcasting message type: %s to %d clients", message.Type, len(h.Clients))
			for client := range h.Clients {
				select {
				case client.Send <- messageJSON:
				default:
					// Client is not reading, close the connection
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(messageType MessageType, payload interface{}) {
	message := Message{
		Type:    messageType,
		Payload: payload,
	}
	h.Broadcast <- message
}

// InitHub initializes the global hub and starts it
func InitHub() {
	GlobalHub = NewHub()
	go GlobalHub.Run()
	log.Println("WebSocket Hub initialized and running")
}
