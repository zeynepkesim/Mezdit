<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { posts, loading, error, loadPosts } from '../store/posts'
import { getCurrentUser } from '../store/auth.js'

const router = useRouter()
const currentUser = getCurrentUser()

// User avatar helper (not used in template now, kept for future use)
const getUserAvatar = (user) => {
  if (user?.profilePhoto) return user.profilePhoto
  if (user?.profilePhotoPath) return user.profilePhotoPath
  if (user?.avatar) return user.avatar
  if (user?.photo) return user.photo
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(user?.username || 'user')}`
}

// Use publisherPhoto from backend if present
const getPostAuthorAvatar = (post) => {
  if (post?.publisherPhoto) return post.publisherPhoto
  if (post?.publisherPhotoPath) return post.publisherPhotoPath
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(post?.publisherName || 'anon')}`
}

onMounted(async () => {
  await loadPosts()
})

function goToCreatePost() {
  router.push('/create-post')
}
</script>

<template>
  <div>
    <v-container class="mt-4">
      <!-- User Welcome Card -->
      <v-row justify="center" class="mb-4">
        <v-col cols="12" md="8">
          <v-card variant="tonal" color="primary" class="text-center">
            <v-card-text>
              <v-icon class="mr-2">mdi-account-circle</v-icon>
              Welcome back, <strong>{{ currentUser?.name }}</strong>!
              <span class="ml-2 text-caption">
                (Member since {{ new Date(currentUser?.createdAt).toLocaleDateString() }})
              </span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Loading State -->
      <v-row v-if="loading" justify="center">
        <v-col cols="auto">
          <v-progress-circular indeterminate size="64" color="primary" />
          <div class="text-center mt-4">Loading posts...</div>
        </v-col>
      </v-row>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" :text="error" class="mb-4" />

      <!-- Posts List -->
      <v-row v-else-if="posts.length" justify="center">
        <v-col cols="12" md="8">
          <v-card
            v-for="post in posts"
            :key="post.id"
            class="mb-4"
            elevation="2"
          >
            <v-card-item>
              <template v-slot:prepend>
                <v-avatar>
                  <v-img 
                    :src="getPostAuthorAvatar(post)"
                    :alt="post.publisherName"
                  />
                </v-avatar>
              </template>

              <v-card-title class="text-h6">{{ post.publisherName }}</v-card-title>
              <v-card-subtitle>{{ new Date(post.createdAt).toLocaleString() }}</v-card-subtitle>
            </v-card-item>

            <v-card-text class="text-body-1">
              {{ post.contentText }}
            </v-card-text>

            <!-- Show image if exists -->
            <v-card-text v-if="post.contentPhotoPath" class="pt-0">
              <v-img
                :src="post.contentPhotoPath"
                max-height="400"
                class="rounded"
              />
            </v-card-text>

            <v-card-actions>
              <v-btn variant="text" color="primary">
                <v-icon start>mdi-thumb-up</v-icon>
                {{ post.voteCount || 0 }}
              </v-btn>
              
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>

      <!-- Empty State -->
      <v-row v-else justify="center">
        <v-col cols="auto" class="text-center">
          <v-icon size="80" color="grey">mdi-post-outline</v-icon>
          <div class="text-h6 mt-4">No posts yet</div>
          <v-btn class="mt-4" color="primary" @click="goToCreatePost">
            <v-icon start>mdi-plus</v-icon>
            Create First Post
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>
