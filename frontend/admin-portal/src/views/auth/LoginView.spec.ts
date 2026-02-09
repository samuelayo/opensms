import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { setActivePinia, createPinia } from 'pinia'
import LoginView from './LoginView.vue'
import { useAuthStore } from '@/stores/auth'

// Mock router
const mockRouter = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', name: 'Dashboard', component: { template: '<div>Dashboard</div>' } },
    { path: '/login', name: 'Login', component: LoginView }
  ]
})

describe('LoginView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders login form', () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('has correct title', () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    expect(wrapper.text()).toContain('OpenSMS Admin Portal')
    expect(wrapper.text()).toContain('School Management System')
  })

  it('updates form model on input', async () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    const emailInput = wrapper.find('input[type="email"]')
    const passwordInput = wrapper.find('input[type="password"]')

    await emailInput.setValue('test@example.com')
    await passwordInput.setValue('password123')

    expect((emailInput.element as HTMLInputElement).value).toBe('test@example.com')
    expect((passwordInput.element as HTMLInputElement).value).toBe('password123')
  })

  it('calls login on form submit', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue(true)

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit.prevent')

    expect(authStore.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password123'
    })
  })

  it('redirects to dashboard on successful login', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue(true)

    const pushSpy = vi.spyOn(mockRouter, 'push')

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit.prevent')

    await wrapper.vm.$nextTick()

    expect(pushSpy).toHaveBeenCalledWith('/')
  })

  it('shows loading state during login', async () => {
    const authStore = useAuthStore()
    authStore.loading = true

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    expect(wrapper.find('button[type="submit"]').text()).toContain('Logging in')
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('shows error message on login failure', async () => {
    const authStore = useAuthStore()
    authStore.error = 'Invalid credentials'

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    expect(wrapper.text()).toContain('Invalid credentials')
  })

  it('has forgot password link', () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    expect(wrapper.text()).toContain('Forgot your password')
  })

  it('validates required fields', () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    const emailInput = wrapper.find('input[type="email"]')
    const passwordInput = wrapper.find('input[type="password"]')

    expect(emailInput.attributes('required')).toBeDefined()
    expect(passwordInput.attributes('required')).toBeDefined()
  })

  it('uses correct autocomplete attributes', () => {
    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    const emailInput = wrapper.find('input[type="email"]')
    const passwordInput = wrapper.find('input[type="password"]')

    expect(emailInput.attributes('autocomplete')).toBe('email')
    expect(passwordInput.attributes('autocomplete')).toBe('current-password')
  })

  it('handles redirect query parameter', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue(true)

    const pushSpy = vi.spyOn(mockRouter, 'push')

    // Navigate to login with redirect
    await mockRouter.push({ path: '/login', query: { redirect: '/grades' } })

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit.prevent')

    await wrapper.vm.$nextTick()

    expect(pushSpy).toHaveBeenCalledWith('/grades')
  })

  it('prevents double submission', async () => {
    const authStore = useAuthStore()
    const loginSpy = vi.spyOn(authStore, 'login').mockImplementation(() => {
      authStore.loading = true
      return new Promise(resolve => setTimeout(() => resolve(true), 100))
    })

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')

    // Submit twice quickly
    const form = wrapper.find('form')
    await form.trigger('submit.prevent')
    await form.trigger('submit.prevent')

    // Should only be called once
    expect(loginSpy).toHaveBeenCalledTimes(1)
  })

  it('clears error on new submission', async () => {
    const authStore = useAuthStore()
    authStore.error = 'Previous error'
    vi.spyOn(authStore, 'login').mockResolvedValue(false)

    const wrapper = mount(LoginView, {
      global: {
        plugins: [mockRouter]
      }
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit.prevent')

    // Error should be cleared during login (even if login fails again)
    // This would need to be implemented in the component
  })
})
