import axios from 'axios'

const API_BASE_URL = '/api'

export async function getAllPosts() {
  try {
    const response = await fetch(`${API_BASE_URL}/posts`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error?.message || 'Failed to get posts')
    }

    return data.data
  } catch (error) {
    console.error('Get posts error:', error)
    throw error
  }
}

export async function createPost(publisherId, contentText, publisherName = "") {
  try {
    const requestBody = {
      publisherId: publisherId,
      publisherName: publisherName,
      contentText: contentText,
      contentPhotoId: null
    }

    // İsteği konsola yazdır
    console.log('🚀 POST /posts isteği gönderiliyor:')
    console.log('📤 Request Body:', JSON.stringify(requestBody, null, 2))
    console.log('🌐 URL:', `${API_BASE_URL}/posts`)

    const response = await fetch(`${API_BASE_URL}/posts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody)
    })

    const data = await response.json()

    // Response'u da konsola yazdır
    console.log('📥 Response Status:', response.status)
    console.log('�� Response Data:', JSON.stringify(data, null, 2))

    if (!response.ok) {
      throw new Error(data.error?.message || 'Failed to create post')
    }

    return data.data
  } catch (error) {
    console.error('❌ Create post error:', error)
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
