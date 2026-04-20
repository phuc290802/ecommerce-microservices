<template>
  <div class="admin-page">
    <header class="section-header">
      <div>
        <h1 class="text-2xl font-bold">👥 Quản lý nhân sự</h1>
        <p class="text-secondary">Tạo và quản lý tài khoản quản trị viên</p>
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">+ Thêm Admin mới</button>
    </header>

    <div class="glass-card mt-6">
      <table class="admin-table">
        <thead>
          <tr>
            <th>Username</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
            <th>Ngày tạo</th>
            <th>Thao tác</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td><strong>{{ user.username }}</strong></td>
            <td>{{ user.email }}</td>
            <td><span class="badge" :class="'badge-' + user.role">{{ user.role }}</span></td>
            <td><span class="status-dot" :class="user.status"></span> {{ user.status }}</td>
            <td>{{ new Date(user.created_at).toLocaleDateString() }}</td>
            <td>
              <button class="btn-icon">✏️</button>
              <button class="btn-icon text-error">🔒</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useAdminStore } from '../../stores/admin'

const admin = useAdminStore()
const users = ref([])
const showCreateModal = ref(false)

const loadUsers = async () => {
  try {
    const res = await axios.get('/api/admin/users', { headers: admin.adminHeaders })
    users.value = res.data || []
  } catch (err) {
    console.error(err)
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.admin-page { padding: 2rem; }
.section-header { display: flex; justify-content: space-between; align-items: center; }

.admin-table {
  width: 100%;
  border-collapse: collapse;
}

.admin-table th {
  text-align: left;
  padding: 1rem;
  color: #64748b;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.admin-table td {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
}

.badge {
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.05);
}

.badge-super_admin { color: #f59e0b; background: rgba(245, 158, 11, 0.1); }
.badge-product_manager { color: #10b981; background: rgba(16, 185, 129, 0.1); }

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 0.5rem;
}
.active { background: #10b981; box-shadow: 0 0 8px #10b981; }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.1rem;
  padding: 0.25rem;
  margin-right: 0.5rem;
}
</style>
