<script setup>
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "vue-toast-notification";
import { posts, loading, error, loadPosts, voteOnPost } from "../store/posts";
import { getCurrentUser } from "../store/auth.js";

const router = useRouter();
const currentUser = getCurrentUser();
const toast = useToast();

// User avatar helper (not used in template now, kept for future use)
const getUserAvatar = (user) => {
  if (user?.profilePhoto) return user.profilePhoto;
  if (user?.profilePhotoPath) return user.profilePhotoPath;
  if (user?.avatar) return user.avatar;
  if (user?.photo) return user.photo;
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(
    user?.username || "user"
  )}`;
};

// Use publisherPhoto from backend if present
const getPostAuthorAvatar = (post) => {
  if (post?.publisherPhoto) return post.publisherPhoto;
  if (post?.publisherPhotoPath) return post.publisherPhotoPath;
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(
    post?.publisherName || "anon"
  )}`;
};

// Get full image URL for post images
const getPostImageUrl = (post) => {
  if (post?.contentPhotoPath) {
    // Backend'den gelen path'i full URL'ye çevir
    console.log(`/api/images/${post.contentPhotoPath.split("/").pop()}`);
    return `/api/images/${post.contentPhotoPath.split("/").pop()}`;
  }
  return null;
};

onMounted(async () => {
  // Kullanıcı giriş yapmışsa userId ile posts al, yoksa null ile
  await loadPosts(currentUser?.id || null);
});

function goToCreatePost() {
  router.push("/create-post");
}

function goToEditPost(postId) {
  router.push(`/edit-post/${postId}`);
}

