import axios from 'axios'

const API_BASE_URL = '/api'

export async function fetchPosts() {
  try {
    const response = await axios.get(`${API_BASE_URL}/posts`)
    
    if (response.data.success) {
      return response.data.data
    } else {
      console.error('API Error:', response.data.error)
      return []
    }
  } catch (error) {
    console.error('Error fetching posts:', error)
    return []
  }
}

export async function createPost(author, content) {
  try {
    const response = await axios.post(`${API_BASE_URL}/posts`, {
      author,
      content
    }, {
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    if (response.status === 201 && response.data.success) {
      return response.data.data
    } else {
      const errorMessage = response.data.error?.message || 'Post oluşturulamadı'
      throw new Error(errorMessage)
    }
  } catch (error) {
    if (error.response?.data?.error?.message) {
      throw new Error(error.response.data.error.message)
    }
    throw error
  }
}

export async function votePost(postId, isLike, username) {
  try {
    const response = await axios.post(`${API_BASE_URL}/vote`, {
      id: postId,
      like: isLike,
      username
    }, {
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    if (response.data.success) {
      return response.data.data
    } else {
      const errorMessage = response.data.error?.message || 'Oylama başarısız'
      throw new Error(errorMessage)
    }
  } catch (error) {
    if (error.response?.data?.error?.message) {
      throw new Error(error.response.data.error.message)
    }
    throw error
  }
}
