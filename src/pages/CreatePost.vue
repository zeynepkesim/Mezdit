<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { addPost } from "../store/posts";
import { getCurrentUser } from "../store/auth.js";
import { validatePostContent, validatePostHeader, validateName } from "../utils/validation.js";

const router = useRouter();
const currentUser = getCurrentUser();

// Display name field is editable, defaults to current user's name
const author = ref(currentUser?.name || "");
const isAnonymous = ref(false);
const isPrivate = ref(false);
const header = ref("");
const content = ref("");
const loading = ref(false);
const error = ref("");
const selectedImage = ref(null);
const imagePreview = ref(null);

// Use profile photo from backend, fallback to Dicebear avatar
const getUserAvatar = (user) => {
  if (user?.profilePhoto) {
    return user.profilePhoto;
  }
  if (user?.profilePhotoPath) {
    return user.profilePhotoPath;
  }
  if (user?.avatar) {
    return user.avatar;
  }
  if (user?.photo) {
    return user.photo;
  }
  // Fallback: Dicebear avatar
  return `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(
    user?.username || "user"
  )}`;
};

function handleImageUpload(event) {
  const file = event.target.files[0];
  if (file) {
    // File size validation (10MB limit)
    if (file.size > 10 * 1024 * 1024) {
      error.value = "Image size must be less than 10MB";
      return;
    }

    // File type validation
    if (!file.type.startsWith("image/")) {
      error.value = "Please select a valid image file";
      return;
    }

    selectedImage.value = file;
    const reader = new FileReader();
    reader.onload = (e) => {
      imagePreview.value = e.target.result;
    };
    reader.readAsDataURL(file);
    error.value = ""; // Clear any previous errors
  }
}

function removeImage() {
  selectedImage.value = null;
  imagePreview.value = null;
}

async function handleSubmit() {
  if (!header.value.trim() || !content.value.trim()) {
    error.value = "Please fill in all fields";
    return;
  }

  // Display name is required when not posting anonymously
  if (!isAnonymous.value && !author.value.trim()) {
    error.value = "Display name is required when not posting anonymously";
    return;
  }

  // Validate header for XSS and length
  const headerValidation = validatePostHeader(header.value);
  if (!headerValidation.valid) {
    error.value = headerValidation.message;
    return;
  }

  // Validate content for XSS and length
  const contentValidation = validatePostContent(content.value);
  if (!contentValidation.valid) {
    error.value = contentValidation.message;
    return;
  }

  // Validate display name if not anonymous
  if (!isAnonymous.value) {
    const nameValidation = validateName(author.value);
    if (!nameValidation.valid) {
      error.value = nameValidation.message;
      return;
    }
  }

  loading.value = true;
  error.value = "";

  try {
    // Use authenticated user's ID
    const publisherId = currentUser?.id || 1;
    // Send empty string if anonymous, otherwise send author name
    const publisherName = isAnonymous.value ? "" : author.value.trim();
    await addPost(
      publisherId,
      header.value.trim(),
      content.value.trim(),
      publisherName,
      selectedImage.value,
      isPrivate.value
    );
    router.push("/posts");
  } catch (err) {
    error.value = err.message || "Failed to create post";
  } finally {
    loading.value = false;
  }
}

function goBack() {
  router.push("/posts");
}
</script>

