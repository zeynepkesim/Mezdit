import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import { initializeAuth } from './store/auth.js'

// Vuetify
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import '@mdi/font/css/materialdesignicons.css'

import VueToast from "vue-toast-notification"
import "vue-toast-notification/dist/theme-sugar.css" // pick theme: 'default', 'sugar', 'bootstrap'

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#667eea',
          secondary: '#764ba2',
          accent: '#f093fb',
          error: '#FF5252',
          info: '#2196F3',
          success: '#4CAF50',
          warning: '#FFC107'
        }
      }
    }
  }
})

// Auth
initializeAuth()

// Create app
const app = createApp(App)

app.use(router)
app.use(vuetify)

app.use(VueToast, {
  duration: 3000,
})

app.mount('#app')
