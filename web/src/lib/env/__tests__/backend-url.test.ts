// Regression tests for the server-side backend URL resolver.
//
// This function is the single source of truth for every server-side
// call from the Next.js dashboard process to the Go backend (DAL,
// proxy silent-refresh, route handlers). A bug here breaks every
// protected page load in deployments that configure only the
// documented NEXT_PUBLIC_API_URL variable — which is the scenario
// the review flagged and this suite locks.

import { afterEach, describe, expect, it, vi } from 'vitest'

import { getBackendBaseURL } from '../backend-url'

describe('getBackendBaseURL', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('returns BROKLE_API_PROXY_TARGET when both vars are set (server-only override wins)', () => {
    vi.stubEnv('BROKLE_API_PROXY_TARGET', 'http://backend.internal:8080')
    vi.stubEnv('NEXT_PUBLIC_API_URL', 'https://api.brokle.com')
    expect(getBackendBaseURL()).toBe('http://backend.internal:8080')
  })

  // The reviewer's P1 regression: only the documented public URL is
  // set. Before the fix this returned localhost:8080 and every
  // protected dashboard page 500'd / redirect-looped. Lock it.
  it('falls back to NEXT_PUBLIC_API_URL when BROKLE_API_PROXY_TARGET is unset', () => {
    vi.stubEnv('BROKLE_API_PROXY_TARGET', '')
    vi.stubEnv('NEXT_PUBLIC_API_URL', 'https://api.brokle.com')
    expect(getBackendBaseURL()).toBe('https://api.brokle.com')
  })

  it('falls back to localhost:8080 when neither var is set', () => {
    vi.stubEnv('BROKLE_API_PROXY_TARGET', '')
    vi.stubEnv('NEXT_PUBLIC_API_URL', '')
    expect(getBackendBaseURL()).toBe('http://localhost:8080')
  })

  // The web Dockerfile bakes `ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}`
  // — an absent build ARG materialises as `""` at runtime, not
  // `undefined`. The `||` chain must treat empty strings as "unset"
  // for this path to work.
  it('treats empty-string BROKLE_API_PROXY_TARGET as unset (Docker ARG default)', () => {
    vi.stubEnv('BROKLE_API_PROXY_TARGET', '')
    vi.stubEnv('NEXT_PUBLIC_API_URL', 'https://api.brokle.com')
    expect(getBackendBaseURL()).toBe('https://api.brokle.com')
  })

  it('treats empty-string NEXT_PUBLIC_API_URL as unset', () => {
    vi.stubEnv('BROKLE_API_PROXY_TARGET', 'http://backend.internal:8080')
    vi.stubEnv('NEXT_PUBLIC_API_URL', '')
    expect(getBackendBaseURL()).toBe('http://backend.internal:8080')
  })
})
