import axios from "axios"

const api = axios.create({
    baseURL: "/api",
    headers: { "Content-Type": "application/json" },
})

// Request interceptor to add JWT token to headers
api.interceptors.request.use(
    (config) => {
        // Get token from localStorage
        const token = localStorage.getItem('devops_token')

        // Add Authorization header if token exists
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }

        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// Response interceptor for error handling
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error("API Error:", error.response?.data || error.message);

        // Handle 401 Unauthorized - token expired or invalid
        if (error.response?.status === 401) {
            const errorCode = error.response?.data?.error?.code

            // Only redirect to login if token is invalid/expired
            // Don't redirect for incorrect_password (change password scenario)
            if (errorCode === 'invalid_token' || errorCode === 'missing_token' || errorCode === 'authentication_failed') {
                // Clear auth data and redirect to login
                localStorage.removeItem('devops_user')
                localStorage.removeItem('devops_token')
                localStorage.removeItem('devops_logged_in')

                // Redirect to login page if not already there
                if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
                    window.location.href = '/login'
                }
            }
            // For incorrect_password, just reject the error without redirecting
        }

        return Promise.reject(error)
    }
)

export default api
