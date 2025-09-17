<script setup>
import { onMounted } from 'vue'
import { posts, upvote, downvote, loading, error, getCurrentUser, loadPosts } from '../store/posts'

const currentUser = getCurrentUser()

// Track which posts the user has voted on (simple UI state)
const userVotes = {}

onMounted(async () => {
  await loadPosts()
})

function handleUpvote(postId) {
  const wasUpvoted = userVotes[postId] === 'up'
  if (wasUpvoted) delete userVotes[postId]; else userVotes[postId] = 'up'
  upvote(postId)
}

function handleDownvote(postId) {
  const wasDownvoted = userVotes[postId] === 'down'
  if (wasDownvoted) delete userVotes[postId]; else userVotes[postId] = 'down'
  downvote(postId)
}

function isUpvoted(postId) { return userVotes[postId] === 'up' }
function isDownvoted(postId) { return userVotes[postId] === 'down' }
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
          
          <div class="vote-inline">
            <button
              class="arrow up"
              :class="{ active: isUpvoted(post.id) }"
              @click="handleUpvote(post.id)"
              aria-label="Upvote"
              title="Upvote"
            >▲</button>
            <span class="like-count" :class="{ positive: (post.likeCount || 0) > 0, negative: (post.likeCount || 0) < 0 }">
              {{ post.likeCount ?? 0 }}
            </span>
            <button
              class="arrow down"
              :class="{ active: isDownvoted(post.id) }"
              @click="handleDownvote(post.id)"
              aria-label="Downvote"
              title="Downvote"
            >▼</button>
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
.error { color: #ef4444; }

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
.body { display: flex; flex-direction: column; gap: 12px; }
.header { display: flex; align-items: center; gap: 10px; }
.avatar { width: 36px; height: 36px; border-radius: 50%; border: 1px solid #e2e8f0; }
.meta { color: #64748b; font-size: 12px; }
.author { font-weight: 600; color: #0f172a; }
.content { white-space: pre-wrap; line-height: 1.5; margin: 0; }

.vote-inline { display: inline-flex; align-items: center; gap: 10px; }
.arrow { 
  width: 36px; height: 32px; 
  display: grid; place-items: center; 
  background: white; border: 1px solid #e2e8f0; 
  border-radius: 8px; 
  font-size: 16px; line-height: 1; 
  cursor: pointer; transition: all 0.2s; 
}
.arrow:hover { background: #f1f5f9; }
.arrow.up.active { background: #22c55e; border-color: #22c55e; color: white; }
.arrow.down.active { background: #ef4444; border-color: #ef4444; color: white; }

.like-count { font-weight: 700; min-width: 28px; text-align: center; }
.like-count.positive { color: #16a34a; }
.like-count.negative { color: #dc2626; }

.empty { color: #64748b; }
.user-info { margin-top: 20px; padding: 8px; background: #f1f5f9; border-radius: 8px; text-align: center; color: #64748b; }
</style>
