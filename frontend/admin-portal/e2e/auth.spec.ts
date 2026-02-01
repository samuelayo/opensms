import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display login page', async ({ page }) => {
    await expect(page).toHaveTitle(/OpenSMS/)
    await expect(page.locator('h1')).toContainText('Login')
  })

  test('should show validation errors for empty fields', async ({ page }) => {
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Email is required')).toBeVisible()
    await expect(page.locator('text=Password is required')).toBeVisible()
  })

  test('should show error for invalid credentials', async ({ page }) => {
    await page.fill('input[name="email"]', 'invalid@example.com')
    await page.fill('input[name="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Invalid credentials')).toBeVisible()
  })

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin@123')
    await page.click('button[type="submit"]')

    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('h1')).toContainText('Dashboard')
  })

  test('should persist authentication on page reload', async ({ page }) => {
    // Login first
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin@123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')

    // Reload page
    await page.reload()

    // Should still be on dashboard
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('h1')).toContainText('Dashboard')
  })

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin@123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')

    // Click logout
    await page.click('button[aria-label="Logout"]')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/dashboard')

    // Should be redirected to login
    await expect(page).toHaveURL(/\/login/)
    await expect(page).toHaveURL(/redirect/)
  })

  test('should preserve redirect query after login', async ({ page }) => {
    await page.goto('/dashboard')

    // Should redirect to login with query param
    await expect(page).toHaveURL(/\/login\?redirect=/)

    // Login
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin@123')
    await page.click('button[type="submit"]')

    // Should redirect back to dashboard
    await expect(page).toHaveURL('/dashboard')
  })

  test('should show password visibility toggle', async ({ page }) => {
    const passwordInput = page.locator('input[name="password"]')
    const toggleButton = page.locator('button[aria-label="Toggle password visibility"]')

    await expect(passwordInput).toHaveAttribute('type', 'password')

    await toggleButton.click()
    await expect(passwordInput).toHaveAttribute('type', 'text')

    await toggleButton.click()
    await expect(passwordInput).toHaveAttribute('type', 'password')
  })

  test('should handle remember me functionality', async ({ page }) => {
    const rememberCheckbox = page.locator('input[name="remember"]')

    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin@123')
    await rememberCheckbox.check()
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')

    // Close and reopen browser
    const context = page.context()
    await context.close()

    // Should still be authenticated
    // This would require proper session handling
  })
})

test.describe('Password Reset', () => {
  test('should navigate to forgot password page', async ({ page }) => {
    await page.goto('/login')
    await page.click('text=Forgot password?')

    await expect(page).toHaveURL('/forgot-password')
    await expect(page.locator('h1')).toContainText('Reset Password')
  })

  test('should submit password reset request', async ({ page }) => {
    await page.goto('/forgot-password')

    await page.fill('input[name="email"]', 'user@example.com')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Password reset email sent')).toBeVisible()
  })

  test('should validate email format on password reset', async ({ page }) => {
    await page.goto('/forgot-password')

    await page.fill('input[name="email"]', 'invalid-email')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Invalid email format')).toBeVisible()
  })
})

test.describe('Multi-factor Authentication', () => {
  test('should prompt for 2FA code after login', async ({ page }) => {
    await page.goto('/login')

    await page.fill('input[name="email"]', 'user-with-2fa@example.com')
    await page.fill('input[name="password"]', 'Password@123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/verify-2fa')
    await expect(page.locator('h1')).toContainText('Two-Factor Authentication')
  })

  test('should verify 2FA code', async ({ page }) => {
    await page.goto('/login')

    await page.fill('input[name="email"]', 'user-with-2fa@example.com')
    await page.fill('input[name="password"]', 'Password@123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/verify-2fa')

    await page.fill('input[name="code"]', '123456')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')
  })

  test('should show error for invalid 2FA code', async ({ page }) => {
    await page.goto('/verify-2fa')

    await page.fill('input[name="code"]', '000000')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Invalid verification code')).toBeVisible()
  })
})
