<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { addPost } from '../store/posts'
import { getCurrentUser, logout } from '../store/auth.js'

const router = useRouter()
const currentUser = getCurrentUser()

// Name alanı değiştirilebilir olsun, default olarak current user'ın adı gelsin
const author = ref(currentUser?.name || '')
const content = ref('')
const loading = ref(false)
const error = ref('')
const selectedImage = ref(null)
const imagePreview = ref(null)

// Backend'den gelen profil fotoğrafını kullan, yoksa Dicebear avatar'ı göster
const getUserAvatar = (user) => {
  if (user?.profilePhoto) {
    return user.profilePhoto
  }
  if (user?.profilePhotoPath) {
    return user.profilePhotoPath
  }
  if (user?.avatar) {
    return user.avatar
  }
  if (user?.photo) {
    return user.photo
  }
  // Fallback: Dicebear avatar
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(user?.username || 'user')}`
}

function handleImageUpload(event) {
  const file = event.target.files[0]
  if (file) {
    selectedImage.value = file
    const reader = new FileReader()
    reader.onload = (e) => {
      imagePreview.value = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

function removeImage() {
  selectedImage.value = null
  imagePreview.value = null
}

async function handleSubmit() {
  if (!content.value.trim()) {
    error.value = 'Please fill in all fields'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    // Giriş yapan kullanıcının ID'sini kullan
    const publisherId = currentUser?.id || 1
    await addPost(publisherId, content.value.trim(), author.value.trim())
    router.push('/posts')
  } catch (err) {
    error.value = err.message || 'Failed to create post'
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/posts')
}

function handleLogout() {
  logout()
  router.push('/')
}

function viewProfile() {
  // TODO: Profile sayfasına yönlendir
  console.log('View profile clicked')
}
</script>

<template>
  <div>
    <v-container class="mt-4">
      <v-row justify="center">
        <v-col cols="12" md="8" lg="6">
          <v-card elevation="3">
            <v-card-title class="text-h5 text-center py-6">
              <v-icon class="mr-2" color="primary">mdi-pencil</v-icon>
              Share Your Thoughts
            </v-card-title>

            <v-divider />

            <v-card-text class="pa-6">
              <!-- Error Alert -->
              <v-alert v-if="error" type="error" class="mb-4" :text="error" />

              <!-- Form -->
              <v-form @submit.prevent="handleSubmit">
                <v-text-field
                  v-model="author"
                  label="Display Name"
                  prepend-inner-icon="mdi-account"
                  variant="outlined"
                  :disabled="loading"
                  class="mb-4"
                  placeholder="Enter display name for this post"
                  hint="You can use any name you want for this post"
                  persistent-hint
                />

                <v-textarea
                  v-model="content"
                  label="What's on your mind?"
                  prepend-inner-icon="mdi-message-text"
                  variant="outlined"
                  :disabled="loading"
                  rows="6"
                  placeholder="Share your thoughts, ideas, or experiences..."
                  auto-grow
                  counter="500"
                  :rules="[v => v.length <= 500 || 'Content must be less than 500 characters']"
                  class="mb-4"
                />

                <!-- Image Upload Section -->
                <v-card variant="outlined" class="mb-4">
                  <v-card-title class="text-body-1 py-3">
                    <v-icon start>mdi-image</v-icon>
                    Add Photo (Optional)
                  </v-card-title>
                  
                  <v-card-text class="pt-0">
                    <!-- Image Preview -->
                    <div v-if="imagePreview" class="mb-4">
                      <v-img
                        :src="imagePreview"
                        max-height="300"
                        class="rounded"
                      />
                      <div class="d-flex justify-end mt-2">
                        <v-btn size="small" color="error" variant="text" @click="removeImage">
                          <v-icon start>mdi-delete</v-icon>
                          Remove
                        </v-btn>
                      </div>
                    </div>

                    <!-- File Input -->
                    <div v-else>
                      <v-file-input
                        accept="image/*"
                        label="Choose image"
                        prepend-icon="mdi-camera"
                        variant="outlined"
                        :disabled="loading"
                        @change="handleImageUpload"
                        hide-details
                      />
                      <div class="text-caption text-medium-emphasis mt-2">
                        Supported formats: JPG, PNG, GIF (Max 10MB)
                        <br>
                        <strong>Note:</strong> Images are for preview only and won't be saved
                      </div>
                    </div>
                  </v-card-text>
                </v-card>
              </v-form>
            </v-card-text>

            <v-card-actions class="px-6 pb-6">
              <v-btn
                variant="outlined"
                @click="goBack"
                :disabled="loading"
              >
                Cancel
              </v-btn>

              <v-spacer />

              <v-btn
                color="primary"
                variant="elevated"
                :loading="loading"
                :disabled="!content.trim() || content.length > 500"
                @click="handleSubmit"
                size="large"
              >
                <v-icon start>mdi-send</v-icon>
                Post
              </v-btn>
            </v-card-actions>
          </v-card>

          <!-- Preview Card -->
          <v-card v-if="content.trim()" class="mt-6" variant="tonal" color="primary">
            <v-card-title class="text-h6">
              <v-icon class="mr-2">mdi-eye</v-icon>
              Preview
            </v-card-title>
            
            <v-card-text>
              <v-card elevation="0" class="pa-4">
                <v-card-item>
                  <template v-slot:prepend>
                    <v-avatar>
                      <v-img 
                        :src="getUserAvatar(currentUser)"
                        :alt="(author.trim() || 'Anonymous')"
                      />
                    </v-avatar>
                  </template>

                  <v-card-title class="text-h6">{{ author.trim() || 'Anonymous' }}</v-card-title>
                  <v-card-subtitle>Just now</v-card-subtitle>
                </v-card-item>

                <v-card-text class="text-body-1 pb-2">
                  {{ content }}
                </v-card-text>

                <!-- Preview Image -->
                <v-card-text v-if="imagePreview" class="pt-0">
                  <v-img
                    :src="imagePreview"
                    max-height="250"
                    class="rounded"
                  />
                  <div class="text-caption text-medium-emphasis mt-2 text-center">
                    * This image is for preview only
                  </div>
                </v-card-text>

                <v-card-actions>
                  <v-btn variant="text" color="primary">
                    <v-icon start>mdi-thumb-up</v-icon>
                    0
                  </v-btn>
                  
                </v-card-actions>
              </v-card>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>
