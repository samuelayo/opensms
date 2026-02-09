import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach } from 'vitest'
import { useUIStore } from './ui'

describe('UI Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('Sidebar State', () => {
    it('initializes with sidebar open', () => {
      const store = useUIStore()
      expect(store.sidebarOpen).toBe(true)
    })

    it('toggles sidebar state', () => {
      const store = useUIStore()
      expect(store.sidebarOpen).toBe(true)

      store.toggleSidebar()
      expect(store.sidebarOpen).toBe(false)

      store.toggleSidebar()
      expect(store.sidebarOpen).toBe(true)
    })

    it('sets sidebar state explicitly', () => {
      const store = useUIStore()

      store.setSidebarOpen(false)
      expect(store.sidebarOpen).toBe(false)

      store.setSidebarOpen(true)
      expect(store.sidebarOpen).toBe(true)
    })

    it('closes sidebar on mobile', () => {
      const store = useUIStore()
      store.isMobile = true
      store.sidebarOpen = true

      store.closeSidebarOnMobile()
      expect(store.sidebarOpen).toBe(false)
    })

    it('does not close sidebar on desktop', () => {
      const store = useUIStore()
      store.isMobile = false
      store.sidebarOpen = true

      store.closeSidebarOnMobile()
      expect(store.sidebarOpen).toBe(true)
    })
  })

  describe('Loading State', () => {
    it('starts with no loading state', () => {
      const store = useUIStore()
      expect(store.loading).toBe(false)
    })

    it('sets loading state', () => {
      const store = useUIStore()

      store.setLoading(true)
      expect(store.loading).toBe(true)

      store.setLoading(false)
      expect(store.loading).toBe(false)
    })

    it('tracks multiple loading states', () => {
      const store = useUIStore()

      store.startLoading('api-call-1')
      expect(store.isLoading('api-call-1')).toBe(true)

      store.startLoading('api-call-2')
      expect(store.isLoading('api-call-2')).toBe(true)
      expect(store.isLoading('api-call-1')).toBe(true)

      store.stopLoading('api-call-1')
      expect(store.isLoading('api-call-1')).toBe(false)
      expect(store.isLoading('api-call-2')).toBe(true)
    })

    it('checks if any loading is active', () => {
      const store = useUIStore()
      expect(store.isAnyLoading).toBe(false)

      store.startLoading('task-1')
      expect(store.isAnyLoading).toBe(true)

      store.startLoading('task-2')
      expect(store.isAnyLoading).toBe(true)

      store.stopLoading('task-1')
      expect(store.isAnyLoading).toBe(true)

      store.stopLoading('task-2')
      expect(store.isAnyLoading).toBe(false)
    })
  })

  describe('Modal State', () => {
    it('starts with no active modal', () => {
      const store = useUIStore()
      expect(store.activeModal).toBeNull()
    })

    it('opens modal', () => {
      const store = useUIStore()

      store.openModal('confirm-delete')
      expect(store.activeModal).toBe('confirm-delete')
    })

    it('closes modal', () => {
      const store = useUIStore()
      store.activeModal = 'test-modal'

      store.closeModal()
      expect(store.activeModal).toBeNull()
    })

    it('checks if specific modal is open', () => {
      const store = useUIStore()

      store.openModal('user-form')
      expect(store.isModalOpen('user-form')).toBe(true)
      expect(store.isModalOpen('other-modal')).toBe(false)
    })
  })

  describe('Notification State', () => {
    it('starts with empty notifications', () => {
      const store = useUIStore()
      expect(store.notifications).toEqual([])
    })

    it('adds notification', () => {
      const store = useUIStore()

      store.addNotification({
        type: 'success',
        message: 'Operation successful'
      })

      expect(store.notifications).toHaveLength(1)
      expect(store.notifications[0].type).toBe('success')
      expect(store.notifications[0].message).toBe('Operation successful')
      expect(store.notifications[0].id).toBeDefined()
    })

    it('adds multiple notifications', () => {
      const store = useUIStore()

      store.addNotification({ type: 'info', message: 'Info 1' })
      store.addNotification({ type: 'warning', message: 'Warning 1' })
      store.addNotification({ type: 'error', message: 'Error 1' })

      expect(store.notifications).toHaveLength(3)
    })

    it('removes notification by id', () => {
      const store = useUIStore()

      store.addNotification({ type: 'info', message: 'Test 1' })
      store.addNotification({ type: 'info', message: 'Test 2' })

      const idToRemove = store.notifications[0].id
      store.removeNotification(idToRemove)

      expect(store.notifications).toHaveLength(1)
      expect(store.notifications[0].message).toBe('Test 2')
    })

    it('clears all notifications', () => {
      const store = useUIStore()

      store.addNotification({ type: 'info', message: 'Test 1' })
      store.addNotification({ type: 'info', message: 'Test 2' })
      store.addNotification({ type: 'info', message: 'Test 3' })

      store.clearNotifications()
      expect(store.notifications).toEqual([])
    })
  })

  describe('Theme State', () => {
    it('initializes with light theme', () => {
      const store = useUIStore()
      expect(store.theme).toBe('light')
    })

    it('toggles theme', () => {
      const store = useUIStore()

      store.toggleTheme()
      expect(store.theme).toBe('dark')

      store.toggleTheme()
      expect(store.theme).toBe('light')
    })

    it('sets theme explicitly', () => {
      const store = useUIStore()

      store.setTheme('dark')
      expect(store.theme).toBe('dark')

      store.setTheme('light')
      expect(store.theme).toBe('light')
    })

    it('checks if dark mode is active', () => {
      const store = useUIStore()
      expect(store.isDarkMode).toBe(false)

      store.setTheme('dark')
      expect(store.isDarkMode).toBe(true)
    })
  })

  describe('Breadcrumb State', () => {
    it('starts with empty breadcrumbs', () => {
      const store = useUIStore()
      expect(store.breadcrumbs).toEqual([])
    })

    it('sets breadcrumbs', () => {
      const store = useUIStore()
      const crumbs = [
        { label: 'Home', path: '/' },
        { label: 'Users', path: '/users' },
        { label: 'Details', path: '/users/123' }
      ]

      store.setBreadcrumbs(crumbs)
      expect(store.breadcrumbs).toEqual(crumbs)
    })

    it('clears breadcrumbs', () => {
      const store = useUIStore()
      store.breadcrumbs = [{ label: 'Test', path: '/test' }]

      store.clearBreadcrumbs()
      expect(store.breadcrumbs).toEqual([])
    })
  })

  describe('Page Title', () => {
    it('starts with default title', () => {
      const store = useUIStore()
      expect(store.pageTitle).toBe('OpenSMS')
    })

    it('sets page title', () => {
      const store = useUIStore()

      store.setPageTitle('User Management')
      expect(store.pageTitle).toBe('User Management')
    })

    it('resets to default title', () => {
      const store = useUIStore()
      store.pageTitle = 'Custom Title'

      store.resetPageTitle()
      expect(store.pageTitle).toBe('OpenSMS')
    })
  })
})
