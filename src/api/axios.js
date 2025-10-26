import axios from "axios"

const api = axios.create({
    baseURL: "/api",
    headers: { "Content-Type": "application/json" },
})

// Add JWT token to all requests
api.interceptors.request.use(
    (config) => {
        const storedUser = localStorage.getItem('devops_user')
        if (storedUser) {
            try {
                const userData = JSON.parse(storedUser)
                if (userData.token) {
                    config.headers.Authorization = `Bearer ${userData.token}`
                }
            } catch (error) {
                console.error("Error parsing stored user data:", error)
            }
        }
        return config
    },
    (error) => Promise.reject(error)
)

// Error handling without toast notifications
// Toast notifications are handled at component level
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error("API Error:", error.response?.data || error.message);
        return Promise.reject(error)
    }
)

export default api
