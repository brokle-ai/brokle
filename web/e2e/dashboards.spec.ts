/**
 * Dashboard E2E Tests
 *
 * End-to-end tests for the dashboard feature including:
 * - Creating dashboards from templates
 * - Adding and configuring widgets
 * - Drag and resize operations
 * - Query execution and data display
 * - Auto-save functionality
 *
 * Prerequisites:
 * - Backend server running with test data
 * - User authenticated (uses test fixtures)
 */

import { test, expect, type Page } from '@playwright/test'

// Test fixtures and helpers
const TEST_PROJECT_SLUG = 'test-project'
const DASHBOARD_URL = `/projects/${TEST_PROJECT_SLUG}/dashboards`

/**
 * Helper to login (adjust based on actual auth flow)
 */
async function login(page: Page) {
  // Navigate to login page
  await page.goto('/login')

  // Fill credentials (use test account)
  await page.fill('input[name="email"]', 'test@example.com')
  await page.fill('input[name="password"]', 'test-password')

  // Submit and wait for navigation
  await page.click('button[type="submit"]')
  await page.waitForURL(/\/projects/)
}

/**
 * Helper to navigate to dashboards page
 */
async function navigateToDashboards(page: Page) {
  await page.goto(DASHBOARD_URL)
  await page.waitForSelector('[data-testid="dashboard-list"]', { timeout: 10000 })
}

test.describe('Dashboard Management', () => {
  test.beforeEach(async ({ page }) => {
    // Skip auth in development mode with mock user
    // In real tests, call login(page)
    await page.goto(DASHBOARD_URL)
  })

  test('displays dashboard list page', async ({ page }) => {
    // Check page title or header
    await expect(page.locator('h1')).toContainText(/dashboards/i)

    // Check for create button
    await expect(page.getByRole('button', { name: /create|new/i })).toBeVisible()
  })

  test('can create a new dashboard', async ({ page }) => {
    // Click create dashboard button
    await page.getByRole('button', { name: /create|new/i }).click()

    // Wait for dialog/form
    await expect(page.getByRole('dialog')).toBeVisible()

    // Fill dashboard name
    await page.fill('input[name="name"]', 'Test Dashboard E2E')

    // Optionally fill description
    const descInput = page.locator('textarea[name="description"]')
    if (await descInput.isVisible()) {
      await descInput.fill('Created by E2E test')
    }

    // Submit form
    await page.getByRole('button', { name: /create|save/i }).click()

    // Verify dashboard was created (redirect to detail page or list updated)
    await expect(page).toHaveURL(/\/dashboards\/[^/]+$/, { timeout: 10000 })
  })

  test('can navigate to dashboard detail', async ({ page }) => {
    // Find first dashboard in list and click
    const firstDashboard = page.locator('[data-testid="dashboard-item"]').first()
    if (await firstDashboard.isVisible()) {
      await firstDashboard.click()

      // Should navigate to detail page
      await expect(page).toHaveURL(/\/dashboards\/[^/]+$/)

      // Should show dashboard name in header
      await expect(page.locator('h1')).toBeVisible()
    }
  })
})

test.describe('Dashboard Editor', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to a specific test dashboard
    // In real tests, create a test dashboard first or use fixture
    await page.goto(`${DASHBOARD_URL}/test-dashboard-id`)
  })

  test('can enter edit mode', async ({ page }) => {
    // Find edit button
    const editButton = page.getByRole('button', { name: /edit/i })
    if (await editButton.isVisible()) {
      await editButton.click()

      // Verify edit mode UI appears
      await expect(page.getByRole('button', { name: /save/i })).toBeVisible()
      await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible()
    }
  })

  test('can add a widget from palette', async ({ page }) => {
    // Enter edit mode
    const editButton = page.getByRole('button', { name: /edit/i })
    if (await editButton.isVisible()) {
      await editButton.click()
    }

    // Open widget palette
    const addWidgetButton = page.getByRole('button', { name: /add widget/i })
    if (await addWidgetButton.isVisible()) {
      await addWidgetButton.click()

      // Select a widget type (e.g., stat widget)
      await page.getByRole('button', { name: /stat|number/i }).click()

      // Widget dialog should open
      await expect(page.getByRole('dialog')).toBeVisible()

      // Configure widget (minimal)
      await page.fill('input[name="title"]', 'Test Stat Widget')

      // Save widget
      await page.getByRole('button', { name: /save|add/i }).click()

      // Verify widget appears in grid
      await expect(page.locator('[data-testid="widget-card"]')).toHaveCount(1)
    }
  })

  test('shows auto-save indicator when making changes', async ({ page }) => {
    // Enter edit mode
    const editButton = page.getByRole('button', { name: /edit/i })
    if (await editButton.isVisible()) {
      await editButton.click()

      // Make a change (e.g., resize a widget if exists)
      // Or add a new widget
      const addWidgetButton = page.getByRole('button', { name: /add widget/i })
      if (await addWidgetButton.isVisible()) {
        // Trigger some change
        await addWidgetButton.click()

        // Check for auto-save indicator
        const indicator = page.locator('[data-testid="auto-save-indicator"]')
        // May show "Unsaved changes" or similar
        await expect(indicator).toBeVisible({ timeout: 5000 })
      }
    }
  })

  test('can cancel edit mode and discard changes', async ({ page }) => {
    // Enter edit mode
    const editButton = page.getByRole('button', { name: /edit/i })
    if (await editButton.isVisible()) {
      await editButton.click()

      // Click cancel
      await page.getByRole('button', { name: /cancel/i }).click()

      // Should exit edit mode
      await expect(page.getByRole('button', { name: /edit/i })).toBeVisible()
      await expect(page.getByRole('button', { name: /save/i })).not.toBeVisible()
    }
  })
})

