import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useAuthStore } from './auth'
import api from '@/services/api'

// Mock the API
vi.mock('@/services/api')

describe('Auth Store', () => {
  beforeEach(() => {
    // Create a new pinia instance for each test
    setActivePinia(createPinia())

    // Clear localStorage
    localStorage.clear()

    // Clear all mocks
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should initialize with null user', () => {
      const store = useAuthStore()
      expect(store.user).toBeNull()
    })

    it('should initialize with token from localStorage if available', () => {
      localStorage.setItem('accessToken', 'test-token')
      const store = useAuthStore()
      expect(store.accessToken).toBe('test-token')
    })

    it('should not be authenticated initially', () => {
      const store = useAuthStore()
      expect(store.isAuthenticated).toBe(false)
    })
  })

  describe('login', () => {
    it('should login successfully with valid credentials', async () => {
      const store = useAuthStore()
      const mockResponse = {
        data: {
          access_token: 'mock-access-token',
          refresh_token: 'mock-refresh-token',
          user: {
            id: '1',
            email: 'test@example.com',
            first_name: 'Test',
            last_name: 'User',
            role: 'teacher',
            permissions: ['grade:read', 'grade:create'],
            tenant_id: 'tenant-1',
            created_at: '2024-01-01',
            updated_at: '2024-01-01'
          }
        }
      }

      vi.mocked(api.post).mockResolvedValue(mockResponse)

      const result = await store.login({
        email: 'test@example.com',
        password: 'password123'
      })

      expect(result).toBe(true)
      expect(store.user).toEqual(mockResponse.data.user)
      expect(store.accessToken).toBe('mock-access-token')
      expect(store.refreshToken).toBe('mock-refresh-token')
      expect(store.isAuthenticated).toBe(true)
      expect(localStorage.getItem('accessToken')).toBe('mock-access-token')
      expect(localStorage.getItem('refreshToken')).toBe('mock-refresh-token')
    })

    it('should handle login failure', async () => {
      const store = useAuthStore()
      const errorMessage = 'Invalid credentials'

      vi.mocked(api.post).mockRejectedValue({
        response: {
          data: {
            message: errorMessage
          }
        }
      })

      const result = await store.login({
        email: 'test@example.com',
        password: 'wrong-password'
      })

      expect(result).toBe(false)
      expect(store.error).toBe(errorMessage)
      expect(store.user).toBeNull()
      expect(store.isAuthenticated).toBe(false)
    })

    it('should set loading state during login', async () => {
      const store = useAuthStore()

      vi.mocked(api.post).mockImplementation(() => {
        expect(store.loading).toBe(true)
        return Promise.resolve({ data: {} })
      })

      await store.login({
        email: 'test@example.com',
        password: 'password123'
      })

      expect(store.loading).toBe(false)
    })
  })

  describe('logout', () => {
    it('should clear user data and tokens', async () => {
      const store = useAuthStore()

      // Set up initial logged-in state
      store.user = {
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
      store.accessToken = 'test-token'
      store.refreshToken = 'test-refresh'
      localStorage.setItem('accessToken', 'test-token')
      localStorage.setItem('refreshToken', 'test-refresh')

      vi.mocked(api.post).mockResolvedValue({})

      await store.logout()

      expect(store.user).toBeNull()
      expect(store.accessToken).toBeNull()
      expect(store.refreshToken).toBeNull()
      expect(localStorage.getItem('accessToken')).toBeNull()
      expect(localStorage.getItem('refreshToken')).toBeNull()
    })

    it('should clear data even if API call fails', async () => {
      const store = useAuthStore()
      store.accessToken = 'test-token'

      vi.mocked(api.post).mockRejectedValue(new Error('API Error'))

      await store.logout()

      expect(store.accessToken).toBeNull()
      expect(store.refreshToken).toBeNull()
    })
  })

  describe('Permissions', () => {
    beforeEach(() => {
      const store = useAuthStore()
      store.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'teacher',
        permissions: ['grade:read', 'grade:create', 'attendance:mark'],
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }
    })

    it('should check if user has specific permission', () => {
      const store = useAuthStore()

      expect(store.hasPermission('grade:read')).toBe(true)
      expect(store.hasPermission('grade:create')).toBe(true)
      expect(store.hasPermission('user:delete')).toBe(false)
    })

    it('should return true for super_admin regardless of permission', () => {
      const store = useAuthStore()
      store.user!.role = 'super_admin'

      expect(store.hasPermission('any:permission')).toBe(true)
      expect(store.hasPermission('unknown:permission')).toBe(true)
    })

    it('should check if user has any of the specified permissions', () => {
      const store = useAuthStore()

      expect(store.hasAnyPermission(['grade:read', 'user:delete'])).toBe(true)
      expect(store.hasAnyPermission(['user:delete', 'user:create'])).toBe(false)
    })

    it('should check if user has all of the specified permissions', () => {
      const store = useAuthStore()

      expect(store.hasAllPermissions(['grade:read', 'grade:create'])).toBe(true)
      expect(store.hasAllPermissions(['grade:read', 'user:delete'])).toBe(false)
    })

    it('should return false for permissions when no user', () => {
      const store = useAuthStore()
      store.user = null

      expect(store.hasPermission('grade:read')).toBe(false)
      expect(store.hasAnyPermission(['grade:read'])).toBe(false)
      expect(store.hasAllPermissions(['grade:read'])).toBe(false)
    })
  })

  describe('Computed Properties', () => {
    it('should compute isAuthenticated correctly', () => {
      const store = useAuthStore()

      expect(store.isAuthenticated).toBe(false)

      store.accessToken = 'test-token'
      expect(store.isAuthenticated).toBe(false) // Still false without user

      store.user = {
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
      expect(store.isAuthenticated).toBe(true)
    })

    it('should compute userRole correctly', () => {
      const store = useAuthStore()

      expect(store.userRole).toBe('')

      store.user = {
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

      expect(store.userRole).toBe('teacher')
    })

    it('should compute userPermissions correctly', () => {
      const store = useAuthStore()

      expect(store.userPermissions).toEqual([])

      const permissions = ['grade:read', 'grade:create']
      store.user = {
        id: '1',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        role: 'teacher',
        permissions,
        tenant_id: 'tenant-1',
        created_at: '2024-01-01',
        updated_at: '2024-01-01'
      }

      expect(store.userPermissions).toEqual(permissions)
    })
  })
})
