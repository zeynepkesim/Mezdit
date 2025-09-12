<script setup>
import { onMounted } from 'vue'
import { posts, upvote, downvote, loading, error, getCurrentUser, loadPosts } from '../store/posts'

const currentUser = getCurrentUser()

// Track which posts the user has voted on (simple implementation)
const userVotes = {}

// Load posts when page mounts
onMounted(async () => {
  await loadPosts()
})

function handleUpvote(postId) {
  const wasUpvoted = userVotes[postId] === 'up'
  
  if (wasUpvoted) {
    // Remove upvote
    delete userVotes[postId]
  } else {
    // Give upvote
    userVotes[postId] = 'up'
  }
  
  upvote(postId)
}

function handleDownvote(postId) {
  const wasDownvoted = userVotes[postId] === 'down'
  
  if (wasDownvoted) {
    // Remove downvote
    delete userVotes[postId]
  } else {
    // Give downvote
    userVotes[postId] = 'down'
  }
  
  downvote(postId)
}

function isUpvoted(postId) {
  return userVotes[postId] === 'up'
}

function isDownvoted(postId) {
  return userVotes[postId] === 'down'
}
</script>

<template>
  <div>
    <div v-if="loading" class="loading">Loading posts...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <ul v-else-if="posts.length" class="list">
      <li v-for="post in posts" :key="post.id" class="item">
        <div class="body">
          <div class="header">
            <img class="avatar" :src="`https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(post.author || 'anon')}`" alt="avatar" />
            <div class="meta">
              <span class="author">{{ post.author }}</span>
              <span class="date">• {{ new Date(post.createdAt).toLocaleString() }}</span>
            </div>
          </div>
          <p class="content">{{ post.content }}</p>
          
          <div class="vote-actions">
            <button 
              class="vote-btn upvote" 
              :class="{ active: isUpvoted(post.id) }"
              @click="handleUpvote(post.id)"
            >
              Upvote ({{ post.likes || 0 }})
            </button>
            <button 
              class="vote-btn downvote" 
              :class="{ active: isDownvoted(post.id) }"
              @click="handleDownvote(post.id)"
            >
              Downvote ({{ post.dislikes || 0 }})
            </button>
          </div>
        </div>
      </li>
    </ul>
    <p v-else class="empty">No posts yet.</p>
    
    <div class="user-info">
      <small>User: {{ currentUser }}</small>
    </div>
  </div>
</template>

<style scoped>
.loading, .error {
  text-align: center;
  padding: 20px;
  color: #64748b;
}
.error {
  color: #ef4444;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  list-style: none;
  padding: 0;
  margin: 0;
}
.item {
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 16px;
  background: white;
}
.body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.header {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
}
.meta {
  color: #64748b;
  font-size: 12px;
}
.author {
  font-weight: 600;
  color: #0f172a;
}
.content {
  white-space: pre-wrap;
  line-height: 1.5;
  margin: 0;
}
.vote-actions {
  display: flex;
  gap: 12px;
  margin-top: 4px;
}
.vote-btn {
  padding: 8px 16px;
  background: white;
  border: 1px solid;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.vote-btn.upvote {
  border-color: #22c55e;
  color: #22c55e;
}
.vote-btn.upvote:hover {
  background: #f0fdf4;
}
.vote-btn.upvote.active {
  background: #22c55e;
  color: white;
}
.vote-btn.downvote {
  border-color: #ef4444;
  color: #ef4444;
}
.vote-btn.downvote:hover {
  background: #fef2f2;
}
.vote-btn.downvote.active {
  background: #ef4444;
  color: white;
}
.empty {
  color: #64748b;
}
.user-info {
  margin-top: 20px;
  padding: 8px;
  background: #f1f5f9;
  border-radius: 8px;
  text-align: center;
  color: #64748b;
}
</style>
