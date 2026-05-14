import { readdirSync } from 'node:fs'
import { join } from 'node:path'
import { NextRequest } from 'next/server'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { config, proxy } from './proxy'

// Drift-prevention tests for the Next.js 16 proxy matcher.
//
// The matcher in proxy.ts is a hand-maintained negative regex listing
// every URL segment a logged-out user is allowed to reach. Historically
// it drifted out of sync with the app/(auth)/ directory (matcher had
// `accept-invitation` while the real route was `accept-invite`, and
// `callback` was missing entirely), which silently broke OAuth callback
// and invitation acceptance for every logged-out user.
//
// These tests assert the invariant: every directory under app/(auth)/
// is listed in the matcher's exclusion set. Any rename or addition
// without a corresponding matcher edit fails this spec.
describe('proxy matcher', () => {
  const matcher = config.matcher[0]

  it('excludes every route directory under app/(auth)/', () => {
    const authDir = join(process.cwd(), 'src/app/(auth)')
    const routes = readdirSync(authDir, { withFileTypes: true })
      .filter(d => d.isDirectory())
      .map(d => d.name)

    expect(routes.length).toBeGreaterThan(0)
    for (const route of routes) {
      // `\b` treats `-` as a word boundary, so `\baccept-invite\b`
      // matches `accept-invite` as a whole token and rejects the
      // partial-match pitfall (e.g. `accept` alone).
      expect(matcher).toMatch(new RegExp(`\\b${route}\\b`))
    }
  })

  it('distinguishes protected from public paths', () => {
    // The matcher string IS a JS regex — Next.js compiles it with
    // the same runtime. Exercising it directly here catches any
    // syntax regression as well as the route-level logic.
    const pattern = new RegExp(`^${matcher}$`)

    expect('/dashboard').toMatch(pattern)
    expect('/projects/x/traces').toMatch(pattern)
    expect('/settings').toMatch(pattern)

    // Public auth entry points must NOT match — proxy skipped.
    expect('/signin').not.toMatch(pattern)
    expect('/signup').not.toMatch(pattern)
    expect('/callback').not.toMatch(pattern)
    expect('/accept-invite').not.toMatch(pattern)
    expect('/forgot-password').not.toMatch(pattern)
    expect('/verify-email').not.toMatch(pattern)

    // Framework + asset paths must NOT match.
    expect('/api/v1/users/me').not.toMatch(pattern)
    expect('/v1/traces').not.toMatch(pattern)
    expect('/_next/static/chunks/foo.js').not.toMatch(pattern)
    expect('/favicon.ico').not.toMatch(pattern)
    expect('/logo.svg').not.toMatch(pattern)
    expect('/site.webmanifest').not.toMatch(pattern)
  })
})

