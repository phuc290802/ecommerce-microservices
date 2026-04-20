<template>
  <div class="page-container">
    <div class="section-header">
      <div>
        <h1 class="page-title">📋 Đơn hàng</h1>
        <p class="page-subtitle">Quản lý và theo dõi tất cả đơn hàng</p>
      </div>
      <button class="btn btn-primary" @click="loadOrders" :disabled="loading">
        <span v-if="loading" class="spinner"></span>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 .49-3.49"/></svg>
        Làm mới
      </button>
    </div>

    <!-- Summary cards -->
    <div class="grid-3 mb-3" v-if="orders.length">
      <div class="stat-card card">
        <div class="stat-icon">📦</div>
        <div>
          <div class="stat-value">{{ orders.length }}</div>
          <div class="stat-label">Tổng đơn hàng</div>
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-icon">💰</div>
        <div>
          <div class="stat-value">${{ totalRevenue.toFixed(2) }}</div>
          <div class="stat-label">Tổng giá trị</div>
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-icon">🛍️</div>
        <div>
          <div class="stat-value">{{ totalItems }}</div>
          <div class="stat-label">Tổng số lượng</div>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width: 2rem; height: 2rem;"></div>
      <p class="text-muted mt-1">Đang tải đơn hàng...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="alert alert-error mb-2">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
      {{ error }}
    </div>

    <!-- Orders table -->
    <div v-else-if="orders.length" class="card p-0 overflow-hidden">
      <table class="data-table">
        <thead>
          <tr>
            <th>Mã đơn</th>
            <th>Sản phẩm</th>
            <th>Số lượng</th>
            <th>Tổng tiền</th>
            <th>Trạng thái</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in orders" :key="order.id" @click="selectedOrder = order">
            <td>
              <span class="order-id">#{{ order.id }}</span>
            </td>
            <td>
              <div class="product-cell">
                <span class="product-dot">{{ getEmoji(order.product) }}</span>
                <span>{{ order.product }}</span>
              </div>
            </td>
            <td>
              <span class="qty-badge">{{ order.quantity }}x</span>
            </td>
            <td>
              <span class="price-cell">${{ order.total_cost.toFixed(2) }}</span>
            </td>
            <td>
              <span class="badge badge-success">✓ Hoàn thành</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty -->
    <div v-else class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>Chưa có đơn hàng</h3>
      <p>Bắt đầu mua sắm để tạo đơn hàng đầu tiên</p>
      <router-link to="/products" class="btn btn-primary" style="margin-top: 1rem;">
        Khám phá sản phẩm
      </router-link>
    </div>

    <!-- Order Detail Modal -->
    <div v-if="selectedOrder" class="modal-overlay" @click.self="selectedOrder = null">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Chi tiết đơn hàng #{{ selectedOrder.id }}</h3>
          <button class="btn-close" @click="selectedOrder = null">✕</button>
        </div>
        <div class="modal-body">
          <div class="detail-row">
            <span class="detail-label">Sản phẩm</span>
            <span>{{ selectedOrder.product }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Số lượng</span>
            <span>{{ selectedOrder.quantity }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Đơn giá</span>
            <span>${{ (selectedOrder.total_cost / selectedOrder.quantity).toFixed(2) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Tổng tiền</span>
            <span class="price-cell">${{ selectedOrder.total_cost.toFixed(2) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Trạng thái</span>
            <span class="badge badge-success">Hoàn thành</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const orders = ref([])
const loading = ref(true)
const error = ref('')
const selectedOrder = ref(null)

const totalRevenue = computed(() => orders.value.reduce((s, o) => s + o.total_cost, 0))
const totalItems = computed(() => orders.value.reduce((s, o) => s + o.quantity, 0))

const EMOJIS = {
  'shirt': '👕', 't-shirt': '👕', 'sneaker': '👟', 'shoe': '👟',
  'mug': '☕', 'coffee': '☕', 'home': '🏠',
}
const getEmoji = (name) => {
  if (!name) return '📦'
  const lower = name.toLowerCase()
  for (const [key, emoji] of Object.entries(EMOJIS)) {
    if (lower.includes(key)) return emoji
  }
  return '📦'
}

const loadOrders = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get('/api/orders', { headers: auth.authHeaders })
    orders.value = res.data || []
  } catch (err) {
    error.value = err.response?.data?.trim() || err.message
  } finally {
    loading.value = false
  }
}

onMounted(loadOrders)
</script>

<style scoped>
.loading-state {
  text-align: center;
  padding: 5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.stat-icon { font-size: 2rem; }
.stat-value {
  font-size: 1.5rem;
  font-weight: 800;
  background: linear-gradient(135deg, #f0f4ff, var(--accent-light));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.stat-label { font-size: 0.8rem; color: var(--text-secondary); }

.order-id { font-family: monospace; font-weight: 700; color: var(--accent-light); }

.product-cell { display: flex; align-items: center; gap: 0.5rem; }
.product-dot { font-size: 1.25rem; }

.qty-badge {
  background: rgba(255,255,255,0.06);
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 700;
}

.price-cell {
  font-weight: 700;
  background: linear-gradient(135deg, #10b981, #059669);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}
.modal-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 20px;
  width: 100%;
  max-width: 480px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
}
.modal-header h3 { font-size: 1.1rem; font-weight: 700; }
.btn-close {
  background: rgba(255,255,255,0.06);
  border: 1px solid var(--border);
  border-radius: 50%;
  width: 32px; height: 32px;
  cursor: pointer;
  color: var(--text-secondary);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}
.btn-close:hover { background: rgba(255,255,255,0.12); color: var(--text-primary); }
.modal-body { padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
.detail-row { display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 0; border-bottom: 1px solid rgba(255,255,255,0.04); }
.detail-row:last-child { border-bottom: none; }
.detail-label { color: var(--text-secondary); font-size: 0.875rem; }
</style>
