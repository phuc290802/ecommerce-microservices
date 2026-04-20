<template>
  <div class="admin-page">
    <header class="page-header">
      <h1 class="text-2xl font-bold">📊 Dashboard Tổng Quan</h1>
      <p class="text-secondary">Chào mừng quay trở lại, {{ admin.adminInfo?.username }}</p>
    </header>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div v-for="stat in statCards" :key="stat.title" class="stat-card glass-card">
        <div class="stat-icon" :style="{ backgroundColor: stat.color + '20', color: stat.color }">
          {{ stat.icon }}
        </div>
        <div class="stat-info">
          <p class="stat-label">{{ stat.title }}</p>
          <h2 class="stat-value">{{ stat.value }}</h2>
          <span class="stat-trend" :class="stat.trend.startsWith('+') ? 'up' : 'down'">
            {{ stat.trend }} so với tháng trước
          </span>
        </div>
      </div>
    </div>

    <!-- Recent Activity & Charts Placeholder -->
    <div class="dashboard-content">
      <div class="activity-section glass-card">
        <div class="section-header">
          <h3 class="font-bold">Hành động gần đây</h3>
          <button class="btn btn-sm btn-ghost">Xem tất cả</button>
        </div>
        <div class="activity-list">
          <div v-for="i in 5" :key="i" class="activity-item">
            <div class="activity-dot"></div>
            <div class="activity-details">
              <p><strong>Admin</strong> đã cập nhật trạng thái đơn hàng #ORD-{{ 1000 + i }}</p>
              <span class="text-xs text-secondary">{{ i }} giờ trước • IP: 192.168.1.{{ i }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="quick-links glass-card">
        <h3 class="font-bold mb-4">Lối tắt quản lý</h3>
        <div class="links-grid">
          <button @click="$router.push('/admin/users')" class="link-btn">
            <span>👥</span> Quản lý nhân sự
          </button>
          <button class="link-btn">
            <span>📦</span> Quản lý kho
          </button>
          <button class="link-btn">
            <span>💰</span> Báo cáo doanh thu
          </button>
          <button class="link-btn">
            <span>🛠️</span> Cài đặt hệ thống
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAdminStore } from '../../stores/admin'
import axios from 'axios'

const admin = useAdminStore()
const stats = ref({})

const statCards = ref([
  { title: 'Doanh thu ngày', value: '$0', icon: '💰', color: '#10b981', trend: '+12%' },
  { title: 'Đơn hàng mới', value: '0', icon: '🛍️', color: '#6366f1', trend: '+5%' },
  { title: 'Khách hàng mới', value: '0', icon: '👤', color: '#f59e0b', trend: '+18%' },
  { title: 'Lượt truy cập', value: '1.2k', icon: '📈', color: '#ec4899', trend: '+3%' },
])

const loadStats = async () => {
  try {
    const res = await axios.get('/api/admin/dashboard/stats', { headers: admin.adminHeaders })
    stats.value = res.data
    statCards.value[0].value = '$' + res.data.daily_revenue.toLocaleString()
    statCards.value[1].value = res.data.pending_orders
    statCards.value[2].value = res.data.new_users_today
  } catch (err) {
    console.error('Failed to load stats', err)
  }
}

onMounted(loadStats)
</script>

<style scoped>
.admin-page {
  padding: 2rem;
}

.page-header {
  margin-bottom: 2.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2.5rem;
}

.stat-card {
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  border-radius: 20px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0.25rem;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 800;
  margin-bottom: 0.25rem;
}

.stat-trend {
  font-size: 0.75rem;
  font-weight: 500;
}
.stat-trend.up { color: #10b981; }
.stat-trend.down { color: #ef4444; }

.dashboard-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 1.5rem;
}

.glass-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 20px;
  padding: 1.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.activity-item {
  display: flex;
  gap: 1rem;
  position: relative;
}

.activity-dot {
  width: 10px;
  height: 10px;
  background: var(--accent);
  border-radius: 50%;
  margin-top: 5px;
  flex-shrink: 0;
  box-shadow: 0 0 10px var(--accent);
}

.links-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.link-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1rem;
  border-radius: 12px;
  text-align: left;
  color: var(--text-primary);
  font-weight: 500;
  transition: all 0.2s;
  cursor: pointer;
}

.link-btn:hover {
  background: rgba(99, 102, 241, 0.1);
  border-color: var(--accent-light);
  transform: translateX(5px);
}
</style>
