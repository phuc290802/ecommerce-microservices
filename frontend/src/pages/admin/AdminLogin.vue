<template>
  <div class="login-wrapper">
    <div class="login-card glass-card">
      <div class="text-center mb-8">
        <h1 class="brand-logo">🛡️ Admin Access</h1>
        <p class="text-secondary mt-2">Vui lòng đăng nhập để quản trị hệ thống</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label>Email Quản trị</label>
          <input 
            v-model="email" 
            type="email" 
            class="form-input" 
            placeholder="admin@shopverse.com"
            required
          />
        </div>

        <div class="form-group mt-4">
          <label>Mật khẩu</label>
          <input 
            v-model="password" 
            type="password" 
            class="form-input" 
            placeholder="••••••••"
            required
          />
        </div>

        <div v-if="adminStore.error" class="alert alert-error mt-4">
          {{ adminStore.error }}
        </div>

        <button type="submit" class="btn btn-primary w-full mt-8" :disabled="adminStore.loading">
          {{ adminStore.loading ? 'Đang xác thực...' : 'Đăng nhập vào Dashboard' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../../stores/admin'

const adminStore = useAdminStore()
const router = useRouter()
const email = ref('')
const password = ref('')

const handleLogin = async () => {
  const success = await adminStore.login(email.value, password.value)
  if (success) {
    router.push('/admin')
  }
}
</script>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at top right, #1e1b4b, #020617);
}

.login-card {
  width: 100%;
  max-width: 420px;
  padding: 2.5rem;
  border-radius: 24px;
}

.brand-logo {
  font-size: 2rem;
  font-weight: 900;
  background: linear-gradient(135deg, #fff, #818cf8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  margin-bottom: 0.5rem;
  color: #94a3b8;
}

.glass-card {
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}
</style>
