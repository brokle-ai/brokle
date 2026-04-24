// DAL tests — lock the /users/me 401 recovery policy:
//
//   1. happy path 200                                  → returns mapped profile
//   2. 401 + refresh cookie + first attempt            → redirect to silent-refresh
//   3. 401 + no refresh cookie                         → redirect to /signin?status=expired
//   4. 401 + brokle_refresh_attempt cookie (loop guard)→ redirect to /signin?status=expired
//
// The DAL module is marked `'server-only'`, which throws at import
// in a non-server context. We stub server-only first, then mock the
// three Next.js runtime surfaces the DAL reads (next/headers,
// next/navigation) and the fetchBackend helper.

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

// Mutable stubs per test. Set inside beforeEach / individual tests.
const mockCookieStore = {
  has: vi.fn<(name: string) => boolean>(),
}
const mockHeaderStore = {
  get: vi.fn<(name: string) => string | null>(),
}

vi.mock('next/headers', () => ({
  cookies: vi.fn(async () => mockCookieStore),
  headers: vi.fn(async () => mockHeaderStore),
}))

// redirect() in Next.js throws a special signal. Reproduce that by
// throwing a tagged error so tests can assert on both the fact of
// redirect AND the destination URL.
class RedirectSignal extends Error {
  constructor(public readonly destination: string) {
    super(`redirect:${destination}`)
    this.name = 'RedirectSignal'
  }
}
vi.mock('next/navigation', () => ({
  redirect: vi.fn((url: string) => {
    throw new RedirectSignal(url)
  }),
}))

const fetchBackendMock =
  vi.fn<(path: string, init?: RequestInit) => Promise<Response>>()
vi.mock('../server-fetch', () => ({
  fetchBackend: (...args: Parameters<typeof fetchBackendMock>) =>
    fetchBackendMock(...args),
}))

// Minimal /users/me response payload — whatever the profile mapper
// accepts for a valid session user. The mapper is imported and run
// as-is, so the shape has to parse cleanly.
const okProfile = {
  id: 'user-1',
  email: 'a@b.c',
  first_name: 'A',
  last_name: 'B',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  is_email_verified: true,
  default_organization_id: 'org-1',
  organizations: [],
}

async function importDal() {
  const mod = await import('../dal')
  return mod
}

describe('dal.getCurrentUser', () => {
  beforeEach(() => {
    vi.resetModules()
    mockCookieStore.has.mockReset()
    mockHeaderStore.get.mockReset()
    fetchBackendMock.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('returns mapped profile on happy-path 200', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) => n === 'access_token',
    )
    mockHeaderStore.get.mockReturnValue('/dashboard')
    fetchBackendMock.mockResolvedValueOnce(
      new Response(JSON.stringify(okProfile), { status: 200 }),
    )

    const { getCurrentUser } = await importDal()
    const profile = await getCurrentUser()
    expect(profile.user.email).toBe('a@b.c')
  })

  it('on 401 + refresh cookie present + first attempt, redirects to silent-refresh', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) => n === 'access_token' || n === 'refresh_token',
    )
    mockHeaderStore.get.mockReturnValue('/dashboard/projects')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    await expect(getCurrentUser()).rejects.toMatchObject({
      name: 'RedirectSignal',
      destination:
        '/api/auth/silent-refresh?return=%2Fdashboard%2Fprojects',
    })
  })

  it('preserves the query string when redirecting to silent-refresh', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) => n === 'access_token' || n === 'refresh_token',
    )
    mockHeaderStore.get.mockReturnValue('/projects/abc?tab=traces')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    await expect(getCurrentUser()).rejects.toMatchObject({
      destination:
        '/api/auth/silent-refresh?return=%2Fprojects%2Fabc%3Ftab%3Dtraces',
    })
  })

  it('on 401 + no refresh cookie, redirects to /signin?status=expired', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) => n === 'access_token',
    )
    mockHeaderStore.get.mockReturnValue('/dashboard')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    let caught: unknown
    try {
      await getCurrentUser()
    } catch (err) {
      caught = err
    }
    expect(caught).toBeInstanceOf(RedirectSignal)
    const destination = (caught as RedirectSignal).destination
    expect(destination).toContain('/signin')
    expect(destination).toContain('status=expired')
    expect(destination).not.toContain('/api/auth/silent-refresh')
  })

  it('on 401 while brokle_refresh_attempt cookie is present, loop-guards to /signin', async () => {
    // Marker cookie set by the route handler on the just-completed
    // refresh. Its presence means a fresh refresh already happened
    // and /users/me is STILL 401 — a second refresh can't help.
    mockCookieStore.has.mockImplementation(
      (n: string) =>
        n === 'access_token' ||
        n === 'refresh_token' ||
        n === 'brokle_refresh_attempt',
    )
    mockHeaderStore.get.mockReturnValue('/dashboard')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    let caught: unknown
    try {
      await getCurrentUser()
    } catch (err) {
      caught = err
    }
    expect(caught).toBeInstanceOf(RedirectSignal)
    const destination = (caught as RedirectSignal).destination
    // Must terminate in /signin, NOT re-enter the refresh loop.
    expect(destination).toContain('/signin')
    expect(destination).toContain('status=expired')
    expect(destination).not.toContain('/api/auth/silent-refresh')
  })

  // The regression the reviewer flagged: once the marker has expired
  // (cookie Max-Age elapsed), a LATER 401 on the same URL must NOT
  // be blocked by a stale loop guard. The cookie's absence is the
  // proof that the earlier refresh window closed.
  it('on 401 after marker cookie expired, re-enters silent-refresh', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) =>
        // Marker cookie missing — Max-Age elapsed and browser
        // evicted it. Proceed with a fresh refresh attempt.
        n === 'access_token' || n === 'refresh_token',
    )
    mockHeaderStore.get.mockReturnValue('/dashboard')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    let caught: unknown
    try {
      await getCurrentUser()
    } catch (err) {
      caught = err
    }
    expect(caught).toBeInstanceOf(RedirectSignal)
    const destination = (caught as RedirectSignal).destination
    expect(destination).toContain('/api/auth/silent-refresh')
    expect(destination).toContain('return=')
  })

  it('falls back to "/" when x-brokle-pathname header is absent', async () => {
    mockCookieStore.has.mockImplementation(
      (n: string) => n === 'access_token' || n === 'refresh_token',
    )
    mockHeaderStore.get.mockReturnValue(null)
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{}', { status: 401 }),
    )

    const { getCurrentUser } = await importDal()
    await expect(getCurrentUser()).rejects.toMatchObject({
      destination: '/api/auth/silent-refresh?return=%2F',
    })
  })

  it('throws on non-401 backend error', async () => {
    mockCookieStore.has.mockReturnValue(true)
    mockHeaderStore.get.mockReturnValue('/dashboard')
    fetchBackendMock.mockResolvedValueOnce(
      new Response('{"error":{}}', { status: 500, statusText: 'ugh' }),
    )

    const { getCurrentUser } = await importDal()
    await expect(getCurrentUser()).rejects.toThrow(
      /backend returned 500/,
    )
  })
})
