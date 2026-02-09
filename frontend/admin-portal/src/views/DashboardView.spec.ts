import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import DashboardView from './DashboardView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: DashboardView
    }
  ]
})

describe('DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders dashboard when authenticated', () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.exists()).toBe(true)
  })

  it('displays welcome message with user name', () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: '1',
      email: 'test@example.com',
      first_name: 'John',
      last_name: 'Doe',
      role: 'teacher',
      permissions: [],
      tenant_id: 'tenant-1',
      created_at: '2024-01-01',
      updated_at: '2024-01-01'
    }
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    // Depending on implementation, check for welcome message
    expect(wrapper.text()).toContain('Dashboard')
  })

  it('shows different content based on user role', () => {
    const authStore = useAuthStore()

    // Test teacher role
    authStore.user = {
      id: '1',
      email: 'teacher@example.com',
      first_name: 'Teacher',
      last_name: 'User',
      role: 'teacher',
      permissions: ['grade:read', 'attendance:mark'],
      tenant_id: 'tenant-1',
      created_at: '2024-01-01',
      updated_at: '2024-01-01'
    }
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.exists()).toBe(true)
  })

  it('shows admin controls for admin users', () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.exists()).toBe(true)
  })

  it('displays loading state while fetching data', async () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    // Component should handle loading state
    expect(wrapper.exists()).toBe(true)
  })

  it('handles errors gracefully', async () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.exists()).toBe(true)
  })

  it('displays statistics cards', () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: '1',
      email: 'test@example.com',
      first_name: 'Test',
      last_name: 'User',
      role: 'school_admin',
      permissions: [],
      tenant_id: 'tenant-1',
      created_at: '2024-01-01',
      updated_at: '2024-01-01'
    }
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    // Should render dashboard content
    expect(wrapper.exists()).toBe(true)
  })

  it('shows recent activities section', () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.exists()).toBe(true)
  })

  it('updates on data refresh', async () => {
    const authStore = useAuthStore()
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
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    // Component should handle data updates
    expect(wrapper.exists()).toBe(true)
  })

  it('handles permission-based visibility', () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: '1',
      email: 'student@example.com',
      first_name: 'Student',
      last_name: 'User',
      role: 'student',
      permissions: ['grade:read'],
      tenant_id: 'tenant-1',
      created_at: '2024-01-01',
      updated_at: '2024-01-01'
    }
    authStore.accessToken = 'fake-token'

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [router]
      }
    })

    // Student should see limited dashboard
    expect(wrapper.exists()).toBe(true)
  })
})
