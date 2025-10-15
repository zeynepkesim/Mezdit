import api from "./axios"

export async function getAllPosts(userId = null) {
  const url = userId ? `/posts?userId=${userId}` : "/posts"
  const response = await api.get(url)
  return response.data.data
}

export async function createPost(publisherId, header, contentText, publisherName = "") {
  const requestBody = {
    publisherId,
    publisherName,
    header,
    contentText,
    contentPhotoId: null,
  }

  const response = await api.post("/posts", requestBody)
  return response.data.data
}

export async function createPostWithImage(publisherId, header, contentText, publisherName = "", imageFile = null) {
  const formData = new FormData()
  formData.append("publisherId", publisherId)
  formData.append("publisherName", publisherName)
  formData.append("header", header)
  formData.append("contentText", contentText)

  if (imageFile) {
    formData.append("image", imageFile)
  }

  const response = await api.post("/posts/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  })
  return response.data.data
}

export async function editPost(postId, userId, header, contentText, publisherName = "") {
  const requestBody = {
    postId: parseInt(postId),
    userId: parseInt(userId),
    header: String(header || "").trim(),
    contentText: String(contentText || "").trim(),
    publisherName: String(publisherName || "").trim(),
  }

  const response = await api.post("/posts/edit", requestBody)
  return response.data.data
}

export async function editPostWithImage(postId, userId, header, contentText, publisherName = "", imageFile = null) {
  const formData = new FormData()
  formData.append("postId", postId)
  formData.append("userId", userId)
  formData.append("header", header)
  formData.append("contentText", contentText)
  formData.append("publisherName", publisherName)

  if (imageFile) {
    formData.append("image", imageFile)
  }

  const response = await api.post("/posts/edit-with-image", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  })
  return response.data.data
}

export async function votePost(postId, userId, upvoted) {
  const requestBody = { postId, userId, upvoted }
  const response = await api.post("/vote", requestBody)
  return response.data.data
}
