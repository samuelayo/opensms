import { test, expect } from '@playwright/test'

test.describe('Student Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[name="email"]', 'teacher@example.com')
    await page.fill('input[name="password"]', 'Teacher@123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/dashboard')

    // Navigate to students page
    await page.click('text=Students')
    await expect(page).toHaveURL('/students')
  })

  test('should display students list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Students')
    await expect(page.locator('table')).toBeVisible()
    await expect(page.locator('tbody tr')).toHaveCount({ min: 1 })
  })

  test('should search for students', async ({ page }) => {
    await page.fill('input[placeholder="Search students"]', 'John')
    await page.waitForTimeout(500) // Debounce

    const rows = page.locator('tbody tr')
    await expect(rows).toHaveCount({ min: 1 })
    await expect(rows.first()).toContainText('John')
  })

  test('should filter students by grade', async ({ page }) => {
    await page.selectOption('select[name="grade"]', '10')
    await page.waitForTimeout(500)

    const rows = page.locator('tbody tr')
    await expect(rows).toHaveCount({ min: 1 })
  })

  test('should open create student modal', async ({ page }) => {
    await page.click('button:has-text("Add Student")')

    await expect(page.locator('dialog')).toBeVisible()
    await expect(page.locator('dialog h2')).toContainText('Add New Student')
  })

  test('should create new student', async ({ page }) => {
    await page.click('button:has-text("Add Student")')

    await page.fill('input[name="first_name"]', 'Test')
    await page.fill('input[name="last_name"]', 'Student')
    await page.fill('input[name="email"]', 'test.student@example.com')
    await page.fill('input[name="student_id"]', 'STU999')
    await page.selectOption('select[name="grade_level"]', '10')
    await page.fill('input[name="date_of_birth"]', '2010-01-01')

    await page.click('button[type="submit"]')

    await expect(page.locator('text=Student created successfully')).toBeVisible()
    await expect(page.locator('tbody tr:has-text("Test Student")')).toBeVisible()
  })

  test('should validate required fields when creating student', async ({ page }) => {
    await page.click('button:has-text("Add Student")')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=First name is required')).toBeVisible()
    await expect(page.locator('text=Last name is required')).toBeVisible()
    await expect(page.locator('text=Email is required')).toBeVisible()
  })

  test('should view student details', async ({ page }) => {
    await page.click('tbody tr:first-child button:has-text("View")')

    await expect(page).toHaveURL(/\/students\/\w+/)
    await expect(page.locator('h1')).toContainText('Student Details')
  })

  test('should edit student information', async ({ page }) => {
    await page.click('tbody tr:first-child button[aria-label="Edit"]')

    await expect(page.locator('dialog h2')).toContainText('Edit Student')

    await page.fill('input[name="first_name"]', 'Updated')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=Student updated successfully')).toBeVisible()
    await expect(page.locator('tbody tr:has-text("Updated")')).toBeVisible()
  })

  test('should delete student', async ({ page }) => {
    await page.click('tbody tr:first-child button[aria-label="Delete"]')

    await expect(page.locator('dialog h2')).toContainText('Confirm Delete')
    await page.click('button:has-text("Delete")')

    await expect(page.locator('text=Student deleted successfully')).toBeVisible()
  })

  test('should cancel delete operation', async ({ page }) => {
    const initialCount = await page.locator('tbody tr').count()

    await page.click('tbody tr:first-child button[aria-label="Delete"]')
    await page.click('button:has-text("Cancel")')

    const finalCount = await page.locator('tbody tr').count()
    expect(finalCount).toBe(initialCount)
  })

  test('should paginate through students', async ({ page }) => {
    // Assuming there are more than 10 students
    await expect(page.locator('button:has-text("Next")')).toBeVisible()

    await page.click('button:has-text("Next")')
    await expect(page).toHaveURL(/page=2/)

    await page.click('button:has-text("Previous")')
    await expect(page).toHaveURL(/page=1/)
  })

  test('should change items per page', async ({ page }) => {
    await page.selectOption('select[name="perPage"]', '25')
    await page.waitForTimeout(500)

    const rows = page.locator('tbody tr')
    await expect(rows).toHaveCount({ min: 1, max: 25 })
  })

  test('should sort students by name', async ({ page }) => {
    await page.click('th:has-text("Name")')
    await page.waitForTimeout(500)

    // Check ascending order
    const firstRow = await page.locator('tbody tr:first-child td:first-child').textContent()

    await page.click('th:has-text("Name")')
    await page.waitForTimeout(500)

    // Check descending order
    const newFirstRow = await page.locator('tbody tr:first-child td:first-child').textContent()
    expect(newFirstRow).not.toBe(firstRow)
  })

  test('should export students list', async ({ page }) => {
    const downloadPromise = page.waitForEvent('download')
    await page.click('button:has-text("Export")')

    const download = await downloadPromise
    expect(download.suggestedFilename()).toMatch(/students.*\.csv/)
  })

  test('should bulk select students', async ({ page }) => {
    await page.click('thead input[type="checkbox"]')

    const checkboxes = page.locator('tbody input[type="checkbox"]')
    const count = await checkboxes.count()

    for (let i = 0; i < count; i++) {
      await expect(checkboxes.nth(i)).toBeChecked()
    }
  })

  test('should bulk delete students', async ({ page }) => {
    await page.click('tbody tr:first-child input[type="checkbox"]')
    await page.click('tbody tr:nth-child(2) input[type="checkbox"]')

    await page.click('button:has-text("Delete Selected")')
    await page.click('dialog button:has-text("Delete")')

    await expect(page.locator('text=2 students deleted successfully')).toBeVisible()
  })
})

test.describe('Student Details', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[name="email"]', 'teacher@example.com')
    await page.fill('input[name="password"]', 'Teacher@123')
    await page.click('button[type="submit"]')
    await page.click('text=Students')
    await page.click('tbody tr:first-child button:has-text("View")')
  })

  test('should display student information tabs', async ({ page }) => {
    await expect(page.locator('button:has-text("Overview")')).toBeVisible()
    await expect(page.locator('button:has-text("Grades")')).toBeVisible()
    await expect(page.locator('button:has-text("Attendance")')).toBeVisible()
    await expect(page.locator('button:has-text("Payments")')).toBeVisible()
  })

  test('should switch between tabs', async ({ page }) => {
    await page.click('button:has-text("Grades")')
    await expect(page.locator('h2:has-text("Academic Performance")')).toBeVisible()

    await page.click('button:has-text("Attendance")')
    await expect(page.locator('h2:has-text("Attendance Record")')).toBeVisible()
  })

  test('should display student grades', async ({ page }) => {
    await page.click('button:has-text("Grades")')
    await expect(page.locator('table')).toBeVisible()
  })

  test('should display attendance records', async ({ page }) => {
    await page.click('button:has-text("Attendance")')
    await expect(page.locator('.attendance-calendar')).toBeVisible()
  })

  test('should display payment history', async ({ page }) => {
    await page.click('button:has-text("Payments")')
    await expect(page.locator('table')).toBeVisible()
  })
})
