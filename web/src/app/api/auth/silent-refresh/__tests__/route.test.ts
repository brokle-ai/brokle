// Silent-refresh route handler tests. Covers the five outcomes the
// DAL depends on to avoid forced sign-outs on recoverable 401s:
//
//   1. no refresh_token cookie            → /signin?status=expired
//   2. backend refresh 200                → bare return URL +
//                                           forwarded Set-Cookie +
//                                           brokle_refresh_attempt
//                                           marker cookie (short
//                                           Max-Age)
//   3. backend refresh non-2xx            → /signin?status=expired
//                                           + forwarded clear-cookie
//   4. backend unreachable (fetch throws) → /signin?status=expired
//   5. open-redirect attempt via `return` → collapsed to "/"

import { NextRequest } from 'next/server'
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from 'vitest'

import { GET } from '../route'

function buildReq(url: string, cookies: Record<string, string>): NextRequest {
  const cookieHeader = Object.entries(cookies)
    .map(([k, v]) => `${k}=${v}`)
    .join('; ')
  if (!cookieHeader) {
    return new NextRequest(url)
  }
  return new NextRequest(url, { headers: { cookie: cookieHeader } })
}

describe('/api/auth/silent-refresh GET', () => {
  const originalFetch = global.fetch
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn()
    global.fetch = fetchMock as unknown as typeof fetch
  })

  afterEach(() => {
    global.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('redirects to /signin?status=expired when refresh_token cookie is absent', async () => {
    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=%2Fdashboard',
      {},
    )
    const res = await GET(req)
    expect(res.status).toBe(307)
    const loc = new URL(res.headers.get('location')!)
    expect(loc.pathname).toBe('/signin')
    expect(loc.searchParams.get('status')).toBe('expired')
    expect(loc.searchParams.get('redirect')).toBe('/dashboard')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('redirects to a clean return URL on refresh 200 and emits marker + rotated cookies', async () => {
    const setCookies = [
      'access_token=new-acc; Path=/; Max-Age=900; HttpOnly; SameSite=Lax',
      'refresh_token=new-ref; Path=/; Max-Age=604800; HttpOnly; SameSite=Strict',
      'csrf_token=new-csrf; Path=/; Max-Age=900; SameSite=Lax',
    ]
    const h = new Headers()
    for (const sc of setCookies) h.append('set-cookie', sc)
    fetchMock.mockResolvedValueOnce(
      new Response('{"expires_at":1,"expires_in":900000}', { status: 200, headers: h }),
    )

    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=%2Fdashboard%2Fprojects',
      { refresh_token: 'stale-ref' },
    )
    const res = await GET(req)

    expect(res.status).toBe(307)
    const loc = new URL(res.headers.get('location')!)
    expect(loc.pathname).toBe('/dashboard/projects')
    // No URL pollution — the loop guard rides a cookie instead.
    expect(loc.searchParams.get('_r')).toBeNull()
    expect(loc.search).toBe('')

    // Rotated session cookies + loop-guard marker all forwarded.
    const forwarded = res.headers.getSetCookie()
    expect(forwarded.some(s => s.startsWith('access_token=new-acc'))).toBe(true)
    expect(forwarded.some(s => s.startsWith('refresh_token=new-ref'))).toBe(true)
    expect(forwarded.some(s => s.startsWith('csrf_token=new-csrf'))).toBe(true)

    const marker = forwarded.find(s =>
      s.startsWith('brokle_refresh_attempt='),
    )
    expect(marker).toBeDefined()
    // Short-lived so a later 401 on the same URL isn't sticky-blocked.
    expect(marker).toMatch(/Max-Age=\d+/)
    const maxAge = Number(/Max-Age=(\d+)/.exec(marker!)![1])
    expect(maxAge).toBeGreaterThan(0)
    expect(maxAge).toBeLessThanOrEqual(30)
    expect(marker).toContain('Path=/')
    expect(marker!.toLowerCase()).toContain('httponly')

    // Backend was called with the inbound cookie jar forwarded.
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/api/v1/auth/refresh')
    expect(init.method).toBe('POST')
    expect((init.headers as Record<string, string>).cookie).toContain(
      'refresh_token=stale-ref',
    )
  })

  it('redirects to /signin?status=expired when backend refresh 401, forwarding clear cookies', async () => {
    const clearHeaders = new Headers()
    clearHeaders.append(
      'set-cookie',
      'access_token=; Path=/; Max-Age=-1; HttpOnly; SameSite=Lax',
    )
    clearHeaders.append(
      'set-cookie',
      'refresh_token=; Path=/; Max-Age=-1; HttpOnly; SameSite=Strict',
    )
    fetchMock.mockResolvedValueOnce(
      new Response('{"error":{"message":"expired"}}', {
        status: 401,
        headers: clearHeaders,
      }),
    )

    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=%2Fdashboard',
      { refresh_token: 'revoked' },
    )
    const res = await GET(req)

    expect(res.status).toBe(307)
    const loc = new URL(res.headers.get('location')!)
    expect(loc.pathname).toBe('/signin')
    expect(loc.searchParams.get('status')).toBe('expired')
    expect(loc.searchParams.get('redirect')).toBe('/dashboard')

    // Clear-cookie headers forwarded so the jar matches server state.
    const forwarded = res.headers.getSetCookie()
    expect(forwarded).toHaveLength(2)
    expect(forwarded.every(s => s.includes('Max-Age=-1'))).toBe(true)
  })

  it('redirects to /signin?status=expired when backend is unreachable', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('fetch failed'))

    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=%2Fdashboard',
      { refresh_token: 'ref' },
    )
    const res = await GET(req)

    expect(res.status).toBe(307)
    const loc = new URL(res.headers.get('location')!)
    expect(loc.pathname).toBe('/signin')
    expect(loc.searchParams.get('status')).toBe('expired')
    expect(loc.searchParams.get('redirect')).toBe('/dashboard')
  })

  it('collapses absolute or protocol-relative `return` to "/" (open-redirect guard)', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response('{}', { status: 200, headers: new Headers() }),
    )

    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=' +
        encodeURIComponent('https://evil.example/'),
      { refresh_token: 'ref' },
    )
    const res = await GET(req)
    const loc = new URL(res.headers.get('location')!)
    // The sanitiser collapsed `return` to "/", so the redirect URL's
    // path is "/" (no query pollution — marker rides a cookie).
    expect(loc.host).toBe('localhost:3000')
    expect(loc.pathname).toBe('/')
    expect(loc.search).toBe('')
  })

  it('rejects protocol-relative `return` values', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response('{}', { status: 200, headers: new Headers() }),
    )
    const req = buildReq(
      'http://localhost:3000/api/auth/silent-refresh?return=' +
        encodeURIComponent('//evil.example/dashboard'),
      { refresh_token: 'ref' },
    )
    const res = await GET(req)
    const loc = new URL(res.headers.get('location')!)
    expect(loc.host).toBe('localhost:3000')
    expect(loc.pathname).toBe('/')
  })
})
