import axios from 'axios'

const API_BASE_URL = '/api'

export async function createPost(author, content) {
  try {
    // URL-encoded format (orijinal curl komutu gibi)
    const data = new URLSearchParams()
    data.append('author', author)
    data.append('content', content)

    const response = await axios.post(`${API_BASE_URL}/posts`, data, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      }
    })

    if (response.status === 201 || response.status === 200) {
      return response.data
    } else {
      throw new Error(`Unexpected status code: ${response.status}`)
    }
  } catch (error) {
    console.error('Error creating post:', error)
    throw error
  }
}
