import api from "./axios" // ✅ use our custom instance


export async function loginUser(username, password) {
  const response = await api.post("/login", { username, password })
  return response.data.data
}


export async function getUserByUsername(username) {
  const response = await api.get(`/users/${username}`)
  return response.data.data
}

export async function registerUser(name, username, password) {
  const response = await api.post("/register", { name, username, password })
  return response.data.data
}
