import { ref } from 'vue'
import { getAllPosts, createPost, createPostWithImage, editPost, editPostWithImage, votePost } from '../api/posts.js'

const posts = ref([]);
const loading = ref(false);
const error = ref("");

// WebSocket message handlers
export function handleWebSocketMessage(message) {
  const { type, payload } = message

  switch (type) {
    case 'POST_CREATED':
      handlePostCreated(payload)
      break
    case 'POST_UPDATED':
      handlePostUpdated(payload)
      break
    case 'POST_VOTED':
      handlePostVoted(payload)
      break
    case 'POST_PRIVACY_CHANGED':
      handlePostPrivacyChanged(payload)
      break
    default:
      console.log('Unknown WebSocket message type:', type)
  }
}

function handlePostCreated(newPost) {
  // Check if post already exists (avoid duplicates)
  const existingIndex = posts.value.findIndex(p => p.id === newPost.id)
  if (existingIndex === -1) {
    // Only add the post if it's public or the current user is the owner
    const currentUser = JSON.parse(localStorage.getItem('devops_user'))
    const isOwner = currentUser && currentUser.id === newPost.publisherId

    if (!newPost.isPrivate || isOwner) {
      // Add new post to the beginning of the list
      posts.value.unshift(newPost)
      console.log('New post added via WebSocket:', newPost.id)
    } else {
      console.log('Ignored private post from WebSocket (not owner):', newPost.id)
    }
  }
}

function handlePostUpdated(updatedPost) {
  const postIndex = posts.value.findIndex(p => p.id === updatedPost.id)
  const currentUser = JSON.parse(localStorage.getItem('devops_user'))
  const isOwner = currentUser && currentUser.id === updatedPost.publisherId

  if (postIndex !== -1) {
    if (updatedPost.isPrivate && !isOwner) {
      // Post became private and the current user is not the owner - remove it
      posts.value.splice(postIndex, 1)
      console.log('Private post removed via WebSocket update:', updatedPost.id)
    } else {
      // Update existing post
      posts.value[postIndex] = updatedPost
      console.log('Post updated via WebSocket:', updatedPost.id)
    }
  } else {
    // Post doesn't exist locally - only add if public or owner
    if (!updatedPost.isPrivate || isOwner) {
      posts.value.unshift(updatedPost)
      console.log('Post added via WebSocket update:', updatedPost.id)
    } else {
      console.log('Ignored private post update (not owner):', updatedPost.id)
    }
  }
}

function handlePostVoted(updatedPost) {
  const postIndex = posts.value.findIndex(p => p.id === updatedPost.id)
  if (postIndex !== -1) {
    // Update vote count and isLiked status
    posts.value[postIndex] = updatedPost
    console.log('Post vote updated via WebSocket:', updatedPost.id, 'new count:', updatedPost.voteCount)
  }
}

function handlePostPrivacyChanged(updatedPost) {
  const postIndex = posts.value.findIndex(p => p.id === updatedPost.id)
  const currentUser = JSON.parse(localStorage.getItem('devops_user'))
  const isOwner = currentUser && currentUser.id === updatedPost.publisherId

  if (postIndex !== -1) {
    if (updatedPost.isPrivate && !isOwner) {
      // Post became private and user is not owner - remove it
      posts.value.splice(postIndex, 1)
      console.log('Private post removed via WebSocket:', updatedPost.id)
    } else if (!updatedPost.isPrivate || isOwner) {
      // Update the post if it's public or user is owner
      posts.value[postIndex] = updatedPost
      console.log('Post privacy changed via WebSocket:', updatedPost.id)
    }
  } else if (!updatedPost.isPrivate || isOwner) {
    // Add post if it's public or user is owner
    posts.value.unshift(updatedPost)
    console.log('Post added via WebSocket:', updatedPost.id)
  }
}

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

export async function addPost(publisherId, header, contentText, publisherName = "", imageFile = null, isPrivate = false) {
  try {
    let newPost
    if (imageFile) {
      // Create post with image
      newPost = await createPostWithImage(publisherId, header, contentText, publisherName, imageFile, isPrivate)
    } else {
      // Create post without image
      newPost = await createPost(publisherId, header, contentText, publisherName, isPrivate)
    }

    // Add new post from backend to the beginning of the list
    posts.value.unshift(newPost)
    return newPost
  } catch (err) {
    console.error('Error creating post:', err)
    throw err
  }
}

export async function updatePost(postId, userId, header, contentText, publisherName = "", imageFile = null, isPrivate = false) {
  try {
    let updatedPost
    if (imageFile) {
      // Update post with image
      updatedPost = await editPostWithImage(postId, userId, header, contentText, publisherName, imageFile, isPrivate)
    } else {
      // Update post without image
      updatedPost = await editPost(postId, userId, header, contentText, publisherName, isPrivate)
    }

    // Update local state
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

    // Update local state
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
