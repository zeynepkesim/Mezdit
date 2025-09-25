<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const content = ref('')
const loading = ref(false)
const postId = ref(route.params.id)

const characterCount = computed(() => content.value.length)
const isOverLimit = computed(() => characterCount.value > 500)

onMounted(() => {
  // TODO: Load post data for editing
  console.log('Editing post with ID:', postId.value)
})

function handleSave() {
  if (!content.value.trim() || isOverLimit.value) {
    return
  }
  
  loading.value = true
  
  // TODO: Implement save logic
  setTimeout(() => {
    loading.value = false
    console.log('Post updated:', content.value)
    router.push('/posts')
  }, 1000)
}

function handleCancel() {
  router.push('/posts')
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
              <v-textarea
                v-model="content"
                label="Post Content"
                placeholder="What are you thinking? Update your thoughts..."
                prepend-inner-icon="mdi-message-text-outline"
                variant="outlined"
                :disabled="loading"
                :rules="[
                  v => !!v || 'Content is required',
                  v => v.length <= 500 || 'Content must be less than 500 characters'
                ]"
                rows="8"
                counter="500"
                :error="isOverLimit"
                :error-messages="isOverLimit ? 'Content exceeds 500 character limit' : ''"
                auto-grow
                class="mb-4"
              />

              <div class="d-flex justify-end align-center mb-4">
                <v-chip 
                  :color="isOverLimit ? 'error' : characterCount > 400 ? 'warning' : 'success'"
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
                Save Changes
              </template>
            </v-btn>
          </v-card-actions>
        </v-card>

        <!-- Preview Card -->
        <v-card v-if="content.trim()" class="mt-6" variant="tonal" color="warning">
          <v-card-title class="text-h6 font-weight-bold">
            <v-icon class="mr-2">mdi-eye</v-icon>
            Updated Preview
          </v-card-title>
          
          <v-card-text>
            <v-card 
              variant="outlined" 
              class="pa-4"
            >
              <v-card-item>
                <template v-slot:prepend>
                  <v-avatar size="40">
                    <v-img 
                      src="https://api.dicebear.com/9.x/identicon/svg?seed=preview"
                      alt="Preview"
                    />
                  </v-avatar>
                </template>

                <v-card-title class="text-body-1 font-weight-bold">
                  Your Name
                </v-card-title>
                
                <v-card-subtitle>
                  Updated just now
                </v-card-subtitle>
              </v-card-item>

              <v-card-text class="text-body-1 pt-2">
                {{ content }}
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