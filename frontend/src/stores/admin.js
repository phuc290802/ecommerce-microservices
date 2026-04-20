import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useAdminStore = defineStore('admin', () => {
  const adminToken = ref(localStorage.getItem('admin_token') || '')
  const adminInfo = ref(JSON.parse(localStorage.getItem('admin_info') || 'null'))
  const loading = ref(false)
  const error = ref('')

  const isAdminAuthenticated = computed(() => !!adminToken.value)
  const adminHeaders = computed(() => ({ 
    Authorization: `Bearer ${adminToken.value}`,
    'X-Admin-ID': adminInfo.value?.id || '',
    'X-Admin-Role': adminInfo.value?.role || ''
  }))

  async function login(email, password) {
    loading.value = true
    error.value = ''
    try {
      const res = await axios.post('/api/admin/login', { email, password })
      adminToken.value = res.data.token
      adminInfo.value = res.data.user
      
      localStorage.setItem('admin_token', adminToken.value)
      localStorage.setItem('admin_info', JSON.stringify(adminInfo.value))
      return true
    } catch (err) {
      error.value = err.response?.data?.trim() || 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  function logout() {
    adminToken.value = ''
    adminInfo.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
  }

  return { adminToken, adminInfo, loading, error, isAdminAuthenticated, adminHeaders, login, logout }
})
