import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/services/api'
import type { User, LoginCredentials, TokenResponse } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem('accessToken'))
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'))
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!accessToken.value && !!user.value)
  const userRole = computed(() => user.value?.role || '')
  const userPermissions = computed(() => user.value?.permissions || [])

  // Actions
  async function login(credentials: LoginCredentials) {
    loading.value = true
    error.value = null

    try {
      const response = await api.post<TokenResponse>('/auth/login', credentials)
      const { access_token, refresh_token, user: userData } = response.data

      // Store tokens
      accessToken.value = access_token
      refreshToken.value = refresh_token
      user.value = userData

      // Persist to localStorage
      localStorage.setItem('accessToken', access_token)
      localStorage.setItem('refreshToken', refresh_token)

      // Set default authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`

      return true
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      // Clear state
      user.value = null
      accessToken.value = null
      refreshToken.value = null

      // Clear localStorage
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')

      // Clear authorization header
      delete api.defaults.headers.common['Authorization']
    }
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) return false

    try {
      const response = await api.post<TokenResponse>('/auth/refresh', {
        refresh_token: refreshToken.value
      })

      const { access_token, refresh_token: newRefreshToken } = response.data

      // Update tokens
      accessToken.value = access_token
      if (newRefreshToken) {
        refreshToken.value = newRefreshToken
        localStorage.setItem('refreshToken', newRefreshToken)
      }

      localStorage.setItem('accessToken', access_token)
      api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`

      return true
    } catch (err) {
      console.error('Token refresh failed:', err)
      await logout()
      return false
    }
  }

  async function checkAuth() {
    if (!accessToken.value) {
      return false
    }

    try {
      // Set authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${accessToken.value}`

      // Get current user
      const response = await api.get<User>('/auth/me')
      user.value = response.data

      return true
    } catch (err) {
      // Try to refresh token
      const refreshed = await refreshAccessToken()
      if (refreshed) {
        return checkAuth()
      }
      return false
    }
  }

  function hasPermission(permission: string): boolean {
    if (!user.value) return false

    // Super admin has all permissions
    if (user.value.role === 'super_admin') return true

    // Check if user has the specific permission
    return userPermissions.value.includes(permission)
  }

  function hasAnyPermission(permissions: string[]): boolean {
    return permissions.some(permission => hasPermission(permission))
  }

  function hasAllPermissions(permissions: string[]): boolean {
    return permissions.every(permission => hasPermission(permission))
  }

  return {
    // State
    user,
    accessToken,
    loading,
    error,

    // Getters
    isAuthenticated,
    userRole,
    userPermissions,

    // Actions
    login,
    logout,
    refreshAccessToken,
    checkAuth,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions
  }
})
