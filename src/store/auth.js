import { ref } from 'vue'
import { loginUser, registerUser } from '../api/auth.js'

const currentUser = ref(null)
const authToken = ref(null)
const isLoggedIn = ref(false)

export async function login(username, password) {
  try {
    const response = await loginUser(username, password)

    // Backend returns: { user: {...}, token: "..." }
    currentUser.value = response.user
    authToken.value = response.token
    isLoggedIn.value = true

    // Save to localStorage
    localStorage.setItem('devops_user', JSON.stringify(response.user))
    localStorage.setItem('devops_token', response.token)
    localStorage.setItem('devops_logged_in', 'true')

    return response
  } catch (error) {
    console.error('Login failed:', error)
    throw error
  }
}

export function logout() {
  currentUser.value = null
  authToken.value = null
  isLoggedIn.value = false

  // Clear from localStorage
  localStorage.removeItem('devops_user')
  localStorage.removeItem('devops_token')
  localStorage.removeItem('devops_logged_in')
}

export function initializeAuth() {
  // Check localStorage on page load
  const storedUser = localStorage.getItem('devops_user')
  const storedToken = localStorage.getItem('devops_token')
  const storedLoggedIn = localStorage.getItem('devops_logged_in')

  if (storedUser && storedToken && storedLoggedIn === 'true') {
    try {
      currentUser.value = JSON.parse(storedUser)
      authToken.value = storedToken
      isLoggedIn.value = true
    } catch (error) {
      console.error('Error parsing stored user data:', error)
      logout() // Clear if data is corrupted
    }
  }
}

export function isAuthenticated() {
  return isLoggedIn.value
}

export function getCurrentUser() {
  return currentUser.value
}

export function getAuthToken() {
  return authToken.value
}

export async function register(name, username, password) {
  try {
    const response = await registerUser(name, username, password)

    // Backend returns: { user: {...}, token: "..." }
    currentUser.value = response.user
    authToken.value = response.token
    isLoggedIn.value = true

    // Save to localStorage
    localStorage.setItem('devops_user', JSON.stringify(response.user))
    localStorage.setItem('devops_token', response.token)
    localStorage.setItem('devops_logged_in', 'true')

    return response
  } catch (error) {
    console.error('Register failed:', error)
    throw error
  }
}
