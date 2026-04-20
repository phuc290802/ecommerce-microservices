import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useAdminStore } from './stores/admin'

const routes = [
  { path: '/', redirect: '/products' },
  {
    path: '/login',
    component: () => import('./pages/LoginPage.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    component: () => import('./pages/RegisterPage.vue'),
    meta: { guest: true }
  },
  {
    path: '/products',
    component: () => import('./pages/ProductsPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/products/:id',
    component: () => import('./pages/ProductDetailPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders',
    component: () => import('./pages/OrdersPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/forgot-password',
    component: () => import('./pages/ForgotPasswordPage.vue'),
    meta: { guest: true }
  },
  {
    path: '/admin/login',
    component: () => import('./pages/admin/AdminLogin.vue'),
    meta: { adminGuest: true }
  },
  {
    path: '/admin',
    component: () => import('./layouts/AdminLayout.vue'),
    meta: { requiresAdmin: true },
    children: [
      { path: '', component: () => import('./pages/admin/AdminDashboard.vue') },
      { path: 'users', component: () => import('./pages/admin/AdminUsers.vue') },
    ]
  },
  {
    path: '/reset-password',
    component: () => import('./pages/ResetPasswordPage.vue'),
    meta: { guest: true }
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() { return { top: 0 } }
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const admin = useAdminStore()

  // Admin Guards
  if (to.meta.requiresAdmin && !admin.isAdminAuthenticated) return '/admin/login'
  if (to.meta.adminGuest && admin.isAdminAuthenticated) return '/admin'

  // User Guards
  if (to.meta.requiresAuth && !auth.token) return '/login'
  if (to.meta.guest && auth.token) return '/products'
})

export default router
