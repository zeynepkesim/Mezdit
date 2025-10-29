<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "vue-toast-notification";
import { getCurrentUser } from "../store/auth.js";
import { changePassword } from "../api/auth.js";
import { MAX_PASSWORD_LENGTH } from "../utils/validation.js";

const router = useRouter();
const currentUser = getCurrentUser();
const toast = useToast();

const currentPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const loading = ref(false);
const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

const passwordRules = [
  (v) => !!v || "Password is required",
  (v) => v.length >= 6 || "Password must be at least 6 characters",
  (v) => v.length <= MAX_PASSWORD_LENGTH || `Password must not exceed ${MAX_PASSWORD_LENGTH} characters`,
];

const confirmPasswordRules = [
  (v) => !!v || "Please confirm your password",
  (v) => v === newPassword.value || "Passwords do not match",
];

async function handleSubmit() {
  // Validation
  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    toast.error("Please fill in all fields", {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }

  if (newPassword.value !== confirmPassword.value) {
    toast.error("New passwords do not match", {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }

  if (newPassword.value.length < 6) {
    toast.error("New password must be at least 6 characters", {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }

  if (newPassword.value.length > MAX_PASSWORD_LENGTH) {
    toast.error(`New password must not exceed ${MAX_PASSWORD_LENGTH} characters`, {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }

  loading.value = true;

  try {
    const response = await changePassword(
      currentUser.id,
      currentPassword.value,
      newPassword.value
    );

    console.log("Change password response:", response);

    // Success
    toast.success("Password changed successfully!", {
      position: "bottom-right",
      duration: 3000,
    });

    // Clear form
    currentPassword.value = "";
    newPassword.value = "";
    confirmPassword.value = "";

    // Redirect to posts immediately
    setTimeout(() => {
      router.push("/posts");
    }, 1000);
  } catch (error) {
    console.error("Change password error:", error);
    console.error("Error response:", error.response);

    const errorMsg = error.response?.data?.error?.message || error.message || "Failed to change password. Please check your current password.";

    toast.error(errorMsg, {
      position: "bottom-right",
      duration: 5000,
    });
  } finally {
    loading.value = false;
  }
}

function handleCancel() {
  router.push("/posts");
}
</script>

<template>
  <v-container class="py-8">
    <v-row justify="center">
      <v-col cols="12" md="6" lg="5">
        <v-card elevation="3">
          <v-card-title class="text-h5 font-weight-bold text-center py-6 bg-primary">
            <v-icon class="mr-2" size="large">mdi-lock-reset</v-icon>
            Change Password
          </v-card-title>

          <v-card-subtitle class="text-center pt-4 pb-2">
            Update your account password
          </v-card-subtitle>

          <v-divider class="mt-2" />

          <v-card-text class="pa-6">
            <v-form @submit.prevent="handleSubmit">
              <!-- Current Password -->
              <v-text-field
                v-model="currentPassword"
                label="Current Password"
                :type="showCurrentPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock"
                :append-inner-icon="showCurrentPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showCurrentPassword = !showCurrentPassword"
                variant="outlined"
                :disabled="loading"
                :rules="[(v) => !!v || 'Current password is required']"
                class="mb-4"
                placeholder="Enter your current password"
              />

              <!-- New Password -->
              <v-text-field
                v-model="newPassword"
                label="New Password"
                :type="showNewPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-plus"
                :append-inner-icon="showNewPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showNewPassword = !showNewPassword"
                variant="outlined"
                :disabled="loading"
                :rules="passwordRules"
                class="mb-4"
                placeholder="Enter your new password"
                :hint="`Password must be 6-${MAX_PASSWORD_LENGTH} characters`"
                persistent-hint
              />

              <!-- Confirm New Password -->
              <v-text-field
                v-model="confirmPassword"
                label="Confirm New Password"
                :type="showConfirmPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-check"
                :append-inner-icon="showConfirmPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showConfirmPassword = !showConfirmPassword"
                variant="outlined"
                :disabled="loading"
                :rules="confirmPasswordRules"
                class="mb-4"
                placeholder="Confirm your new password"
              />

              <!-- Password Strength Indicator -->
              <v-card v-if="newPassword" variant="tonal" class="mb-4">
                <v-card-text class="py-3">
                  <div class="text-caption text-medium-emphasis mb-2">
                    Password Strength
                  </div>
                  <v-progress-linear
                    :model-value="
                      newPassword.length < 6 ? 33 : newPassword.length < 10 ? 66 : 100
                    "
                    :color="
                      newPassword.length < 6
                        ? 'error'
                        : newPassword.length < 10
                        ? 'warning'
                        : 'success'
                    "
                    height="8"
                    rounded
                  />
                  <div class="text-caption mt-2">
                    {{
                      newPassword.length < 6
                        ? "Weak"
                        : newPassword.length < 10
                        ? "Medium"
                        : "Strong"
                    }}
                  </div>
                </v-card-text>
              </v-card>
            </v-form>
          </v-card-text>

          <v-card-actions class="px-6 pb-6">
            <v-btn
              variant="outlined"
              color="grey"
              @click="handleCancel"
              :disabled="loading"
              prepend-icon="mdi-cancel"
            >
              Cancel
            </v-btn>

            <v-spacer />

            <v-btn
              color="primary"
              variant="elevated"
              :loading="loading"
              :disabled="
                !currentPassword ||
                !newPassword ||
                !confirmPassword ||
                newPassword !== confirmPassword ||
                newPassword.length < 6 ||
                newPassword.length > MAX_PASSWORD_LENGTH
              "
              @click="handleSubmit"
              prepend-icon="mdi-check"
              size="large"
            >
              Change Password
            </v-btn>
          </v-card-actions>
        </v-card>

        <!-- Security Tips -->
        <v-card class="mt-4" variant="tonal" color="info">
          <v-card-text>
            <div class="text-subtitle-2 font-weight-bold mb-2">
              <v-icon start>mdi-information</v-icon>
              Security Tips
            </div>
            <ul class="text-caption">
              <li>Use 6-{{ MAX_PASSWORD_LENGTH }} characters</li>
              <li>Include uppercase and lowercase letters</li>
              <li>Add numbers and special characters</li>
              <li>Avoid common words or personal information</li>
            </ul>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.bg-primary {
  background-color: rgb(var(--v-theme-primary));
  color: white;
}
</style>
