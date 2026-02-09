import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { requiresAuth: false, layout: 'auth' }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/DashboardView.vue'),
    meta: { requiresAuth: true, layout: 'default' }
  },
  {
    path: '/students',
    name: 'Students',
    component: () => import('@/views/students/StudentsView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'student:read' }
  },
  {
    path: '/teachers',
    name: 'Teachers',
    component: () => import('@/views/teachers/TeachersView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'user:read' }
  },
  {
    path: '/classes',
    name: 'Classes',
    component: () => import('@/views/academic/ClassesView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'class:read' }
  },
  {
    path: '/grades',
    name: 'Grades',
    component: () => import('@/views/academic/GradesView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'grade:read' }
  },
  {
    path: '/attendance',
    name: 'Attendance',
    component: () => import('@/views/academic/AttendanceView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'attendance:read' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue'),
    meta: { requiresAuth: true, layout: 'default' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { requiresAuth: false }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  if (requiresAuth && !authStore.isAuthenticated) {
    // Redirect to login if not authenticated
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    // Redirect to dashboard if already logged in
    next({ name: 'Dashboard' })
  } else {
    // Check permission if specified
    const permission = to.meta.permission as string
    if (permission && !authStore.hasPermission(permission)) {
      // Redirect to dashboard or show 403
      next({ name: 'Dashboard' })
    } else {
      next()
    }
  }
})

export default router
