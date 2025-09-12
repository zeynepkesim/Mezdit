import { ref } from 'vue'

const notifications = ref([])

export function useNotifications() {
  function showNotification(message, type = 'success') {
    const id = Date.now()
    const notification = {
      id,
      message,
      type,
      timestamp: new Date()
    }
    
    notifications.value.push(notification)
    
    // Auto remove after 3 seconds
    setTimeout(() => {
      removeNotification(id)
    }, 3000)
    
    return id
  }
  
  function removeNotification(id) {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }
  
  function showSuccess(message) {
    return showNotification(message, 'success')
  }
  
  function showError(message) {
    return showNotification(message, 'error')
  }
  
  return {
    notifications,
    showNotification,
    showSuccess,
    showError,
    removeNotification
  }
}
