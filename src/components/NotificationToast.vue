<script setup>
import { useNotifications } from "../composables/notifications";

const { notifications, removeNotification } = useNotifications();
</script>

<template>
  <div class="toast-container">
    <div
      v-for="notification in notifications"
      :key="notification.id"
      class="toast"
      :class="notification.type"
      @click="removeNotification(notification.id)"
    >
      <div class="toast-content">
        <div class="toast-icon">
          <span v-if="notification.type === 'success'">✓</span>
          <span v-else-if="notification.type === 'error'">✕</span>
        </div>
        <div class="toast-message">{{ notification.message }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toast {
  min-width: 300px;
  padding: 12px 16px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  animation: slideIn 0.3s ease-out;
  transition: all 0.2s;
}

.toast:hover {
  transform: translateX(-4px);
}

.toast.success {
  background: #10b981;
  color: white;
}

.toast.error {
  background: #ef4444;
  color: white;
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toast-icon {
  font-weight: bold;
  font-size: 16px;
}

.toast-message {
  flex: 1;
  font-size: 14px;
  line-height: 1.4;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>
