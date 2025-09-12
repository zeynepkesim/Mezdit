import { ref } from 'vue'
import { fetchPosts, createPost as apiCreatePost, votePost } from '../api/posts'

export const posts = ref([])
export const loading = ref(false)
export const error = ref(null)

// Simple username (in real app would come from auth)
const currentUser = ref('user' + Math.floor(Math.random() * 1000))

export async function loadPosts() {
  loading.value = true
  error.value = null
  try {
    const data = await fetchPosts()
    posts.value = data || []
  } catch (err) {
    error.value = 'Error loading posts'
    console.error('Error loading posts:', err)
  } finally {
    loading.value = false
  }
}

export async function addPost(author, content) {
  try {
    const newPost = await apiCreatePost(author, content)
    // Add the returned post to the list
    posts.value.unshift(newPost)
    return newPost
  } catch (err) {
    throw err
  }
}

export async function upvote(postId) {
  try {
    const updatedPost = await votePost(postId, true, currentUser.value)
    // Update local state
    const index = posts.value.findIndex(p => p.id === postId)
    if (index !== -1) {
      posts.value[index] = updatedPost
    }
  } catch (err) {
    console.error('Upvote error:', err)
    // Could show error to user
  }
}

export async function downvote(postId) {
  try {
    const updatedPost = await votePost(postId, false, currentUser.value)
    // Update local state
    const index = posts.value.findIndex(p => p.id === postId)
    if (index !== -1) {
      posts.value[index] = updatedPost
    }
  } catch (err) {
    console.error('Downvote error:', err)
    // Could show error to user
  }
}

export function getCurrentUser() {
  return currentUser.value
}
