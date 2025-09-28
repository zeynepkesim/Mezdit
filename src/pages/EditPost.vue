<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { updatePost, posts } from '../store/posts'
import { getCurrentUser } from '../store/auth.js'

const route = useRoute()
const router = useRouter()
const currentUser = getCurrentUser()

const content = ref('')
const username = ref('')
const loading = ref(false)
const postId = ref(route.params.id)
const selectedImage = ref(null)
const imagePreview = ref(null)

const characterCount = computed(() => content.value.length);
const isOverLimit = computed(() => characterCount.value > 500);

// Mevcut post'u bul ve içeriğini yükle
const currentPost = computed(() => {
  return posts.value.find((p) => p.id === parseInt(postId.value));
});


// Get full image URL for post images
const getPostImageUrl = (post) => {
  if (post?.contentPhotoPath) {
    return `/api/images/${post.contentPhotoPath.split('/').pop()}`
  }
  return null
}

onMounted(() => {
  if (currentPost.value) {
    content.value = currentPost.value.contentText
    username.value = currentPost.value.publisherName || ''
  } else {
    // Post bulunamadı, ana sayfaya yönlendir
    router.push('/posts')
  }
})

function handleImageUpload(event) {
  const file = event.target.files[0]
  if (file) {
    // File size validation (10MB limit)
    if (file.size > 10 * 1024 * 1024) {
      alert('Image size must be less than 10MB')
      return
    }
    
    // File type validation
    if (!file.type.startsWith('image/')) {
      alert('Please select a valid image file')
      return
    }
    
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

async function handleSave() {
  if (!content.value.trim() || isOverLimit.value) {
    return
  }
  
  loading.value = true
  
  try {
    // Image varsa editPostWithImage, yoksa normal editPost kullan
    await updatePost(
      parseInt(postId.value), 
      currentUser.id, 
      content.value.trim(), 
      username.value.trim(),
      selectedImage.value
    )
    router.push('/posts')
  } catch (error) {
    console.error('Error updating post:', error)
    alert('Post güncellenirken bir hata oluştu: ' + error.message)
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  router.push("/posts");
}
</script>

<template>
  <v-container class="py-8">
    <v-row justify="center">
      <v-col cols="12" md="8" lg="6">
        <v-card elevation="3">
          <v-card-title class="text-h5 font-weight-bold text-center py-6">
            <v-icon class="mr-2" color="warning">mdi-pencil-outline</v-icon>
            Edit Post
          </v-card-title>

          <v-card-subtitle class="text-center pb-4">
            Update your existing post
          </v-card-subtitle>

          <v-divider />

          <v-card-text class="pa-6">
            <v-form @submit.prevent="handleSave">
              <!-- Username Field -->
              <v-text-field
                v-model="username"
                label="Display Name"
                placeholder="Enter your display name (optional)"
                prepend-inner-icon="mdi-account"
                variant="outlined"
                :disabled="loading"
                class="mb-4"
                hint="Leave empty to use your account name"
                persistent-hint
              />

              <!-- Content Field -->
              <!-- Username Field -->
              <v-text-field
                v-model="username"
                label="Display Name"
                placeholder="Enter your display name (optional)"
                prepend-inner-icon="mdi-account"
                variant="outlined"
                :disabled="loading"
                class="mb-4"
                hint="Leave empty to use your account name"
                persistent-hint
              />

              <!-- Content Field -->
              <v-textarea
                v-model="content"
                label="Post Content"
                placeholder="What are you thinking? Update your thoughts..."
                prepend-inner-icon="mdi-message-text-outline"
                variant="outlined"
                :disabled="loading"
                :rules="[
                  (v) => !!v || 'Content is required',
                  (v) =>
                    v.length <= 500 ||
                    'Content must be less than 500 characters',
                ]"
                rows="8"
                counter="500"
                :error="isOverLimit"
                :error-messages="
                  isOverLimit ? 'Content exceeds 500 character limit' : ''
                "
                auto-grow
                class="mb-4"
              />

              <!-- Current Image Display -->
              <v-card v-if="getPostImageUrl(currentPost)" variant="outlined" class="mb-4">
                <v-card-title class="text-body-1 py-3">
                  <v-icon start>mdi-image</v-icon>
                  Current Image
                </v-card-title>
                <v-card-text class="pt-0">
                  <v-img
                    :src="getPostImageUrl(currentPost)"
                    max-height="300"
                    class="rounded"
                  />
                  <div class="text-caption text-medium-emphasis mt-2">
                    Current post image
                  </div>
                </v-card-text>
              </v-card>

              <!-- Image Upload Section -->
              <v-card variant="outlined" class="mb-4">
                <v-card-title class="text-body-1 py-3">
                  <v-icon start>mdi-image-plus</v-icon>
                  Update Image (Optional)
                </v-card-title>
                
                <v-card-text class="pt-0">
                  <!-- New Image Preview -->
                  <div v-if="imagePreview" class="mb-4">
                    <v-img
                      :src="imagePreview"
                      max-height="300"
                      class="rounded"
                    />
                    <div class="d-flex justify-end mt-2">
                      <v-btn size="small" color="error" variant="text" @click="removeImage">
                        <v-icon start>mdi-delete</v-icon>
                        Remove New Image
                      </v-btn>
                    </div>
                  </div>

                  <!-- File Input -->
                  <div v-else>
                    <v-file-input
                      accept="image/*"
                      label="Choose new image"
                      prepend-icon="mdi-camera"
                      variant="outlined"
                      :disabled="loading"
                      @change="handleImageUpload"
                      hide-details
                    />
                    <div class="text-caption text-medium-emphasis mt-2">
                      Supported formats: JPG, PNG, GIF (Max 10MB)
                    </div>
                  </div>
                </v-card-text>
              </v-card>

              <div class="d-flex justify-end align-center mb-4">
                <v-chip
                  :color="
                    isOverLimit
                      ? 'error'
                      : characterCount > 400
                      ? 'warning'
                      : 'success'
                  "
                  variant="tonal"
                  size="small"
                >
                  {{ characterCount }}/500 characters
                </v-chip>
              </div>
            </v-form>
          </v-card-text>

          <v-card-actions class="px-6 pb-6">
            <v-btn
              variant="outlined"
              color="grey"
              @click="handleCancel"
              :disabled="loading"
              prepend-icon="mdi-cancel"
            >
              Cancel
            </v-btn>

            <v-spacer />

            <v-btn
              color="warning"
              variant="elevated"
              :loading="loading"
              :disabled="!content.trim() || isOverLimit"
              @click="handleSave"
              prepend-icon="mdi-content-save"
              size="large"
            >
              <template v-if="loading">
                Saving...
              </template>
              <template v-else>
                {{ selectedImage ? 'Save with New Image' : 'Save Changes' }}
              </template>
            </v-btn>
          </v-card-actions>
        </v-card>

        <!-- Preview Card -->
        <v-card
          v-if="content.trim()"
          class="mt-6"
          variant="tonal"
          color="warning"
        >
          <v-card-title class="text-h6 font-weight-bold">
            <v-icon class="mr-2">mdi-eye</v-icon>
            Updated Preview
          </v-card-title>

          <v-card-text>
            <v-card variant="outlined" class="pa-4">
              <v-card-item>
                <template v-slot:prepend>
                  <v-avatar size="40">
                    <v-img 
                      :src="currentPost?.publisherPhoto || `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(username || 'anon')}`"
                      :alt="username || currentPost?.publisherName"
                    />
                  </v-avatar>
                </template>

                <v-card-title class="text-body-1 font-weight-bold">
                  {{ username || currentPost?.publisherName || 'Your Name' }}
                </v-card-title>

                <v-card-subtitle>
                  Updated just now
                </v-card-subtitle>
              </v-card-item>

              <v-card-text class="text-body-1 pt-2">
                {{ content }}
              </v-card-text>

              <!-- Preview Image - Show new image if selected, otherwise current image -->
              <v-card-text v-if="imagePreview || getPostImageUrl(currentPost)" class="pt-0">
                <v-img
                  :src="imagePreview || getPostImageUrl(currentPost)"
                  max-height="250"
                  class="rounded"
                />
                <div v-if="imagePreview" class="text-caption text-medium-emphasis mt-2 text-center">
                  * New image preview
                </div>
                <div v-else class="text-caption text-medium-emphasis mt-2 text-center">
                  * Current image
                </div>
              </v-card-text>
            </v-card>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
/* No custom CSS needed with Vuetify! */
</style>
