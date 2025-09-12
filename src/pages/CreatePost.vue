<script setup>
import { ref, computed } from 'vue'
import { addPost } from '../store/posts'
import { useRouter } from 'vue-router'
import { useNotifications } from '../composables/notifications'

const router = useRouter()
const { showSuccess, showError } = useNotifications()
const author = ref('')
const content = ref('')
const submitting = ref(false)

const canSubmit = computed(() => author.value.trim().length > 0 && content.value.trim().length > 0 && !submitting.value)

async function submit() {
  if (!canSubmit.value) return
  
  submitting.value = true
  try {
    await addPost(author.value, content.value)
    showSuccess('Post created successfully!')
    router.push({ name: 'home' })
  } catch (error) {
    console.error('Error creating post:', error)
    // Show error message from backend
    const errorMessage = error.response?.data?.message || error.message || 'An error occurred while creating the post'
    showError(errorMessage)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div>
    <div class="form">
      <label class="label" for="author">Author Name</label>
      <input id="author" class="input" v-model="author" placeholder="Your name" :disabled="submitting" />

      <label class="label" for="content">Content</label>
      <textarea id="content" class="textarea" v-model="content" rows="10" placeholder="What are you thinking?" :disabled="submitting"></textarea>

      <div class="actions">
        <button class="ghost" @click="$router.push('/')" :disabled="submitting">Cancel</button>
        <button class="submit" :disabled="!canSubmit" @click="submit">
          {{ submitting ? 'Creating...' : 'Create Post' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.label {
  font-size: 12px;
  color: #64748b;
}
.input, .textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}
.input:disabled, .textarea:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.actions {
  display: flex;
  gap: 8px;
}
.ghost {
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #0f172a;
  border-radius: 8px;
}
.ghost:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.submit {
  padding: 10px 14px;
  border: 1px solid #0ea5e9;
  background: #0ea5e9;
  color: white;
  border-radius: 8px;
  cursor: pointer;
}
.submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
