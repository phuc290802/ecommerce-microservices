<template>
  <div class="page-container">
    <!-- Header -->
    <div class="section-header">
      <div class="header-main">
        <h1 class="page-title">🛍️ Sản phẩm</h1>
        <p class="page-subtitle">Khám phá tất cả sản phẩm của chúng tôi</p>
      </div>
      <div class="filter-actions-row">
        <div class="search-wrapper">
          <span class="search-icon">🔍</span>
          <input
            v-model="search"
            class="custom-input"
            placeholder="Tìm kiếm sản phẩm..."
          />
        </div>
        <div class="select-wrapper">
          <select v-model="categoryFilter" class="custom-select">
            <option value="">Tất cả danh mục</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <span class="select-arrow">▾</span>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width: 2rem; height: 2rem;"></div>
      <p class="text-muted mt-1">Đang tải sản phẩm...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="alert alert-error mb-2">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
      {{ error }}
      <button class="btn btn-sm btn-secondary" @click="loadAll" style="margin-left:auto">Thử lại</button>
    </div>

    <!-- Products Grid -->
    <div v-else>
      <div v-if="filteredProducts.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <h3>Không tìm thấy sản phẩm</h3>
        <p>Thử tìm kiếm với từ khóa khác</p>
      </div>
      <div v-else class="grid-3">
        <div
          v-for="product in filteredProducts"
          :key="product.id"
          class="product-card card"
          @click="$router.push(`/products/${product.id}`)"
        >
          <div class="product-image">
            <span class="product-emoji">{{ getEmoji(product.name) }}</span>
          </div>
          <div class="product-body">
            <div class="flex justify-between items-center mb-1">
              <span class="badge badge-accent">
                {{ getCategoryName(product.category_id) }}
              </span>
              <span class="product-id text-xs text-muted">#{{ product.id }}</span>
            </div>
            <h3 class="product-name">{{ product.name }}</h3>
            <div class="product-meta">
              <span class="product-price">${{ product.price.toFixed(2) }}</span>
              <span class="product-date text-xs text-muted">{{ formatDate(product.created_at) }}</span>
            </div>
          </div>
          <div class="product-footer">
            <button class="btn btn-primary btn-sm" @click.stop="$router.push(`/products/${product.id}`)">
              Xem chi tiết →
            </button>
          </div>
        </div>
      </div>

      <!-- Stats -->
      <div class="stats-bar" v-if="!loading && products.length">
        <span class="text-sm text-muted">
          Hiển thị {{ filteredProducts.length }} / {{ products.length }} sản phẩm
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const products = ref([])
const categories = ref([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const categoryFilter = ref('')

const EMOJIS = {
  'shirt': '👕', 't-shirt': '👕', 'áo': '👕', 'clothing': '👗',
  'sneaker': '👟', 'shoe': '👟', 'giày': '👟', 'footwear': '👟',
  'mug': '☕', 'coffee': '☕', 'cốc': '☕', 'home': '🏠',
  'phone': '📱', 'laptop': '💻', 'watch': '⌚',
}

const getEmoji = (name) => {
  const lower = name.toLowerCase()
  for (const [key, emoji] of Object.entries(EMOJIS)) {
    if (lower.includes(key)) return emoji
  }
  return '📦'
}

const getCategoryName = (id) => {
  return categories.value.find(c => c.id === id)?.name || `Cat ${id}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('vi-VN')
}

const filteredProducts = computed(() => {
  let list = products.value
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(p => p.name.toLowerCase().includes(q))
  }
  if (categoryFilter.value !== '') {
    list = list.filter(p => p.category_id === Number(categoryFilter.value))
  }
  return list
})

const loadAll = async () => {
  loading.value = true
  error.value = ''
  try {
    const [prodRes, catRes] = await Promise.all([
      axios.get('/api/products', { headers: auth.authHeaders }),
      axios.get('/api/bff/summary?product_id=1', { headers: auth.authHeaders })
        .then(() => null).catch(() => null) // optional
    ])
    products.value = prodRes.data || []

    // Try to get categories from direct services (not gated by auth)
    try {
      const _ = await axios.get('/api/bff/summary?product_id=1', { headers: auth.authHeaders })
    } catch {}

    // Build categories from products
    const catIds = [...new Set(products.value.map(p => p.category_id))]
    categories.value = catIds.map(id => {
      const names = { 1: 'Clothing', 2: 'Footwear', 3: 'Home' }
      return { id, name: names[id] || `Category ${id}` }
    })
  } catch (err) {
    error.value = err.response?.data?.trim() || err.message
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
</script>

<style scoped>
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2.5rem;
  gap: 1.5rem;
}

.filter-actions-row {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.search-wrapper, .select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.875rem;
  font-size: 0.9rem;
  color: var(--text-secondary);
  pointer-events: none;
  opacity: 0.7;
}

.custom-input, .custom-select {
  background: rgba(15, 22, 41, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 0.875rem;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  height: 42px;
}

.custom-input {
  width: 260px;
  padding: 0 1rem 0 2.5rem;
}

.custom-select {
  width: 180px;
  padding: 0 2.5rem 0 1rem;
  appearance: none;
  cursor: pointer;
}

.select-arrow {
  position: absolute;
  right: 1rem;
  pointer-events: none;
  color: var(--text-secondary);
  font-size: 0.8rem;
  opacity: 0.6;
}

.custom-input:focus, .custom-select:focus {
  outline: none;
  border-color: var(--accent-light);
  background: rgba(15, 22, 41, 0.8);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
}

.custom-select option {
  background: #0f1629;
  color: var(--text-primary);
}

.loading-state {
  text-align: center;
  padding: 5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.product-card {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
  overflow: hidden;
  border-radius: 20px; /* Bo tròn mạnh hơn */
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.product-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.4);
}

.product-image {
  background: linear-gradient(135deg, rgba(99,102,241,0.1) 0%, rgba(139,92,246,0.1) 100%);
  height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.product-emoji { font-size: 4rem; }

.product-body {
  padding: 1.25rem;
  flex: 1;
}

.product-name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.75rem;
}

.product-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.product-price {
  font-size: 1.25rem;
  font-weight: 800;
  background: linear-gradient(135deg, #10b981, #059669);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.product-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255,255,255,0.04);
  background: rgba(255,255,255,0.02);
}

.stats-bar {
  margin-top: 2rem;
  text-align: center;
}
</style>
