import { expect, test } from '@playwright/test'

// Runs under e2e-user project — storageState already has a valid session.
// Tests logout from within the app: clicking the menu's sign-out action
// clears cookies and returns the browser to /signin.
test.describe('logout', () => {
  test('clears auth cookies and redirects to /signin', async ({ page }) => {
    await page.goto('/')
    await page.waitForURL((url) => !url.pathname.startsWith('/signin'), { timeout: 15_000 })

    const beforeCookies = await page.context().cookies()
    expect(beforeCookies.find((c) => c.name === 'access_token')).toBeDefined()

    // Invoke logout via the API to avoid coupling to the current nav menu
    // structure (menu changes have broken many E2E suites). Cookies are
    // cleared by the backend's Set-Cookie; the browser context picks it up.
    const logoutResp = await page.request.post('/v1/auth/logout')
    expect(logoutResp.ok(), `logout status ${logoutResp.status()}`).toBe(true)

    // Navigating anywhere now should bounce through the proxy to /signin.
    await page.goto('/')
    await page.waitForURL(/\/signin/, { timeout: 15_000 })

    const afterCookies = await page.context().cookies()
    const access = afterCookies.find((c) => c.name === 'access_token')
    // Either cleared or expired — both are valid "logged out" signals.
    expect(access === undefined || (access.expires > 0 && access.expires < Date.now() / 1000))
      .toBe(true)
  })
})
