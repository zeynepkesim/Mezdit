import { ref, onUnmounted } from 'vue'

const WS_URL = 'ws://localhost:8080/ws'

export function useWebSocket() {
  const ws = ref(null)
  const isConnected = ref(false)
  const reconnectDelay = ref(1000)
  const maxReconnectDelay = 30000
  const reconnectTimeout = ref(null)

  const connect = (onMessage) => {
    try {
      ws.value = new WebSocket(WS_URL)

      ws.value.onopen = () => {
        console.log('✓ WebSocket connected')
        isConnected.value = true
        reconnectDelay.value = 1000 // Reset reconnect delay on successful connection
      }

      ws.value.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          console.log('WebSocket message received:', message.type, message.payload)
          if (onMessage) {
            onMessage(message)
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      ws.value.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      ws.value.onclose = () => {
        console.log('WebSocket disconnected. Reconnecting...')
        isConnected.value = false

        // Auto-reconnect with exponential backoff
        if (reconnectTimeout.value) {
          clearTimeout(reconnectTimeout.value)
        }

        reconnectTimeout.value = setTimeout(() => {
          connect(onMessage)
        }, reconnectDelay.value)

        // Increase reconnect delay exponentially, up to max
        reconnectDelay.value = Math.min(reconnectDelay.value * 2, maxReconnectDelay)
      }
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
    }
  }

  const disconnect = () => {
    if (reconnectTimeout.value) {
      clearTimeout(reconnectTimeout.value)
      reconnectTimeout.value = null
    }

    if (ws.value) {
      ws.value.close()
      ws.value = null
      isConnected.value = false
    }
  }

  // Cleanup on component unmount
  onUnmounted(() => {
    disconnect()
  })

  return {
    connect,
    disconnect,
    isConnected
  }
}
