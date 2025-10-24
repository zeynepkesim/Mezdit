import { ref } from 'vue'
import { loginUser, registerUser } from '../api/auth.js'

const currentUser = ref(null)
const isLoggedIn = ref(false)

export async function login(username, password) {
  try {
    const userData = await loginUser(username, password)

    // Save user data to state
    currentUser.value = userData
    isLoggedIn.value = true

    // Save to localStorage
    localStorage.setItem('devops_user', JSON.stringify(userData))
    localStorage.setItem('devops_logged_in', 'true')

    return userData
  } catch (error) {
    console.error('Login failed:', error)
    throw error
  }
}

export function logout() {
  currentUser.value = null
  isLoggedIn.value = false

  // Clear from localStorage
  localStorage.removeItem('devops_user')
  localStorage.removeItem('devops_logged_in')
}

export function initializeAuth() {
  // Check localStorage on page load
  const storedUser = localStorage.getItem('devops_user')
  const storedLoggedIn = localStorage.getItem('devops_logged_in')

  if (storedUser && storedLoggedIn === 'true') {
    try {
      currentUser.value = JSON.parse(storedUser)
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

export async function register(name, username, password) {
  try {
    const userData = await registerUser(name, username, password)
    currentUser.value = userData
    isLoggedIn.value = true
    localStorage.setItem('devops_user', JSON.stringify(userData))
    localStorage.setItem('devops_logged_in', 'true')
    return userData
  } catch (error) {
    console.error('Register failed:', error)
    throw error
  }
}