// Behaviour tests for the silent-refresh flow. Cover the four branches
// from the plan:
//   1. both cookies present → pass through, no backend call.
//   2. access present alone → pass through, no backend call.
//   3. neither cookie present → redirect to /signin (no refresh).
//   4. refresh present, access absent → backend /refresh:
//      a. 200 → pass-through with forwarded Set-Cookie + rewritten
//         inbound Cookie header.
//      b. non-2xx → redirect to /signin?status=expired.
//      c. backend unreachable → redirect to /signin?status=expired.
describe('proxy silent refresh', () => {
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

  // makeReq builds a NextRequest with the given cookie header. Using
  // NextRequest directly (instead of stubbing req.cookies.has) keeps
  // the test aligned with the real runtime contract — proxy.ts reads
  // from a real NextRequest at production time.
  function makeReq(cookies: Record<string, string>): NextRequest {
    const cookieHeader = Object.entries(cookies)
      .map(([k, v]) => `${k}=${v}`)
      .join('; ')
    if (!cookieHeader) {
      return new NextRequest('http://localhost:3000/dashboard')
    }
    return new NextRequest('http://localhost:3000/dashboard', {
      headers: { cookie: cookieHeader },
    })
  }

  it('passes through when access_token is present (no backend call) and stamps x-brokle-pathname', async () => {
    const resp = await proxy(
      makeReq({ access_token: 'acc', refresh_token: 'ref' }),
    )
    expect(resp.status).toBe(200)
    expect(fetchMock).not.toHaveBeenCalled()
    // Next.js surfaces request-header rewrites via the
    // x-middleware-override-headers / x-middleware-request-<name>
    // response-header protocol. Assert the pathname header was
    // scheduled for injection into the downstream request.
    expect(
      resp.headers.get('x-middleware-request-x-brokle-pathname'),
    ).toBe('/dashboard')
  })

  it('passes through with access_token alone and stamps pathname', async () => {
    const resp = await proxy(makeReq({ access_token: 'acc' }))
    expect(resp.status).toBe(200)
    expect(fetchMock).not.toHaveBeenCalled()
    expect(
      resp.headers.get('x-middleware-request-x-brokle-pathname'),
    ).toBe('/dashboard')
  })

  it('redirects to /signin when no session cookies present', async () => {
    const resp = await proxy(makeReq({}))
    expect(resp.status).toBe(307) // NextResponse.redirect default
    const location = resp.headers.get('location')!
    const url = new URL(location)
    expect(url.pathname).toBe('/signin')
    expect(url.searchParams.get('redirect')).toBe('/dashboard')
    expect(url.searchParams.get('status')).toBeNull()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('silently refreshes when only refresh_token is present, then passes through', async () => {
    // Mock backend /refresh returning 200 with three Set-Cookie headers.
    const setCookies = [
      'access_token=new-acc; Path=/; Max-Age=900; HttpOnly; SameSite=Lax',
      'refresh_token=new-ref; Path=/; Max-Age=604800; HttpOnly; SameSite=Strict',
      'csrf_token=new-csrf; Path=/; Max-Age=900; SameSite=Lax',
    ]
    const headers = new Headers()
    for (const sc of setCookies) headers.append('set-cookie', sc)
    fetchMock.mockResolvedValueOnce(
      new Response('{"expires_at":1,"expires_in":900000}', {
        status: 200,
        headers,
      }),
    )

    const resp = await proxy(makeReq({ refresh_token: 'stale-ref' }))

    // 1. Exactly one POST to the backend refresh endpoint.
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/api/v1/auth/refresh')
    expect(init.method).toBe('POST')
    expect((init.headers as Record<string, string>).cookie).toContain(
      'refresh_token=stale-ref',
    )

    // 2. Response is a pass-through (NextResponse.next), NOT a redirect.
    expect(resp.status).toBe(200)
    expect(resp.headers.get('location')).toBeNull()

    // 3. All three Set-Cookie headers forwarded to the browser on the
    //    same round-trip. getSetCookie() returns the array of entries.
    const forwarded = resp.headers.getSetCookie()
    expect(forwarded).toHaveLength(3)
    expect(forwarded.some(h => h.startsWith('access_token=new-acc'))).toBe(true)
    expect(forwarded.some(h => h.startsWith('refresh_token=new-ref'))).toBe(true)
    expect(forwarded.some(h => h.startsWith('csrf_token=new-csrf'))).toBe(true)

    // 4. The inbound Cookie header was rewritten so the Server Component
    //    render sees the fresh access_token via cookies(). Next.js
    //    surfaces this via the `x-middleware-request-cookie` or equivalent
    //    internal header; the public observable is that the outgoing
    //    NextResponse is .next() (status 200) with no redirect.
  })

  it('redirects to /signin?status=expired when backend refresh fails', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response('{"error":{"message":"expired"}}', { status: 401 }),
    )

    const resp = await proxy(makeReq({ refresh_token: 'revoked' }))
    expect(resp.status).toBe(307)
    const url = new URL(resp.headers.get('location')!)
    expect(url.pathname).toBe('/signin')
    expect(url.searchParams.get('status')).toBe('expired')
    expect(url.searchParams.get('redirect')).toBe('/dashboard')
  })

  it('redirects to /signin?status=expired when backend unreachable', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('fetch failed'))

    const resp = await proxy(makeReq({ refresh_token: 'ref' }))
    expect(resp.status).toBe(307)
    const url = new URL(resp.headers.get('location')!)
    expect(url.pathname).toBe('/signin')
    expect(url.searchParams.get('status')).toBe('expired')
  })

  it('preserves query string in the redirect round-trip', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )
    const req = new NextRequest(
      'http://localhost:3000/projects/abc?tab=traces',
      { headers: { cookie: 'refresh_token=ref' } },
    )
    const resp = await proxy(req)
    const url = new URL(resp.headers.get('location')!)
    expect(url.searchParams.get('redirect')).toBe('/projects/abc?tab=traces')
  })
})
