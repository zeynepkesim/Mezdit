<script setup>
import NotificationToast from './components/NotificationToast.vue'
</script>

<template>
  <header class="topbar">
    <nav class="nav">
      <router-link class="brand" to="/">Reddit Demo</router-link>
      <router-link class="create" to="/create">Create Post</router-link>
    </nav>
  </header>
  <main class="content">
    <router-view />
  </main>
  <NotificationToast />
</template>

<style scoped>
.topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffffcc;
  backdrop-filter: saturate(180%) blur(8px);
}
.nav {
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}
.brand {
  font-weight: 800;
  color: #0f172a;
}
.create {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  background: #0ea5e9;
  color: white;
  border-radius: 8px;
}
.content {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px 16px;
}
</style>
