<template>
  <div class="admin-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <span class="logo-icon">🛡️</span>
        <span class="logo-text">Admin Panel</span>
      </div>
      
      <nav class="sidebar-nav">
        <router-link to="/admin" class="nav-item" active-class="active" exact>
          <span class="icon">📊</span> Dashboard
        </router-link>
        <router-link to="/admin/users" class="nav-item" active-class="active">
          <span class="icon">👥</span> Nhân sự
        </router-link>
        <div class="nav-group">QUẢN LÝ CỬA HÀNG</div>
        <router-link to="/admin/products" class="nav-item" active-class="active">
          <span class="icon">📦</span> Sản phẩm
        </router-link>
        <router-link to="/admin/orders" class="nav-item" active-class="active">
          <span class="icon">🛍️</span> Đơn hàng
        </router-link>
        <div class="nav-group">HỆ THỐNG</div>
        <router-link to="/admin/logs" class="nav-item" active-class="active">
          <span class="icon">📝</span> Audit Logs
        </router-link>
        <router-link to="/admin/settings" class="nav-item" active-class="active">
          <span class="icon">⚙️</span> Cài đặt
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="admin-profile">
          <div class="avatar">{{ admin.adminInfo?.username?.charAt(0).toUpperCase() }}</div>
          <div class="info">
            <p class="name">{{ admin.adminInfo?.username }}</p>
            <p class="role">{{ admin.adminInfo?.role }}</p>
          </div>
        </div>
        <button @click="handleLogout" class="logout-btn">
          <span>🚪</span> Đăng xuất
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <router-view></router-view>
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores/admin'

const admin = useAdminStore()
const router = useRouter()

const handleLogout = () => {
  admin.logout()
  router.push('/admin/login')
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: #050811; /* Đậm hơn nền khách hàng */
  color: #e2e8f0;
}

.sidebar {
  width: 260px;
  background: #0a0f1d;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
}

.sidebar-header {
  padding: 2rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.25rem;
  font-weight: 800;
  color: white;
}

.sidebar-nav {
  flex: 1;
  padding: 0 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border-radius: 12px;
  color: #94a3b8;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.2s;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: white;
}

.nav-item.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(99, 102, 241, 0.05));
  border: 1px solid rgba(99, 102, 241, 0.3);
  color: #818cf8;
}

.nav-group {
  padding: 1.5rem 1rem 0.5rem;
  font-size: 0.7rem;
  font-weight: 700;
  color: #475569;
  letter-spacing: 0.1em;
}

.sidebar-footer {
  padding: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.admin-profile {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #6366f1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
}

.name { font-size: 0.9rem; font-weight: 600; }
.role { font-size: 0.75rem; color: #64748b; }

.logout-btn {
  padding: 0.75rem;
  border-radius: 8px;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.logout-btn:hover { background: #ef4444; color: white; }

.main-content {
  flex: 1;
  margin-left: 260px;
  min-height: 100vh;
}
</style>
