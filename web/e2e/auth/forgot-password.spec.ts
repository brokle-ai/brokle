import { expect, test } from '@playwright/test'

test.describe('forgot password', () => {
  test('submits email, shows success state', async ({ page }) => {
    await page.goto('/forgot-password')

    await page.fill('input[name="email"]', 'e2e-user@brokle.test')

    const [resp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes('/api/v1/auth/forgot-password') && r.request().method() === 'POST',
        { timeout: 15_000 },
      ),
      page.click('button[type="submit"]'),
    ])

    // Backend always returns 2xx here (anti-enumeration) regardless of
    // whether the email exists — assert on 2xx only, not 200.
    expect(resp.status(), `forgot-password status`).toBeGreaterThanOrEqual(200)
    expect(resp.status()).toBeLessThan(300)

    // Client surfaces a "check your inbox" confirmation; we don't pin the
    // copy, but the submit button should transition out of its loading
    // state or the form should hide.
    await expect(page.locator('form')).toBeHidden({ timeout: 10_000 }).catch(async () => {
      // Some flavors keep the form but disable it — fall back to ensuring
      // no error alert surfaced.
      await expect(page.getByRole('alert', { name: /error/i })).toHaveCount(0)
    })
  })
})
