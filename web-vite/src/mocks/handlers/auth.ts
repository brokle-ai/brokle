import { http, HttpResponse } from 'msw'

const MOCK_USER = {
  id: 'usr_test_0000000000000000000000',
  email: 'e2e-user@brokle.test',
  first_name: 'E2E',
  last_name: 'User',
  default_organization_id: 'org_test_0000000000000000000000',
  is_email_verified: true,
  created_at: new Date().toISOString(),
}

function authCookies(): Array<[string, string]> {
  return [
    ['Set-Cookie', 'access_token=mock-access; Path=/; HttpOnly; SameSite=Lax; Max-Age=900'],
    ['Set-Cookie', 'refresh_token=mock-refresh; Path=/; HttpOnly; SameSite=Lax; Max-Age=604800'],
    ['Set-Cookie', 'csrf_token=mock-csrf; Path=/; SameSite=Lax; Max-Age=900'],
  ]
}

export const authHandlers = [
  http.post('*/api/v1/auth/login', async ({ request }) => {
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
      { status: 200, headers: authCookies() },
    )
  }),

  http.post('*/api/v1/auth/signup', async ({ request }) => {
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
      { status: 201, headers: authCookies() },
    )
  }),

  http.post('*/api/v1/auth/logout', () => {
    return new HttpResponse(null, {
      status: 204,
      headers: [
        ['Set-Cookie', 'access_token=; Path=/; HttpOnly; Max-Age=0'],
        ['Set-Cookie', 'refresh_token=; Path=/; HttpOnly; Max-Age=0'],
        ['Set-Cookie', 'csrf_token=; Path=/; Max-Age=0'],
      ],
    })
  }),

  http.post('*/api/v1/auth/refresh', () =>
    HttpResponse.json(
      { expires_at: Date.now() + 15 * 60 * 1000, expires_in: 15 * 60 * 1000 },
      { status: 200, headers: authCookies() },
    ),
  ),

  http.post('*/api/v1/auth/forgot-password', () => new HttpResponse(null, { status: 204 })),

  http.get('*/api/v1/users/me', () => HttpResponse.json(MOCK_USER)),
]
