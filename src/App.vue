<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { logout, getCurrentUser } from "./store/auth.js";

const route = useRoute();
const router = useRouter();

// Reactive currentUser - her değişiklikte güncellenir
const currentUser = computed(() => getCurrentUser());

// Login sayfası kontrolü
const isLoginPage = computed(
  () => route.name === "login" || route.name === "register"
);

// Backend'den gelen profil fotoğrafını kullan, yoksa Dicebear avatar'ı göster
const getUserAvatar = (user) => {
  if (user?.profilePhoto) {
    return user.profilePhoto;
  }
  if (user?.profilePhotoPath) {
    return user.profilePhotoPath;
  }
  if (user?.avatar) {
    return user.avatar;
  }
  if (user?.photo) {
    return user.photo;
  }
  // Fallback: Dicebear avatar
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(
    user?.username || "user"
  )}`;
};

function handleLogout() {
  logout();
  router.push("/");
}
</script>

<template>
  <v-app>
    <!-- App Bar (Header) -->
    <v-app-bar v-if="!isLoginPage" color="white" elevation="1" border>
      <v-app-bar-title>
        <v-btn
          variant="text"
          @click="router.push('/posts')"
          class="text-h6 font-weight-bold"
          color="primary"
        >
          <v-icon start>mdi-flash</v-icon>
          Mezdit
        </v-btn>
      </v-app-bar-title>

      <v-spacer />

      <v-btn
        color="primary"
        variant="elevated"
        prepend-icon="mdi-plus"
        @click="router.push('/create-post')"
        class="mr-4"
      >
        Create Post
      </v-btn>

      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn v-bind="props" icon size="large" class="user-avatar-btn">
            <v-avatar size="40">
              <v-img
                :src="getUserAvatar(currentUser)"
                :alt="currentUser?.name || currentUser?.username"
              />
            </v-avatar>
          </v-btn>
        </template>

        <v-list min-width="200">
          <v-list-item class="px-4 py-3">
            <template v-slot:prepend>
              <v-avatar size="48">
                <v-img
                  :src="getUserAvatar(currentUser)"
                  :alt="currentUser?.name"
                />
              </v-avatar>
            </template>
            <v-list-item-title class="font-weight-bold">
              {{ currentUser?.name }}
            </v-list-item-title>
            <v-list-item-subtitle>
              @{{ currentUser?.username }}
            </v-list-item-subtitle>
          </v-list-item>

          <v-divider />

          <v-list-item @click="router.push('/change-password')" prepend-icon="mdi-lock-reset">
            <v-list-item-title>Change Password</v-list-item-title>
          </v-list-item>

          <v-list-item @click="handleLogout" prepend-icon="mdi-logout">
            <v-list-item-title>Logout</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <v-main v-if="!isLoginPage">
      <v-container>
        <router-view />
      </v-container>
    </v-main>

    <v-main v-else class="pa-0">
      <router-view />
    </v-main>
    <!--
   <NotificationToast />
  -->
  </v-app>
</template>

<style scoped>
.user-avatar-btn {
  border: 2px solid rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
}

.user-avatar-btn:hover {
  border-color: rgba(0, 0, 0, 0.3);
  transform: scale(1.05);
}
</style>
