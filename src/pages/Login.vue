<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { login } from "../store/auth.js";

const router = useRouter();

const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");
const showPassword = ref(false);

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = "Please enter username and password";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    await login(username.value, password.value);
    router.push("/posts");
  } catch (err) {
    error.value = err.message || "Login failed. Please check your credentials.";
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
          style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
        >
          <div class="text-white">
            <v-icon size="120" class="mb-6">mdi-flash</v-icon>
            <h1 class="text-h2 font-weight-bold mb-4">DevOps Progress</h1>
            <p class="text-h5 opacity-80">
              Streamline your development workflow
            </p>
          </div>
        </v-sheet>
      </v-col>

      <!-- Right Side - Login Form (37.5%) -->
      <v-col cols="12" md="4">
        <v-sheet
          class="fill-height d-flex align-center justify-center"
          color="grey-lighten-4"
        >
          <v-card class="pa-8" width="100%" max-width="400" elevation="0">
            <!-- Header -->
            <div class="text-center mb-8">
              <h2 class="text-h4 mb-2">Welcome</h2>
              <p class="text-body-1 text-medium-emphasis">
                Sign in to your account
              </p>
            </div>

            <!-- Login Form -->
            <v-form @submit.prevent="handleLogin">
              <v-text-field
                v-model="username"
                label="Username"
                prepend-inner-icon="mdi-account"
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
                :disabled="!username || !password"
                class="mb-4"
              >
                Sign In
              </v-btn>
            </v-form>

            <div class="text-center mb-6">
              <span class="text-medium-emphasis">Don't have an account?</span>
              <v-btn
                variant="text"
                color="primary"
                @click="router.push('/register')"
                >Create Account</v-btn
              >
            </div>
          </v-card>
        </v-sheet>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.login-page {
  height: 100vh;
  width: 100vw;
}

code {
  font-family: "Monaco", "Menlo", monospace;
  font-size: 0.75rem;
  font-weight: 600;
}
</style>
