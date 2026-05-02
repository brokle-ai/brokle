import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Reset module state between tests — the refreshWithLock module caches
// an in-flight Promise at module scope. `vi.resetModules()` clears it.

describe('refreshWithLock — single-flight invariant', () => {
  let fetchSpy: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.resetModules()
    fetchSpy = vi.fn().mockResolvedValue(new Response(null, { status: 200 }))
    vi.stubGlobal('fetch', fetchSpy)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('fires exactly ONE refresh request for N concurrent callers', async () => {
    const { refreshWithLock } = await import('../auth-refresh')
    const N = 10
    const results = await Promise.allSettled(
      Array.from({ length: N }, () => refreshWithLock()),
    )
    expect(results.every((r) => r.status === 'fulfilled')).toBe(true)
    expect(fetchSpy).toHaveBeenCalledTimes(1)
    expect(fetchSpy.mock.calls[0]![0]).toMatch(/\/v1\/auth\/refresh$/)
    expect(fetchSpy.mock.calls[0]![1]).toMatchObject({ method: 'POST', credentials: 'include' })
  })

  it('allows subsequent refreshes after the first completes', async () => {
    const { refreshWithLock } = await import('../auth-refresh')
    await refreshWithLock()
    await refreshWithLock()
    expect(fetchSpy).toHaveBeenCalledTimes(2)
  })

  it('rejects all concurrent callers when refresh fails', async () => {
    fetchSpy.mockResolvedValueOnce(new Response(null, { status: 401 }))
    const { refreshWithLock } = await import('../auth-refresh')
    const results = await Promise.allSettled([refreshWithLock(), refreshWithLock()])
    expect(results.every((r) => r.status === 'rejected')).toBe(true)
    expect(fetchSpy).toHaveBeenCalledTimes(1)
  })
})
