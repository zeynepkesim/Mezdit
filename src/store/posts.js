import { ref } from 'vue'
import { getAllPosts, createPost } from '../api/posts.js'

const posts = ref([])
const loading = ref(false)
const error = ref('')

export async function loadPosts() {
  loading.value = true
  error.value = ''
  try {
    const data = await getAllPosts()
    posts.value = data
  } catch (err) {
    error.value = err.message || 'Failed to load posts'
    console.error('Error loading posts:', err)
  } finally {
    loading.value = false
  }
}

export async function addPost(publisherId, contentText, publisherName = "") {
  try {
    const newPost = await createPost(publisherId, contentText, publisherName)
    // Backend'den dönen yeni post'u listenin başına ekle
    posts.value.unshift(newPost)
    return newPost
  } catch (err) {
    console.error('Error creating post:', err)
    throw err
  }
}

export { posts, loading, error }
