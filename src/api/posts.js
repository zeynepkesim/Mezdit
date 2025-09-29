import axios from "axios";

const API_BASE_URL = "/api";

export async function getAllPosts(userId = null) {
  try {
    let url = `${API_BASE_URL}/posts`;
    if (userId) {
      url += `?userId=${userId}`;
    }

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || "Failed to get posts");
    }

    return data.data;
  } catch (error) {
    console.error("Get posts error:", error);
    throw error;
  }
}

export async function createPost(publisherId, contentText, publisherName = "") {
  try {
    const requestBody = {
      publisherId: publisherId,
      publisherName: publisherName,
      contentText: contentText,
      contentPhotoId: null,
    };

    // İsteği konsola yazdır

    const response = await fetch(`${API_BASE_URL}/posts`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();

    // Response'u da konsola yazdır

    if (!response.ok) {
      throw new Error(data.error?.message || "Failed to create post");
    }

    return data.data;
  } catch (error) {
    console.error("Create post error:", error);
    throw error;
  }
}

export async function createPostWithImage(
  publisherId,
  contentText,
  publisherName = "",
  imageFile = null
) {
  try {
    const formData = new FormData();
    formData.append("publisherId", publisherId);
    formData.append("publisherName", publisherName);
    formData.append("contentText", contentText);

    if (imageFile) {
      formData.append("image", imageFile);
    }

    for (let [key, value] of formData.entries()) {
      if (key === "image") {
        console.log(`${key}:`, value.name, `(${value.size} bytes)`);
      } else {
        console.log(`${key}:`, value);
      }
    }

    const response = await fetch(`${API_BASE_URL}/posts/upload`, {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(
        data.error?.message || "Failed to create post with image"
      );
    }

    return data.data;
  } catch (error) {
    console.error("Create post with image error:", error);
    throw error;
  }
}

export async function editPost(
  postId,
  userId,
  contentText,
  publisherName = ""
) {
  try {
    // String değerleri temizle ve validate et
    const cleanContentText = String(contentText || "").trim();
    const cleanPublisherName = String(publisherName || "").trim();

    const requestBody = {
      postId: parseInt(postId),
      userId: parseInt(userId),
      contentText: cleanContentText,
      publisherName: cleanPublisherName,
    };

    // JSON stringify'i test et
    const jsonString = JSON.stringify(requestBody);

    const response = await fetch(`${API_BASE_URL}/posts/edit`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: jsonString,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || "Failed to edit post");
    }

    return data.data;
  } catch (error) {
    console.error("Edit post error:", error);
    throw error;
  }
}

export async function editPostWithImage(
  postId,
  userId,
  contentText,
  publisherName = "",
  imageFile = null
) {
  try {
    const formData = new FormData();
    formData.append("postId", postId);
    formData.append("userId", userId);
    formData.append("contentText", contentText);
    formData.append("publisherName", publisherName);

    if (imageFile) {
      formData.append("image", imageFile);
    }

    for (let [key, value] of formData.entries()) {
      if (key === "image") {
        console.log(`${key}:`, value.name, `(${value.size} bytes)`);
      } else {
        console.log(`${key}:`, value);
      }
    }

    const response = await fetch(`${API_BASE_URL}/posts/edit-with-image`, {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || "Failed to edit post with image");
    }

    return data.data;
  } catch (error) {
    console.error("Edit post with image error:", error);
    throw error;
  }
}

export async function votePost(postId, userId, upvoted) {
  try {
    const requestBody = {
      postId: postId,
      userId: userId,
      upvoted: upvoted,
    };


    const response = await fetch(`${API_BASE_URL}/vote`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();


    if (!response.ok) {
      throw new Error(data.error?.message || "Failed to vote");
    }

    return data.data;
  } catch (error) {
    console.error("Vote error:", error);
    throw error;
  }
}
