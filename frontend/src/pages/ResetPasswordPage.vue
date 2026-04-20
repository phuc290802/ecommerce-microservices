<template>
  <div class="auth-page">
    <div class="auth-bg">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
    </div>
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-logo">
          <span class="logo-icon">🔒</span>
          <span class="logo-text">ShopVerse</span>
        </div>

        <h2 class="auth-title">Đặt lại mật khẩu</h2>
        <p class="auth-subtitle">Nhập mật khẩu mới của bạn</p>

        <div v-if="auth.error" class="alert alert-error mb-2">{{ auth.error }}</div>
        <div v-if="success" class="alert alert-success mb-2">✓ Đã đặt lại mật khẩu thành công!</div>

        <form v-if="!success" @submit.prevent="handleReset" class="auth-form">
          <div class="form-group">
            <label class="form-label">Token reset</label>
            <input v-model="token" type="text" class="form-input" placeholder="Token từ email" required />
          </div>
          <div class="form-group">
            <label class="form-label">Mật khẩu mới</label>
            <input v-model="password" type="password" class="form-input" placeholder="Ít nhất 6 ký tự" required minlength="6" />
          </div>
          <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading" style="justify-content: center; margin-top: 0.5rem;">
            <span v-if="auth.loading" class="spinner"></span>
            <span v-else>Đặt lại mật khẩu</span>
          </button>
        </form>

        <div v-if="success" style="text-align:center; margin-top: 1rem;">
          <router-link to="/login" class="btn btn-primary">Đăng nhập ngay</router-link>
        </div>

        <p class="auth-redirect mt-2">
          <router-link to="/login" class="link-accent">← Quay lại đăng nhập</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const token = ref(route.query.token || '')
const password = ref('')
const success = ref(false)
auth.error = ''

const handleReset = async () => {
  const ok = await auth.resetPassword(token.value, password.value)
  if (ok) success.value = true
}
</script>

<style scoped>
.auth-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; position: relative; overflow: hidden; }
.auth-bg { position: absolute; inset: 0; pointer-events: none; }
.bg-orb { position: absolute; border-radius: 50%; filter: blur(80px); opacity: 0.15; }
.orb-1 { width: 400px; height: 400px; background: #10b981; top: -100px; left: -80px; animation: float1 8s ease-in-out infinite; }
.orb-2 { width: 350px; height: 350px; background: #6366f1; bottom: -80px; right: -60px; animation: float2 10s ease-in-out infinite; }
@keyframes float1 { 0%,100% { transform: translateY(0); } 50% { transform: translateY(20px); } }
@keyframes float2 { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-20px); } }

.auth-container { width: 100%; max-width: 440px; padding: 1.5rem; position: relative; z-index: 1; }
.auth-card { background: rgba(15,22,41,0.8); border: 1px solid rgba(255,255,255,0.08); border-radius: 24px; padding: 2.5rem; backdrop-filter: blur(30px); box-shadow: 0 20px 80px rgba(0,0,0,0.5); }
.auth-logo { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1.5rem; }
.logo-icon { font-size: 1.5rem; }
.logo-text { font-size: 1.25rem; font-weight: 800; background: linear-gradient(135deg, #f0f4ff, #818cf8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.auth-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; }
.auth-subtitle { color: var(--text-secondary); font-size: 0.875rem; margin-bottom: 1.5rem; }
.auth-form { display: flex; flex-direction: column; gap: 1rem; }
.auth-redirect { text-align: center; font-size: 0.875rem; color: var(--text-secondary); }
.link-accent { color: var(--accent-light); text-decoration: none; font-weight: 600; }
.link-accent:hover { text-decoration: underline; }
</style>
