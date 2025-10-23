import axios from "axios"

const api = axios.create({
    baseURL: "/api",
    headers: { "Content-Type": "application/json" },
})

// Error handling yapılıyor ama toast gösterilmiyor
// Toast notifications component seviyesinde gösteriliyor
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error("API Error:", error.response?.data || error.message);
        return Promise.reject(error)
    }
)

export default api
