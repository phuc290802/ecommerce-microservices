<template>
  <div class="auth-page">
    <div class="auth-bg">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
    </div>

    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-logo">
          <span class="logo-icon">🛒</span>
          <span class="logo-text">ShopVerse</span>
        </div>

        <h2 class="auth-title">Tạo tài khoản</h2>
        <p class="auth-subtitle">Tham gia ShopVerse ngay hôm nay</p>

        <div v-if="auth.error" class="alert alert-error mb-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          {{ auth.error }}
        </div>
        <div v-if="success" class="alert alert-success mb-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
          Đăng ký thành công! Đang chuyển đến trang đăng nhập...
        </div>

        <form @submit.prevent="handleRegister" class="auth-form">
          <div class="form-group">
            <label class="form-label">Tên người dùng</label>
            <input v-model="username" type="text" class="form-input" placeholder="Nhập tên người dùng" required />
          </div>
          <div class="form-group">
            <label class="form-label">Email</label>
            <input v-model="email" type="email" class="form-input" placeholder="you@example.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">Số điện thoại <span style="color:var(--text-muted)">(tuỳ chọn)</span></label>
            <input v-model="phone" type="tel" class="form-input" placeholder="0xxxxxxxxx" />
          </div>
          <div class="form-group">
            <label class="form-label">Mật khẩu</label>
            <div class="password-wrapper">
              <input v-model="password" :type="showPwd ? 'text' : 'password'" class="form-input" placeholder="Ít nhất 6 ký tự" required minlength="6" />
              <button type="button" class="password-toggle" @click="showPwd = !showPwd">{{ showPwd ? '🙈' : '👁️' }}</button>
            </div>
          </div>

          <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading" style="justify-content: center; margin-top: 0.5rem;">
            <span v-if="auth.loading" class="spinner"></span>
            <span v-else>Đăng ký</span>
          </button>
        </form>

        <p class="auth-redirect">
          Đã có tài khoản?
          <router-link to="/login" class="link-accent">Đăng nhập</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const email = ref('')
const phone = ref('')
const password = ref('')
const showPwd = ref(false)
const success = ref(false)

auth.error = ''

const handleRegister = async () => {
  const ok = await auth.register(username.value, email.value, password.value)
  if (ok) {
    success.value = true
    setTimeout(() => router.push('/login'), 1500)
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.auth-bg { position: absolute; inset: 0; pointer-events: none; }
.bg-orb { position: absolute; border-radius: 50%; filter: blur(80px); opacity: 0.15; }
.orb-1 { width: 500px; height: 500px; background: #8b5cf6; top: -150px; left: -100px; animation: float1 9s ease-in-out infinite; }
.orb-2 { width: 400px; height: 400px; background: #6366f1; bottom: -100px; right: -80px; animation: float2 7s ease-in-out infinite; }
@keyframes float1 { 0%,100% { transform: translateY(0); } 50% { transform: translateY(20px); } }
@keyframes float2 { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-20px); } }

.auth-container {
  width: 100%;
  max-width: 440px;
  padding: 1.5rem;
  position: relative;
  z-index: 1;
}
.auth-card {
  background: rgba(15, 22, 41, 0.8);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 24px;
  padding: 2.5rem;
  backdrop-filter: blur(30px);
  box-shadow: 0 20px 80px rgba(0,0,0,0.5), 0 0 60px rgba(139,92,246,0.08);
}
.auth-logo { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1.5rem; }
.logo-icon { font-size: 1.5rem; }
.logo-text {
  font-size: 1.25rem;
  font-weight: 800;
  background: linear-gradient(135deg, #f0f4ff, #818cf8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.auth-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; }
.auth-subtitle { color: var(--text-secondary); font-size: 0.875rem; margin-bottom: 1.5rem; }
.auth-form { display: flex; flex-direction: column; gap: 1rem; }
.password-wrapper { position: relative; }
.password-wrapper .form-input { padding-right: 3rem; }
.password-toggle { position: absolute; right: 0.75rem; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; font-size: 1rem; padding: 0; }
.auth-redirect { margin-top: 1.5rem; text-align: center; font-size: 0.875rem; color: var(--text-secondary); }
.link-accent { color: var(--accent-light); text-decoration: none; font-weight: 600; }
.link-accent:hover { text-decoration: underline; }
</style>
