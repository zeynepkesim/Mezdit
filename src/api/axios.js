import axios from "axios"
import { toast } from "../toast.js"

const api = axios.create({
    baseURL: "/api",
    headers: { "Content-Type": "application/json" },
})

api.interceptors.response.use(
    (response) => {

        if (["post", "put", "delete"].includes(response.config.method)) {
            toast.success("Success!!")
        }
        return response
    },
    (error) => {
        const message =
            error.response?.data?.error?.message ||
            error.message ||
            "Something went wrong"
        toast.error(message)
        return Promise.reject(error)
    }
)

export default api
