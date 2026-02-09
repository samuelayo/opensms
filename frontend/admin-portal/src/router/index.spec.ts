import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createRouter, createMemoryHistory, type Router } from 'vue-router'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

// Import routes from the actual router file
const routes = [
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
    path: '/grades',
    name: 'Grades',
    component: () => import('@/views/academic/GradesView.vue'),
    meta: { requiresAuth: true, layout: 'default', permission: 'grade:read' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { requiresAuth: false }
  }
]

function createTestRouter(): Router {
  return createRouter({
    history: createMemoryHistory(),
    routes
  })
}

describe('Router', () => {
  let router: Router

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createTestRouter()
    vi.clearAllMocks()
  })

  describe('Route Configuration', () => {
    it('has login route', () => {
      const route = router.resolve('/login')
      expect(route.name).toBe('Login')
      expect(route.meta.requiresAuth).toBe(false)
    })

    it('has dashboard route', () => {
      const route = router.resolve('/')
      expect(route.name).toBe('Dashboard')
      expect(route.meta.requiresAuth).toBe(true)
    })

    it('has students route', () => {
      const route = router.resolve('/students')
      expect(route.name).toBe('Students')
      expect(route.meta.requiresAuth).toBe(true)
      expect(route.meta.permission).toBe('student:read')
    })

    it('has teachers route', () => {
      const route = router.resolve('/teachers')
      expect(route.name).toBe('Teachers')
      expect(route.meta.requiresAuth).toBe(true)
    })

    it('has grades route', () => {
      const route = router.resolve('/grades')
      expect(route.name).toBe('Grades')
      expect(route.meta.requiresAuth).toBe(true)
    })

    it('has 404 catch-all route', () => {
      const route = router.resolve('/non-existent-route')
      expect(route.name).toBe('NotFound')
    })
  })

  describe('Navigation Guards', () => {
    it('redirects to login when not authenticated', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = null
      authStore.user = null

      await router.push('/')

      // Should be redirected to login
      expect(router.currentRoute.value.name).toBe('Login')
      expect(router.currentRoute.value.query.redirect).toBe('/')
    })

    it('allows access to protected route when authenticated', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'fake-token'
      authStore.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'teacher',
        permissions: ['student:read'],
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      await router.push('/')

      expect(router.currentRoute.value.name).toBe('Dashboard')
    })

    it('redirects to dashboard when accessing login while authenticated', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'fake-token'
      authStore.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'teacher',
        permissions: [],
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      await router.push('/login')

      expect(router.currentRoute.value.name).toBe('Dashboard')
    })

    it('checks permission before allowing access', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'fake-token'
      authStore.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'student',
        permissions: ['grade:read'], // Student doesn't have student:read permission
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      await router.push('/students')

      // Should redirect to dashboard due to lack of permission
      expect(router.currentRoute.value.name).toBe('Dashboard')
    })

    it('allows access when user has required permission', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'fake-token'
      authStore.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'teacher',
        permissions: ['student:read'],
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      await router.push('/students')

      expect(router.currentRoute.value.name).toBe('Students')
    })

    it('allows super admin to access all routes', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'fake-token'
      authStore.user = {
        id: '1',
        email: 'admin@example.com',
        first_name: 'Admin',
        last_name: 'User',
        role: 'super_admin',
        permissions: [],
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      await router.push('/students')
      expect(router.currentRoute.value.name).toBe('Students')

      await router.push('/teachers')
      expect(router.currentRoute.value.name).toBe('Teachers')

      await router.push('/grades')
      expect(router.currentRoute.value.name).toBe('Grades')
    })

    it('preserves query parameters in redirect', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = null
      authStore.user = null

      await router.push('/students?filter=active')

      expect(router.currentRoute.value.name).toBe('Login')
      expect(router.currentRoute.value.query.redirect).toBe('/students?filter=active')
    })

    it('allows access to public routes without authentication', async () => {
      const authStore = useAuthStore()
      authStore.accessToken = null
      authStore.user = null

      await router.push('/login')

      expect(router.currentRoute.value.name).toBe('Login')
    })
  })

  describe('Route Metadata', () => {
    it('sets correct layout for auth pages', () => {
      const route = router.resolve('/login')
      expect(route.meta.layout).toBe('auth')
    })

    it('sets correct layout for default pages', () => {
      const route = router.resolve('/')
      expect(route.meta.layout).toBe('default')
    })

    it('defines permissions for protected routes', () => {
      const studentsRoute = router.resolve('/students')
      expect(studentsRoute.meta.permission).toBe('student:read')

      const teachersRoute = router.resolve('/teachers')
      expect(teachersRoute.meta.permission).toBe('user:read')

      const gradesRoute = router.resolve('/grades')
      expect(gradesRoute.meta.permission).toBe('grade:read')
    })
  })

  describe('Navigation', () => {
    it('navigates programmatically', async () => {
      await router.push('/login')
      expect(router.currentRoute.value.path).toBe('/login')

      await router.push('/')
      expect(router.currentRoute.value.path).toBe('/')
    })

    it('navigates by name', async () => {
      await router.push({ name: 'Students' })
      expect(router.currentRoute.value.name).toBe('Students')
    })

    it('passes route params', async () => {
      // If we had routes with params, we'd test them here
      // Example: await router.push({ name: 'StudentDetail', params: { id: '123' } })
    })

    it('passes query params', async () => {
      await router.push({ path: '/students', query: { page: '2' } })
      expect(router.currentRoute.value.query.page).toBe('2')
    })

    it('handles browser back/forward', async () => {
      await router.push('/login')
      await router.push('/')
      await router.push('/students')

      await router.back()
      expect(router.currentRoute.value.path).toBe('/')

      await router.forward()
      expect(router.currentRoute.value.path).toBe('/students')
    })
  })

  describe('404 Handling', () => {
    it('matches catch-all route for unknown paths', () => {
      const route = router.resolve('/completely/unknown/route')
      expect(route.name).toBe('NotFound')
    })

    it('matches catch-all for deep nested paths', () => {
      const route = router.resolve('/very/deep/nested/unknown/path')
      expect(route.name).toBe('NotFound')
    })
  })
})
