import { ref } from 'vue'
import { getAllPosts, createPost, createPostWithImage, editPost, editPostWithImage, votePost } from '../api/posts.js'

const posts = ref([]);
const loading = ref(false);
const error = ref("");

export async function loadPosts(userId = null) {
  loading.value = true
  error.value = ''
  try {
    const data = await getAllPosts(userId)
    posts.value = data
  } catch (err) {
    error.value = err.message || "Failed to load posts";
    console.error("Error loading posts:", err);
  } finally {
    loading.value = false;
  }
}

export async function addPost(publisherId, contentText, publisherName = "", imageFile = null) {
  try {
    let newPost
    if (imageFile) {
      // Image ile post oluştur
      newPost = await createPostWithImage(publisherId, contentText, publisherName, imageFile)
    } else {
      // Normal post oluştur
      newPost = await createPost(publisherId, contentText, publisherName)
    }
    
    // Backend'den dönen yeni post'u listenin başına ekle
    posts.value.unshift(newPost)
    return newPost
  } catch (err) {
    console.error('Error creating post:', err)
    throw err
  }
}

export async function updatePost(postId, userId, contentText, publisherName = "", imageFile = null) {
  try {
    let updatedPost
    if (imageFile) {
      // Image ile post güncelle
      updatedPost = await editPostWithImage(postId, userId, contentText, publisherName, imageFile)
    } else {
      // Normal post güncelle
      updatedPost = await editPost(postId, userId, contentText, publisherName)
    }
    
    // Local state'i güncelle
    const postIndex = posts.value.findIndex(p => p.id === postId)
    if (postIndex !== -1) {
      posts.value[postIndex] = updatedPost
    }
    
    return updatedPost
  } catch (err) {
    console.error('Error updating post:', err)
    throw err
  }
}

export async function voteOnPost(postId, userId, upvoted) {
  try {
    const updatedPost = await votePost(postId, userId, upvoted)
    
    // Local state'i güncelle
    const postIndex = posts.value.findIndex(p => p.id === postId)
    if (postIndex !== -1) {
      posts.value[postIndex] = updatedPost
    }
    
    return updatedPost
  } catch (err) {
    console.error('Error voting on post:', err)
    throw err
  }
}

export { posts, loading, error }
