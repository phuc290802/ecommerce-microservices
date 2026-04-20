import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

import axios from 'axios'
import { useAuthStore } from './stores/auth'
import { useAdminStore } from './stores/admin'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Axios Interceptor for token expiration
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    const isAuthError = error.response?.status === 401 || 
                      error.response?.data?.includes('expired') ||
                      error.response?.data?.includes('invalid token')

    if (isAuthError) {
      const auth = useAuthStore()
      const admin = useAdminStore()
      
      // Determine if it was an admin or user route
      if (window.location.pathname.startsWith('/admin')) {
        admin.logout()
        router.push('/admin/login')
      } else {
        auth.logout()
        router.push('/login')
      }
    }
    return Promise.reject(error)
  }
)

app.mount('#app')
