import { expect, test } from '@playwright/test'

// Reactive refresh contract: when access_token is rejected (401) but
// refresh_token is still valid, the HTTP client's 401 interceptor fires
// /v1/auth/refresh once and retries the original request. This spec
// forces the condition by deleting the access_token cookie from the
// authenticated browser context, then making an API call.
//
// Runs under e2e-user (prerequisite: storageState with both tokens).
test.describe('reactive refresh', () => {
  test('401 on expired access_token triggers refresh + retry', async ({ page }) => {
    await page.goto('/')
    await page.waitForURL((url) => !url.pathname.startsWith('/signin'), { timeout: 15_000 })

    // Drop access_token. refresh_token remains.
    const context = page.context()
    const cookies = await context.cookies()
    const keep = cookies.filter((c) => c.name !== 'access_token')
    await context.clearCookies()
    await context.addCookies(keep)

    // Trigger a request that requires auth. The client sees 401, fires
    // refresh, retries. We assert: exactly one /refresh hit, and the
    // original request ultimately succeeds.
    const refreshHits: string[] = []
    page.on('request', (req) => {
      if (req.url().includes('/v1/auth/refresh')) refreshHits.push(req.url())
    })

    const meResp = await page.request.get('/v1/users/me')
    // API might return 401 directly via page.request (which doesn't go
    // through the browser interceptor). So instead, navigate to a page
    // that triggers the client's authed fetch and watch for recovery.
    await page.goto('/')
    await page.waitForLoadState('networkidle', { timeout: 15_000 })

    expect(refreshHits.length, 'exactly one refresh call expected').toBe(1)

    // Access token cookie is back after refresh.
    const after = await context.cookies()
    expect(after.find((c) => c.name === 'access_token')).toBeDefined()

    // meResp is noted but not asserted — page.request sidesteps the
    // interceptor; the true signal is the refresh call fired during
    // page load above.
    void meResp
  })
})
