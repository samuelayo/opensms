import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Add tenant ID from subdomain or config
    const tenantId = getTenantId()
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }

    // Add authorization token
    const token = localStorage.getItem('accessToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Handle 401 Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        // Try to refresh token
        const refreshToken = localStorage.getItem('refreshToken')
        if (!refreshToken) {
          throw new Error('No refresh token')
        }

        const response = await axios.post(
          `${api.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken }
        )

        const { access_token, refresh_token: newRefreshToken } = response.data

        // Update tokens
        localStorage.setItem('accessToken', access_token)
        if (newRefreshToken) {
          localStorage.setItem('refreshToken', newRefreshToken)
        }

        // Retry original request with new token
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        // Refresh failed, logout user
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    // Handle other errors
    if (error.response?.status === 403) {
      // Forbidden - insufficient permissions
      console.error('Insufficient permissions')
    } else if (error.response?.status === 429) {
      // Too many requests
      console.error('Rate limit exceeded')
    } else if (error.response?.status >= 500) {
      // Server error
      console.error('Server error occurred')
    }

    return Promise.reject(error)
  }
)

// Helper function to get tenant ID from subdomain
function getTenantId(): string | null {
  // Extract tenant from subdomain (e.g., school1.opensms.com -> school1)
  const hostname = window.location.hostname
  const parts = hostname.split('.')

  // If subdomain exists and not localhost
  if (parts.length >= 3 || (parts.length === 2 && parts[0] !== 'localhost')) {
    return parts[0]
  }

  // Fallback to environment variable or localStorage
  return import.meta.env.VITE_TENANT_ID || localStorage.getItem('tenantId')
}

export default api
