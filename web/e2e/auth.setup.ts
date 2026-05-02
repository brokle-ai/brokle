import { test as setup, expect, type APIRequestContext, type Page } from '@playwright/test'
import { mkdirSync } from 'node:fs'
import { dirname } from 'node:path'

const USER_FILE = 'playwright/.auth/user.json'
const ADMIN_FILE = 'playwright/.auth/admin.json'

mkdirSync(dirname(USER_FILE), { recursive: true })

// Credentials drive auth-setup. Seed these via the backend before running
// E2E (see docs). Any spec that runs under the e2e-user / e2e-admin
// projects depends on this setup step succeeding — if the env is unset
// the step fails loudly rather than silently skipping, because a green
// suite with skipped setup is a false signal.
const USER_EMAIL = process.env.E2E_USER_EMAIL ?? 'e2e-user@brokle.test'
const USER_PASSWORD = process.env.E2E_USER_PASSWORD ?? 'e2e-user-password-123!'
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'e2e-admin@brokle.test'
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'e2e-admin-password-123!'

async function ensureAccount(
  request: APIRequestContext,
  email: string,
  password: string,
): Promise<void> {
  // Try login first (idempotent path for repeated CI runs).
  const loginResp = await request.post('/api/v1/auth/login', {
    data: { email, password },
  })
  if (loginResp.ok()) return

  // 401 / 404 from login — bootstrap via signup so fresh environments
  // self-seed. Anything else is a real failure.
  if (loginResp.status() !== 401 && loginResp.status() !== 404) {
    throw new Error(
      `Unexpected login failure for ${email}: ${loginResp.status()} ${await loginResp.text()}`,
    )
  }

  const signupResp = await request.post('/api/v1/auth/signup', {
    data: {
      email,
      password,
      first_name: 'E2E',
      last_name: email.split('@')[0],
    },
  })
  if (!signupResp.ok()) {
    throw new Error(
      `Signup failed for ${email}: ${signupResp.status()} ${await signupResp.text()}`,
    )
  }
}

async function captureStorageState(
  request: APIRequestContext,
  page: Page,
  email: string,
  password: string,
  outPath: string,
): Promise<void> {
  await ensureAccount(request, email, password)

  // Fresh login confirms cookies land on the right origin; visit `/` once
  // so the CSRF cookie is written + the browser context adopts the
  // response Set-Cookie headers before we persist.
  const loginResp = await request.post('/api/v1/auth/login', {
    data: { email, password },
  })
  expect(loginResp.ok(), `login for ${email} must succeed`).toBeTruthy()

  await page.goto('/')
  await page.context().storageState({ path: outPath })
}

setup('authenticate user role', async ({ request, page }) => {
  await captureStorageState(request, page, USER_EMAIL, USER_PASSWORD, USER_FILE)
})

setup('authenticate admin role', async ({ request, page }) => {
  await captureStorageState(request, page, ADMIN_EMAIL, ADMIN_PASSWORD, ADMIN_FILE)
})
