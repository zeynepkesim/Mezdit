package websocket

import (
	"testing"
	"time"
)

func TestClient_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    time.Duration
		expected time.Duration
	}{
		{"writeWait", writeWait, 10 * time.Second},
		{"pongWait", pongWait, 60 * time.Second},
		{"pingPeriod", pingPeriod, (pongWait * 9) / 10},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.value)
			}
		})
	}
}

func TestClient_MaxMessageSize(t *testing.T) {
	expectedSize := int64(512)
	if maxMessageSize != expectedSize {
		t.Errorf("Expected maxMessageSize to be %d, got %d", expectedSize, maxMessageSize)
	}
}

func TestClient_PingPeriodCalculation(t *testing.T) {
	// Verify pingPeriod is less than pongWait
	if pingPeriod >= pongWait {
		t.Errorf("pingPeriod (%v) should be less than pongWait (%v)", pingPeriod, pongWait)
	}
	
	// Verify the exact calculation
	expected := (pongWait * 9) / 10
	if pingPeriod != expected {
		t.Errorf("Expected pingPeriod to be %v, got %v", expected, pingPeriod)
	}
}

func TestClient_Creation(t *testing.T) {
	hub := NewHub()
	
	client := &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}
	
	if client.Hub == nil {
		t.Error("Client hub is nil")
	}
	
	if client.Send == nil {
		t.Error("Client send channel is nil")
	}
	
	if client.Hub != hub {
		t.Error("Client hub does not match expected hub")
	}
}

func TestClient_SendChannelCapacity(t *testing.T) {
	client := &Client{
		Hub:  NewHub(),
		Send: make(chan []byte, 256),
	}
	
	// Test that we can send up to 256 messages without blocking
	for i := 0; i < 256; i++ {
		select {
		case client.Send <- []byte("test"):
			// Success
		default:
			t.Fatalf("Channel blocked at message %d (expected capacity 256)", i)
		}
	}
	
	// The 257th message should block (or we can check it doesn't accept immediately)
	select {
	case client.Send <- []byte("test"):
		t.Error("Expected channel to be full after 256 messages")
	default:
		// Expected behavior - channel is full
	}
}

func TestClient_ChannelClose(t *testing.T) {
	client := &Client{
		Hub:  NewHub(),
		Send: make(chan []byte, 256),
	}
	
	// Close the channel
	close(client.Send)
	
	// Verify we can't send after close (would panic if not handled)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when sending to closed channel")
		}
	}()
	
	client.Send <- []byte("test")
}

func TestClient_StructFields(t *testing.T) {
	hub := NewHub()
	sendChan := make(chan []byte, 256)
	
	client := &Client{
		Hub:  hub,
		Send: sendChan,
	}
	
	// Verify all fields are properly assigned
	if client.Hub != hub {
		t.Error("Hub field not properly assigned")
	}
	
	if client.Send != sendChan {
		t.Error("Send field not properly assigned")
	}
	
	// Conn field should be nil since we haven't set it
	if client.Conn != nil {
		t.Error("Expected Conn to be nil for new client")
	}
}

