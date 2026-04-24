import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Sanity tests for the openapi-fetch middleware chain. Focus on the
// invariants that are easy to regress silently:
//   - CSRF attaches on mutation methods, NOT on GET
//   - non-2xx responses throw a typed BrokleError subclass
//   - 401 triggers exactly one refresh + one retry attempt

function setCsrfCookie(value: string | null) {
  if (value === null) {
    Object.defineProperty(document, 'cookie', { value: '', configurable: true, writable: true })
    return
  }
  Object.defineProperty(document, 'cookie', {
    value: `csrf_token=${value}`,
    configurable: true,
    writable: true,
  })
}

describe('api client middleware chain', () => {
  let fetchSpy: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.resetModules()
    setCsrfCookie('csrf-abc')
    fetchSpy = vi.fn()
    vi.stubGlobal('fetch', fetchSpy)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    setCsrfCookie(null)
  })

  it('attaches X-CSRF-Token on POST only', async () => {
    fetchSpy.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))
    const { rawFetch } = await import('../client')

    await rawFetch('/v1/resource', { method: 'POST' })
    const postCall = fetchSpy.mock.calls.at(-1)!
    const postHeaders = (postCall[1] as RequestInit).headers as Headers
    expect(postHeaders.get('X-CSRF-Token')).toBe('csrf-abc')

    await rawFetch('/v1/resource', { method: 'GET' })
    const getCall = fetchSpy.mock.calls.at(-1)!
    const getHeaders = (getCall[1] as RequestInit).headers as Headers
    expect(getHeaders.get('X-CSRF-Token')).toBeNull()
  })

  it('throws ValidationError on 422', async () => {
    fetchSpy.mockResolvedValue(
      new Response(JSON.stringify({ error: { type: 'validation', message: 'nope' } }), {
        status: 422,
        headers: { 'content-type': 'application/json' },
      }),
    )
    const { rawFetch } = await import('../client')
    const { ValidationError } = await import('../errors')
    await expect(rawFetch('/v1/resource')).rejects.toBeInstanceOf(ValidationError)
  })

  it('on 401, calls refresh once and retries the original request', async () => {
    fetchSpy
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(new Response(null, { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }))

    const { rawFetch } = await import('../client')
    const resp = await rawFetch('/v1/resource', { method: 'GET' })
    expect(resp.status).toBe(200)
    // First call = original (401). Second = refresh. Third = retry.
    expect(fetchSpy).toHaveBeenCalledTimes(3)
    expect(fetchSpy.mock.calls[1]![0]).toMatch(/\/v1\/auth\/refresh$/)
  })
})