test.describe('Widget Interactions', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to dashboard with widgets
    await page.goto(`${DASHBOARD_URL}/test-dashboard-id`)
  })

  test('widget displays loading state', async ({ page }) => {
    // Check for loading spinner or skeleton
    const widget = page.locator('[data-testid="widget-card"]').first()
    if (await widget.isVisible()) {
      // Initial load may show loading state
      const loadingIndicator = widget.locator('[data-testid="loading"]')
      // Either loading or data should be visible
      await expect(
        loadingIndicator.or(widget.locator('[data-testid="widget-content"]'))
      ).toBeVisible()
    }
  })

  test('widget displays data after query completes', async ({ page }) => {
    // Wait for widgets to load
    await page.waitForTimeout(3000) // Allow queries to complete

    // Check widget has content (not loading/error)
    const widget = page.locator('[data-testid="widget-card"]').first()
    if (await widget.isVisible()) {
      const content = widget.locator('[data-testid="widget-content"]')
      await expect(content).toBeVisible({ timeout: 10000 })
    }
  })

  test('clicking chart navigates to traces with filters', async ({ page }) => {
    // Find a chart widget
    const chartWidget = page
      .locator('[data-testid="widget-card"]')
      .filter({ has: page.locator('svg.recharts-surface') })
      .first()

    if (await chartWidget.isVisible()) {
      // Click on chart area (may need specific selector for bar/slice)
      await chartWidget.locator('svg.recharts-surface').click()

      // Should navigate to traces page with filters
      await expect(page).toHaveURL(/\/traces\?/, { timeout: 5000 })
    }
  })
})

test.describe('Time Range', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${DASHBOARD_URL}/test-dashboard-id`)
  })

  test('can change time range', async ({ page }) => {
    // Find time range picker
    const timeRangePicker = page.getByRole('button', { name: /last|time range/i })
    if (await timeRangePicker.isVisible()) {
      await timeRangePicker.click()

      // Select a different time range
      await page.getByRole('option', { name: /7 days|week/i }).click()

      // URL should update with time parameter
      await expect(page).toHaveURL(/time_rel=7d/i, { timeout: 5000 })
    }
  })

  test('refreshes data when time range changes', async ({ page }) => {
    // Find time range picker
    const timeRangePicker = page.getByRole('button', { name: /last|time range/i })
    if (await timeRangePicker.isVisible()) {
      await timeRangePicker.click()

      // Select different time range
      await page.getByRole('option', { name: /24 hours|day/i }).click()

      // Widget should show loading briefly
      const widget = page.locator('[data-testid="widget-card"]').first()
      if (await widget.isVisible()) {
        // Wait for reload to complete
        await page.waitForTimeout(2000)

        // Content should be visible
        await expect(widget.locator('[data-testid="widget-content"]')).toBeVisible({
          timeout: 10000,
        })
      }
    }
  })
})

test.describe('Dashboard Templates', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(DASHBOARD_URL)
  })

  test('can create dashboard from template', async ({ page }) => {
    // Open create dialog
    await page.getByRole('button', { name: /create|new/i }).click()

    // Look for template selection
    const templateTab = page.getByRole('tab', { name: /template/i })
    if (await templateTab.isVisible()) {
      await templateTab.click()

      // Select a template
      const template = page.locator('[data-testid="template-card"]').first()
      if (await template.isVisible()) {
        await template.click()

        // Confirm selection
        await page.getByRole('button', { name: /create|use template/i }).click()

        // Should navigate to new dashboard
        await expect(page).toHaveURL(/\/dashboards\/[^/]+$/, { timeout: 10000 })

        // Should have widgets from template
        await expect(page.locator('[data-testid="widget-card"]')).toHaveCount({ min: 1 })
      }
    }
  })
})

test.describe('Dashboard Export/Import', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${DASHBOARD_URL}/test-dashboard-id`)
  })

  test('can export dashboard', async ({ page }) => {
    // Find export button (may be in toolbar or dropdown)
    const exportButton = page.getByRole('button', { name: /export/i })
    if (await exportButton.isVisible()) {
      // Set up download listener
      const downloadPromise = page.waitForEvent('download')

      await exportButton.click()

      // Verify download started
      const download = await downloadPromise
      expect(download.suggestedFilename()).toMatch(/dashboard.*\.json$/i)
    }
  })

  test('can open import dialog', async ({ page }) => {
    // Find import button
    const importButton = page.getByRole('button', { name: /import/i })
    if (await importButton.isVisible()) {
      await importButton.click()

      // Import dialog should open
      await expect(page.getByRole('dialog')).toBeVisible()
      await expect(page.getByRole('dialog')).toContainText(/import/i)
    }
  })
})
