<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "vue-toast-notification";
import { register } from "../store/auth.js";

const router = useRouter();
const toast = useToast();

const name = ref("");
const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");
const showPassword = ref(false);

async function handleRegister() {
  if (!name.value || !username.value || !password.value) {
    toast.error("Please fill in all fields", {
      position: "bottom-right",
      duration: 4000,
    });
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    await register(name.value, username.value, password.value);
    toast.success("Account created successfully!", {
      position: "bottom-right",
      duration: 3000,
    });
    router.push("/posts");
  } catch (err) {
    const errorMsg = err.response?.data?.error?.message || err.message || "Registration failed. Please try again.";
    toast.error(errorMsg, {
      position: "bottom-right",
      duration: 5000,
    });
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <v-container fluid class="fill-height pa-0">
    <v-row no-gutters class="fill-height">
      <!-- Left Side - Branding (62.5%) -->
      <v-col cols="12" md="8" class="d-none d-md-flex">
        <v-sheet
          class="fill-height d-flex flex-column justify-center align-center text-center"
          color="primary"
          style="
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            width: 100%;
          "
        >
          <div class="text-white w-100">
            <v-icon size="120" class="mb-6">mdi-flash</v-icon>
            <h1 class="text-h2 font-weight-bold mb-4">Mezdit</h1>
            <p class="text-h5 opacity-80">Share your thoughts with others</p>
          </div>
        </v-sheet>
      </v-col>

      <!-- Right Side - Register Form (37.5%) -->
      <v-col cols="12" md="4">
        <v-sheet
          class="fill-height d-flex align-center justify-center"
          color="grey-lighten-4"
        >
          <v-card class="pa-8" width="100%" max-width="400" elevation="0">
            <div class="text-center mb-8">
              <h2 class="text-h4 mb-2">Create your account</h2>
              <p class="text-body-1 text-medium-emphasis">
                It only takes a minute
              </p>
            </div>

            <v-form @submit.prevent="handleRegister">
              <v-text-field
                v-model="name"
                label="Full Name"
                prepend-inner-icon="mdi-account"
                variant="outlined"
                :disabled="loading"
                class="mb-3"
              />

              <v-text-field
                v-model="username"
                label="Username"
                prepend-inner-icon="mdi-account-circle"
                variant="outlined"
                :disabled="loading"
                class="mb-3"
              />

              <v-text-field
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                label="Password"
                prepend-inner-icon="mdi-lock"
                :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append-inner="showPassword = !showPassword"
                variant="outlined"
                :disabled="loading"
                class="mb-4"
              />

              <v-btn
                type="submit"
                block
                size="large"
                color="primary"
                :loading="loading"
                :disabled="!name || !username || !password"
                class="mb-4"
              >
                Create Account
              </v-btn>

              <div class="text-center">
                <span class="text-medium-emphasis"
                  >Already have an account?</span
                >
                <v-btn variant="text" color="primary" @click="router.push('/login')"
                  >Sign In</v-btn
                >
              </div>
            </v-form>
          </v-card>
        </v-sheet>
      </v-col>
    </v-row>
  </v-container>
</template>
