import { expect, test } from '@playwright/test'

// New-account signup. Uses a per-run email so reruns don't collide with a
// prior seed. Does not assert on verification email delivery — that's the
// backend's job and would require MailHog or similar wired in; we just
// assert the signup endpoint accepted the request and the client routed
// correctly.
test.describe('signup', () => {
  test('creates a new account and routes to the verification flow', async ({ page }) => {
    const email = `e2e-signup-${Date.now()}@brokle.test`

    await page.goto('/signup')

    await page.fill('input[name="firstName"]', 'E2E')
    await page.fill('input[name="lastName"]', 'Signup')
    await page.fill('input[name="email"]', email)
    await page.fill('input[name="password"]', 'e2e-signup-password-123!')

    const [signupResp] = await Promise.all([
      page.waitForResponse(
        (resp) =>
          resp.url().includes('/v1/auth/signup') && resp.request().method() === 'POST',
        { timeout: 15_000 },
      ),
      page.click('button[type="submit"]'),
    ])

    expect(signupResp.ok(), `signup response: ${signupResp.status()}`).toBe(true)

    // Post-signup path is /verify-email or /onboarding depending on backend
    // config; we only pin that we left /signup.
    await page.waitForURL((url) => !url.pathname.startsWith('/signup'), { timeout: 15_000 })
  })
})