<template>
  <div>
    <v-container class="mt-4">
      <v-row justify="center">
        <v-col cols="12" md="8" lg="6">
          <v-card elevation="3">
            <v-card-title class="text-h5 text-center py-6">
              <v-icon class="mr-2" color="primary">mdi-pencil</v-icon>
              Share Your Thoughts
            </v-card-title>

            <v-divider />

            <v-card-text class="pa-6">
              <!-- Form -->
              <v-form @submit.prevent="handleSubmit">
                <!-- Privacy and Anonymous Toggles -->
                <v-card variant="outlined" class="mb-4">
                  <v-card-text>
                    <!-- Anonymous Toggle -->
                    <div class="d-flex align-center mb-3">
                      <v-switch
                        v-model="isAnonymous"
                        color="primary"
                        hide-details
                        class="me-3"
                      />
                      <div>
                        <div class="text-body-1 font-weight-medium">
                          <v-icon start>mdi-incognito</v-icon>
                          Post as Anonymous
                        </div>
                        <div class="text-caption text-medium-emphasis">
                          Your identity will be hidden from other users
                        </div>
                      </div>
                    </div>

                    <v-divider class="my-3" />

                    <!-- Private Toggle -->
                    <div class="d-flex align-center">
                      <v-switch
                        v-model="isPrivate"
                        color="warning"
                        hide-details
                        class="me-3"
                      />
                      <div>
                        <div class="text-body-1 font-weight-medium">
                          <v-icon start>mdi-lock</v-icon>
                          Make Post Private
                        </div>
                        <div class="text-caption text-medium-emphasis">
                          Only you will be able to see this post
                        </div>
                      </div>
                    </div>
                  </v-card-text>
                </v-card>

                <v-text-field
                  v-model="author"
                  label="Display Name"
                  prepend-inner-icon="mdi-account"
                  variant="outlined"
                  :disabled="loading || isAnonymous"
                  :required="!isAnonymous"
                  :rules="!isAnonymous ? [(v) => !!v || 'Display name is required'] : []"
                  class="mb-4"
                  placeholder="Enter display name for this post"
                  :hint="isAnonymous ? 'Disabled when posting anonymously' : 'You can use any name you want for this post'"
                  persistent-hint
                />

                <v-text-field
                  v-model="header"
                  label="Post Title"
                  prepend-inner-icon="mdi-format-title"
                  variant="outlined"
                  :disabled="loading"
                  class="mb-4"
                  placeholder="Give your post a catchy title"
                  required
                  :rules="[(v) => !!v || 'Title is required']"
                />

                <v-textarea
                  v-model="content"
                  label="What's on your mind?"
                  prepend-inner-icon="mdi-message-text"
                  variant="outlined"
                  :disabled="loading"
                  rows="6"
                  placeholder="Share your thoughts, ideas, or experiences..."
                  auto-grow
                  counter="500"
                  :rules="[
                    (v) => !!v || 'Content is required',
                    (v) =>
                      v.length <= 500 ||
                      'Content must be less than 500 characters',
                  ]"
                  class="mb-4"
                />

                <!-- Image Upload Section -->
                <v-card variant="outlined" class="mb-4">
                  <v-card-title class="text-body-1 py-3">
                    <v-icon start>mdi-image</v-icon>
                    Add Photo (Optional)
                  </v-card-title>

                  <v-card-text class="pt-0">
                    <!-- Image Preview -->
                    <div v-if="imagePreview" class="mb-4">
                      <v-img
                        :src="imagePreview"
                        max-height="300"
                        class="rounded"
                      />
                      <div class="d-flex justify-end mt-2">
                        <v-btn
                          size="small"
                          color="error"
                          variant="text"
                          @click="removeImage"
                        >
                          <v-icon start>mdi-delete</v-icon>
                          Remove
                        </v-btn>
                      </div>
                    </div>

                    <!-- File Input -->
                    <div v-else>
                      <v-file-input
                        accept="image/*"
                        label="Choose image"
                        prepend-icon="mdi-camera"
                        variant="outlined"
                        :disabled="loading"
                        @change="handleImageUpload"
                        hide-details
                      />
                      <div class="text-caption text-medium-emphasis mt-2">
                        Supported formats: JPG, PNG, GIF (Max 10MB)
                      </div>
                    </div>
                  </v-card-text>
                </v-card>
              </v-form>
            </v-card-text>

            <v-card-actions class="px-6 pb-6">
              <v-btn variant="outlined" @click="goBack" :disabled="loading">
                Cancel
              </v-btn>

              <v-spacer />

              <v-btn
                color="primary"
                variant="elevated"
                :loading="loading"
                :disabled="!header.trim() || !content.trim() || content.length > 500 || (!isAnonymous && !author.trim())"
                @click="handleSubmit"
                size="large"
              >
                <v-icon start>mdi-send</v-icon>
                {{ selectedImage ? "Post with Image" : "Post" }}
              </v-btn>
            </v-card-actions>
          </v-card>

          <!-- Preview Card -->
          <v-card
            v-if="content.trim() && header.trim()"
            class="mt-6"
            variant="tonal"
            color="primary"
          >
            <v-card-title class="text-h6">
              <v-icon class="mr-2">mdi-eye</v-icon>
              Preview
            </v-card-title>

            <v-card-text>
              <v-card elevation="0" class="pa-4">
                <v-card-item>
                  <template v-slot:prepend>
                    <v-avatar>
                      <v-img
                        :src="getUserAvatar(currentUser)"
                        :alt="author.trim() || 'Anonymous'"
                      />
                    </v-avatar>
                  </template>

                  <v-card-title class="text-h6">
                    <span v-if="isAnonymous">
                      <v-icon start size="small">mdi-incognito</v-icon>
                      Anonymous
                    </span>
                    <span v-else>
                      {{ author.trim() || "Anonymous" }}
                    </span>
                  </v-card-title>
                  <v-card-subtitle>Just now</v-card-subtitle>
                </v-card-item>

                <v-card-title class="text-h5 font-weight-bold mt-2 mb-2">
                  {{ header }}
                </v-card-title>

                <v-card-text class="text-body-1 pb-2">
                  {{ content }}
                </v-card-text>

                <!-- Preview Image -->
                <v-card-text v-if="imagePreview" class="pt-0">
                  <v-img :src="imagePreview" max-height="250" class="rounded" />
                </v-card-text>
              </v-card>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>