// Vote fonksiyonu
async function handleVote(postId, upvoted) {
  if (!currentUser) {
    toast.error("You must login to vote", {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }

  try {
    await voteOnPost(postId, currentUser.id, upvoted);
  } catch (error) {
    console.error("Vote error:", error);
  }
}

// Vote durum kontrolü için yardımcı fonksiyonlar
function getVoteButtonColor(post, isUpvote) {
  if (post.isLiked === null) return "grey";
  if (isUpvote && post.isLiked === true) return "success";
  if (!isUpvote && post.isLiked === false) return "error";
  return "grey";
}

function getVoteButtonVariant(post, isUpvote) {
  if (post.isLiked === null) return "text";
  if (isUpvote && post.isLiked === true) return "flat";
  if (!isUpvote && post.isLiked === false) return "flat";
  return "text";
}

// Post sahibi kontrolü
function isPostOwner(post) {
  return currentUser && post.publisherId === currentUser.id;
}
</script>

<template>
  <div>
    <v-container class="mt-4">
      <!-- User Welcome Card - Only show for authenticated users -->
      <v-row v-if="currentUser" justify="center" class="mb-4">
        <v-col cols="12" md="8">
          <v-card variant="tonal" color="primary" class="text-center">
            <v-card-text>
              <v-icon class="mr-2">mdi-account-circle</v-icon>
              Welcome back, <strong>{{ currentUser?.name }}</strong
              >!
              <span class="ml-2 text-caption">
                (Member since
                {{ new Date(currentUser?.createdAt).toLocaleDateString() }})
              </span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Loading State -->
      <v-row v-if="loading" justify="center">
        <v-col cols="auto">
          <v-progress-circular indeterminate size="64" color="primary" />
          <div class="text-center mt-4">Loading posts...</div>
        </v-col>
      </v-row>

      <!-- Posts List -->
      <v-row v-else-if="posts.length" justify="center">
        <v-col cols="12" md="8">
          <v-card
            v-for="post in posts"
            :key="post.id"
            class="mb-4 post-card"
            elevation="2"
          >
            <v-row no-gutters>
              <!-- Vote Section -->
              <v-col cols="1" class="d-flex flex-column align-center pa-2">
                <!-- Upvote Button -->
                <v-btn
                  :color="getVoteButtonColor(post, true)"
                  :variant="getVoteButtonVariant(post, true)"
                  icon="mdi-arrow-up-bold"
                  size="small"
                  @click="handleVote(post.id, true)"
                  class="mb-1 vote-button"
                  rounded="lg"
                />

                <!-- Vote Count -->
                <span class="text-body-2 font-weight-bold my-1 vote-count">
                  {{ post.voteCount || 0 }}
                </span>

                <!-- Downvote Button -->
                <v-btn
                  :color="getVoteButtonColor(post, false)"
                  :variant="getVoteButtonVariant(post, false)"
                  icon="mdi-arrow-down-bold"
                  size="small"
                  @click="handleVote(post.id, false)"
                  class="mt-1 vote-button"
                  rounded="lg"
                />
              </v-col>

              <!-- Post Content -->
              <v-col cols="11">
                <v-card-item class="position-relative">
                  <!-- Edit Button - Only show for post owner -->
                  <v-btn
                    v-if="isPostOwner(post)"
                    icon="mdi-pencil"
                    size="small"
                    color="primary"
                    variant="text"
                    @click="goToEditPost(post.id)"
                    class="edit-button"
                  >
                    <v-icon>mdi-pencil</v-icon>
                    <v-tooltip activator="parent" location="bottom">
                      Edit Post
                    </v-tooltip>
                  </v-btn>

                  <template v-slot:prepend>
                    <v-avatar>
                      <v-img
                        :src="getPostAuthorAvatar(post)"
                        :alt="post.publisherName"
                      />
                    </v-avatar>
                  </template>

                  <v-card-title class="text-h6">{{
                    post.publisherName
                  }}</v-card-title>
                  <v-card-subtitle>{{
                    new Date(post.createdAt).toLocaleString("en-US", {
                      timeZone: "UTC",
                    })
                  }}</v-card-subtitle>
                </v-card-item>

                <!-- Post Header/Title with Private Badge -->
                <div class="d-flex align-center mb-2">
                  <v-card-title v-if="post.header" class="text-h5 font-weight-bold flex-grow-0 pa-0 ps-4">
                    {{ post.header }}
                  </v-card-title>
                  <v-chip
                    v-if="post.isPrivate"
                    color="warning"
                    size="small"
                    class="ms-3"
                  >
                    <v-icon start size="small">mdi-lock</v-icon>
                    Private
                  </v-chip>
                </div>

                <v-card-text class="text-body-1">
                  {{ post.contentText }}
                </v-card-text>

                <!-- Show image if exists -->
                <v-card-text v-if="getPostImageUrl(post)" class="pt-0">
                  <v-img
                    :src="getPostImageUrl(post)"
                    max-height="400"
                    class="rounded"
                    @error="console.error('Image load error:', $event)"
                  />
                </v-card-text>
              </v-col>
            </v-row>
          </v-card>
        </v-col>
      </v-row>

      <!-- Empty State -->
      <v-row v-else justify="center">
        <v-col cols="auto" class="text-center">
          <v-icon size="80" color="grey">mdi-post-outline</v-icon>
          <div class="text-h6 mt-4">No posts yet</div>
          <v-btn v-if="currentUser" class="mt-4" color="primary" @click="goToCreatePost">
            <v-icon start>mdi-plus</v-icon>
            Create First Post
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<style scoped>
.v-card {
  transition: all 0.3s ease;
}

.v-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15) !important;
}

.vote-button {
  width: 40px !important;
  height: 40px !important;
  min-width: 40px !important;
  border-radius: 8px !important;
  transition: all 0.2s ease;
}

.vote-button:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.vote-count {
  font-size: 14px;
  font-weight: 600;
  color: #666;
  min-width: 20px;
  text-align: center;
}

.edit-button {
  position: absolute !important;
  top: 8px;
  right: 8px;
  z-index: 10;
  opacity: 0.7;
  transition: all 0.2s ease;
}

.edit-button:hover {
  opacity: 1;
  transform: scale(1.1);
}

.post-card {
  position: relative;
}
</style>
