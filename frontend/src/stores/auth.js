import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userEmail = ref(localStorage.getItem('email') || '')
  const userName = ref(localStorage.getItem('username') || '')
  const loading = ref(false)
  const error = ref('')

  const isLoggedIn = computed(() => !!token.value)

  const authHeaders = computed(() =>
    token.value ? { Authorization: `Bearer ${token.value}` } : {}
  )

  async function login(email, password) {
    loading.value = true
    error.value = ''
    try {
      const res = await axios.post('/api/auth/login', { email, password })
      token.value = res.data.access_token
      userEmail.value = email
      userName.value = res.data.username || ''
      localStorage.setItem('token', token.value)
      localStorage.setItem('email', email)
      localStorage.setItem('username', userName.value)
      return true
    } catch (err) {
      error.value = err.response?.data?.trim() || err.message
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(usernameVal, email, password) {
    loading.value = true
    error.value = ''
    try {
      await axios.post('/api/auth/register', {
        username: usernameVal,
        email,
        password
      })
      userName.value = usernameVal
      localStorage.setItem('username', usernameVal)
      return true
    } catch (err) {
      error.value = err.response?.data?.trim() || err.message
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await axios.post('/api/auth/logout', {}, { headers: authHeaders.value })
    } catch { /* ignore */ }
    token.value = ''
    userEmail.value = ''
    userName.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('email')
    localStorage.removeItem('username')
  }

  async function forgotPassword(email) {
    loading.value = true
    error.value = ''
    try {
      await axios.post('/api/auth/forgot-password', { email })
      return true
    } catch (err) {
      error.value = err.response?.data?.trim() || err.message
      return false
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(token_val, password) {
    loading.value = true
    error.value = ''
    try {
      await axios.post('/api/auth/reset-password', { token: token_val, password })
      return true
    } catch (err) {
      error.value = err.response?.data?.trim() || err.message
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    token, userEmail, userName, loading, error, isLoggedIn, authHeaders,
    login, register, logout, forgotPassword, resetPassword
  }
})
