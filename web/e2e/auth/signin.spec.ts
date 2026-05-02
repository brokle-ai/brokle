import { expect, test } from '@playwright/test'

const USER_EMAIL = process.env.E2E_USER_EMAIL ?? 'e2e-user@brokle.test'
const USER_PASSWORD = process.env.E2E_USER_PASSWORD ?? 'e2e-user-password-123!'

test.describe('signin', () => {
  test('submits valid credentials, lands on dashboard, cookies set', async ({ page }) => {
    await page.goto('/signin')

    await page.fill('input[name="email"]', USER_EMAIL)
    await page.fill('input[name="password"]', USER_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL((url) => !url.pathname.startsWith('/signin'), { timeout: 15_000 })

    const cookies = await page.context().cookies()
    const names = cookies.map((c) => c.name)
    expect(names).toContain('access_token')
    expect(names).toContain('refresh_token')
    expect(names).toContain('csrf_token')

    const accessCookie = cookies.find((c) => c.name === 'access_token')
    expect(accessCookie?.httpOnly, 'access_token must be httpOnly').toBe(true)

    const csrfCookie = cookies.find((c) => c.name === 'csrf_token')
    expect(csrfCookie?.httpOnly, 'csrf_token must NOT be httpOnly (double-submit pattern)').toBe(false)
  })

  test('rejects invalid credentials, stays on signin with error', async ({ page }) => {
    await page.goto('/signin')

    await page.fill('input[name="email"]', 'nobody@brokle.test')
    await page.fill('input[name="password"]', 'wrong-password-000!')
    await page.click('button[type="submit"]')

    // Backend returns 401; client stays on /signin and surfaces an error. We
    // don't pin the exact copy — just assert we didn't navigate away.
    await page.waitForTimeout(1_000)
    expect(new URL(page.url()).pathname).toMatch(/\/signin/)
    const cookies = await page.context().cookies()
    expect(cookies.find((c) => c.name === 'access_token')).toBeUndefined()
  })
})
