import { http, HttpResponse } from 'msw'

// Backend contract (see internal/transport/http/handlers/auth/). Response
// bodies follow the Stripe/OpenAI style — raw resource on 2xx, no `{success,
// data, meta}` envelope (CLAUDE.md gotcha #23). Tokens live in httpOnly
// cookies; the body only carries `user` + expiry metadata.

const MOCK_USER = {
  id: 'usr_test_0000000000000000000000',
  email: 'e2e-user@brokle.test',
  first_name: 'E2E',
  last_name: 'User',
  default_organization_id: 'org_test_0000000000000000000000',
  is_email_verified: true,
  created_at: new Date().toISOString(),
}

function authCookies(): string[] {
  return [
    'access_token=mock-access; Path=/; HttpOnly; SameSite=Lax; Max-Age=900',
    'refresh_token=mock-refresh; Path=/; HttpOnly; SameSite=Lax; Max-Age=604800',
    // Non-httpOnly so the client can read it for double-submit.
    'csrf_token=mock-csrf; Path=/; SameSite=Lax; Max-Age=900',
  ]
}

export const authHandlers = [
  http.post('*/v1/auth/login', async ({ request }) => {
    const body = (await request.json()) as { email?: string; password?: string }
    if (!body.email || !body.password) {
      return HttpResponse.json(
        { error: { type: 'validation', message: 'email and password required' } },
        { status: 422 },
      )
    }
    return HttpResponse.json(
      {
        user: MOCK_USER,
        expires_at: Date.now() + 15 * 60 * 1000,
        expires_in: 15 * 60 * 1000,
      },
      {
        status: 200,
        headers: authCookies().map((c) => ['Set-Cookie', c] as [string, string]),
      },
    )
  }),

  http.post('*/v1/auth/signup', async ({ request }) => {
    const body = (await request.json()) as { email?: string; password?: string }
    if (!body.email || !body.password) {
      return HttpResponse.json(
        { error: { type: 'validation', message: 'email and password required' } },
        { status: 422 },
      )
    }
    return HttpResponse.json(
      {
        user: { ...MOCK_USER, email: body.email, is_email_verified: false },
        expires_at: Date.now() + 15 * 60 * 1000,
        expires_in: 15 * 60 * 1000,
      },
      {
        status: 201,
        headers: authCookies().map((c) => ['Set-Cookie', c] as [string, string]),
      },
    )
  }),

  http.post('*/v1/auth/logout', () => {
    return new HttpResponse(null, {
      status: 204,
      headers: [
        ['Set-Cookie', 'access_token=; Path=/; HttpOnly; Max-Age=0'],
        ['Set-Cookie', 'refresh_token=; Path=/; HttpOnly; Max-Age=0'],
        ['Set-Cookie', 'csrf_token=; Path=/; Max-Age=0'],
      ],
    })
  }),

  http.post('*/v1/auth/refresh', () => {
    return HttpResponse.json(
      { expires_at: Date.now() + 15 * 60 * 1000, expires_in: 15 * 60 * 1000 },
      {
        status: 200,
        headers: authCookies().map((c) => ['Set-Cookie', c] as [string, string]),
      },
    )
  }),

  http.post('*/v1/auth/forgot-password', () => {
    // Anti-enumeration: backend returns 204 for any email.
    return new HttpResponse(null, { status: 204 })
  }),

  http.get('*/v1/users/me', () => {
    return HttpResponse.json(MOCK_USER)
  }),
]
