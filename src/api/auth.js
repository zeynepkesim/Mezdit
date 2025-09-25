const API_BASE_URL = '/api' // Vite proxy kullanıyoruz

export async function loginUser(username, password) {
  try {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: username,
        password: password
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error?.message || 'Login failed')
    }

    return data.data
  } catch (error) {
    console.error('Login error:', error)
    throw error
  }
}

export async function getUserByUsername(username) {
  try {
    const response = await fetch(`${API_BASE_URL}/users/${username}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error?.message || 'Failed to get user')
    }

    return data.data
  } catch (error) {
    console.error('Get user error:', error)
    throw error
  }
} 

export async function registerUser(name, username, password) {
  try {
    const response = await fetch(`${API_BASE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, username, password })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error?.message || 'Registration failed')
    }

    return data.data
  } catch (error) {
    console.error('Register error:', error)
    throw error
  }
}
