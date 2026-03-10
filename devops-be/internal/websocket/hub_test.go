package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	
	if hub.Clients == nil {
		t.Error("Clients map not initialized")
	}
	
	if hub.Broadcast == nil {
		t.Error("Broadcast channel not initialized")
	}
	
	if hub.Register == nil {
		t.Error("Register channel not initialized")
	}
	
	if hub.Unregister == nil {
		t.Error("Unregister channel not initialized")
	}
	
	if len(hub.Clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(hub.Clients))
	}
}

func TestHub_BroadcastMessage(t *testing.T) {
	hub := NewHub()
	
	// Start hub in background
	go hub.Run()
	defer close(hub.Broadcast)
	
	// Give hub time to start
	time.Sleep(10 * time.Millisecond)
	
	tests := []struct {
		name        string
		messageType MessageType
		payload     interface{}
	}{
		{
			name:        "POST_CREATED message",
			messageType: MessageTypePostCreated,
			payload:     map[string]string{"id": "123"},
		},
		{
			name:        "POST_UPDATED message",
			messageType: MessageTypePostUpdated,
			payload:     map[string]int{"post_id": 456},
		},
		{
			name:        "POST_VOTED message",
			messageType: MessageTypePostVoted,
			payload:     "test payload",
		},
		{
			name:        "POST_PRIVACY_CHANGED message",
			messageType: MessageTypePostPrivacy,
			payload:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			hub.BroadcastMessage(tt.messageType, tt.payload)
			
			// Give time for message to be processed
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestHub_ClientRegistration(t *testing.T) {
	hub := NewHub()
	
	// Start hub in background
	go hub.Run()
	defer close(hub.Broadcast)
	
	// Create mock client
	client := &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}
	
	// Register client
	hub.Register <- client
	time.Sleep(20 * time.Millisecond)
	
	hub.mu.RLock()
	clientCount := len(hub.Clients)
	hub.mu.RUnlock()
	
	if clientCount != 1 {
		t.Errorf("Expected 1 client after registration, got %d", clientCount)
	}
	
	// Unregister client
	hub.Unregister <- client
	time.Sleep(20 * time.Millisecond)
	
	hub.mu.RLock()
	clientCount = len(hub.Clients)
	hub.mu.RUnlock()
	
	if clientCount != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", clientCount)
	}
}

func TestHub_MessageBroadcast(t *testing.T) {
	hub := NewHub()
	
	// Start hub in background
	go hub.Run()
	defer close(hub.Broadcast)
	
	// Create mock client
	client := &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}
	
	// Register client
	hub.Register <- client
	time.Sleep(20 * time.Millisecond)
	
	// Broadcast a message
	testPayload := map[string]string{"test": "data"}
	hub.BroadcastMessage(MessageTypePostCreated, testPayload)
	
	// Wait for message
	select {
	case msg := <-client.Send:
		var received Message
		if err := json.Unmarshal(msg, &received); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received.Type != MessageTypePostCreated {
			t.Errorf("Expected message type %s, got %s", MessageTypePostCreated, received.Type)
		}
		
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast message")
	}
	
	// Unregister client
	hub.Unregister <- client
	time.Sleep(20 * time.Millisecond)
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	
	// Start hub in background
	go hub.Run()
	defer close(hub.Broadcast)
	
	// Create multiple clients
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		clients[i] = &Client{
			Hub:  hub,
			Send: make(chan []byte, 256),
		}
		hub.Register <- clients[i]
	}
	
	time.Sleep(30 * time.Millisecond)
	
	hub.mu.RLock()
	clientCount := len(hub.Clients)
	hub.mu.RUnlock()
	
	if clientCount != 3 {
		t.Errorf("Expected 3 clients, got %d", clientCount)
	}
	
	// Broadcast message to all
	hub.BroadcastMessage(MessageTypePostUpdated, "test")
	time.Sleep(30 * time.Millisecond)
	
	// Check all clients received message
	for i, client := range clients {
		select {
		case <-client.Send:
			// Success
		default:
			t.Errorf("Client %d did not receive message", i)
		}
	}
	
	// Unregister all clients
	for _, client := range clients {
		hub.Unregister <- client
	}
	time.Sleep(30 * time.Millisecond)
	
	hub.mu.RLock()
	clientCount = len(hub.Clients)
	hub.mu.RUnlock()
	
	if clientCount != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", clientCount)
	}
}

func TestInitHub(t *testing.T) {
	// Save original hub
	originalHub := GlobalHub
	defer func() { GlobalHub = originalHub }()
	
	// Reset global hub
	GlobalHub = nil
	
	// Initialize hub
	InitHub()
	
	if GlobalHub == nil {
		t.Fatal("InitHub failed to initialize GlobalHub")
	}
	
	if GlobalHub.Clients == nil {
		t.Error("GlobalHub.Clients not initialized")
	}
	
	// Cleanup
	if GlobalHub != nil && GlobalHub.Broadcast != nil {
		close(GlobalHub.Broadcast)
		time.Sleep(20 * time.Millisecond)
	}
}

func TestMessageType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"POST_CREATED", MessageTypePostCreated, "POST_CREATED"},
		{"POST_UPDATED", MessageTypePostUpdated, "POST_UPDATED"},
		{"POST_VOTED", MessageTypePostVoted, "POST_VOTED"},
		{"POST_PRIVACY_CHANGED", MessageTypePostPrivacy, "POST_PRIVACY_CHANGED"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.msgType))
			}
		})
	}
}

func TestMessage_JSONMarshaling(t *testing.T) {
	msg := Message{
		Type:    MessageTypePostCreated,
		Payload: map[string]string{"id": "123"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Expected type %s, got %s", msg.Type, decoded.Type)
	}
}

func TestClient_Struct(t *testing.T) {
	hub := NewHub()
	client := &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}

	if client.Hub != hub {
		t.Error("Client hub not set correctly")
	}

	if client.Send == nil {
		t.Error("Client send channel not initialized")
	}

	if cap(client.Send) != 256 {
		t.Errorf("Expected send channel capacity 256, got %d", cap(client.Send))
	}
}

