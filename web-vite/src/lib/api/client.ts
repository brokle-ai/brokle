import createClient, { type Middleware } from 'openapi-fetch'
import { getRuntimeConfig } from '@/lib/config'
import { refreshWithLock } from './auth-refresh'
import { MUTATION_METHODS, readCsrfCookie } from './csrf'
import { throwTypedError } from './errors'
import type { paths as DashboardPaths } from './generated/dashboard'
import type { paths as SdkPaths } from './generated/sdk'

// One-shot retry marker: the first 401 triggers a refresh; if the
// replayed request ALSO gets 401, we surrender instead of looping.
const RETRY_HEADER = 'X-Brokle-Retried'

// Paths whose 401 means "the operation failed" (bad credentials,
// invalid OAuth state, expired reset token, missing refresh cookie),
// NOT "your session expired." These endpoints establish or manage the
// session — they have no session to refresh — so the auth-retry-via-
// refresh path must be skipped, otherwise:
//   - Login 401 (wrong password) silently fires a /auth/refresh, and
//     the original "Invalid email or password" message gets replaced
//     by a synthesized "session expired."
//   - Refresh 401 would recurse via the same code path until the
//     RETRY_HEADER short-circuits — wasteful even when bounded.
//   - Logout 401 (already logged out) triggers a pointless refresh.
//
// Mirrors SigNoz's exclusion list at
// `competitors/signoz/frontend/src/api/index.ts:111-120` and the
// `shouldRefresh(error)` URL-predicate documented in
// `axios-auth-refresh` (~2M weekly DLs). 401 is industry-standard
// for "wrong credentials" (RFC 9110 §15.5.2, OWASP, GitHub, Auth0,
// Stripe); the backend is correct, the client must distinguish.
function isAuthBoundaryPath(url: string): boolean {
  try {
    const pathname = url.startsWith('http')
      ? new URL(url).pathname
      : (url.split('?')[0] ?? '')
    return pathname.startsWith('/api/v1/auth/')
  } catch {
    return false
  }
}

const csrfMiddleware: Middleware = {
  async onRequest({ request }) {
    if (MUTATION_METHODS.has(request.method.toUpperCase())) {
      const csrf = readCsrfCookie()
      if (csrf) request.headers.set('X-CSRF-Token', csrf)
    }
    return request
  },
}

const authRetryMiddleware: Middleware = {
  async onResponse({ request, response }) {
    if (response.status !== 401) return response
    if (request.headers.get(RETRY_HEADER)) return response
    if (isAuthBoundaryPath(request.url)) return response

    try {
      await refreshWithLock()
    } catch {
      // Refresh failed; session is done. Return the original 401 so the
      // error envelope middleware still throws AuthenticationError and
      // the auth store reacts.
      return response
    }

    const retry = new Request(request, {
      headers: new Headers(request.headers),
    })
    retry.headers.set(RETRY_HEADER, '1')
    return fetch(retry)
  },
}

const errorEnvelopeMiddleware: Middleware = {
  async onResponse({ response }) {
    if (response.ok) return response
    await throwTypedError(response)
    // throwTypedError always throws; this return is unreachable but
    // keeps the middleware's type signature honest.
    return response
  },
}

function makeClient<Paths extends object>() {
  const client = createClient<Paths>({
    baseUrl: getRuntimeConfig().API_URL,
    credentials: 'include',
  })
  client.use(csrfMiddleware, authRetryMiddleware, errorEnvelopeMiddleware)
  return client
}

// Two separate clients, one per OpenAPI surface. Route-level code
// picks the right one based on whether the request goes through the
// dashboard plane (cookie + JWT) or the SDK plane (X-API-Key). Most
// dashboard code uses `api`; the SDK client is only for features that
// exercise the SDK plane directly (e.g. key-validation test flows).
export const api = makeClient<DashboardPaths>()
export const sdk = makeClient<SdkPaths>()

// Re-export for call sites that want to narrow responses.
export type { DashboardPaths, SdkPaths }

// Escape hatch for surfaces that need a raw fetch with the same
// middleware semantics (e.g. polling endpoints that hit an unversioned
// ingress path). Keep use rare — prefer `api` / `sdk`.
export async function rawFetch(input: string, init: RequestInit = {}): Promise<Response> {
  const cfg = getRuntimeConfig()
  const url = input.startsWith('http') ? input : `${cfg.API_URL}${input}`
  const method = (init.method ?? 'GET').toUpperCase()
  const headers = new Headers(init.headers)
  if (MUTATION_METHODS.has(method)) {
    const csrf = readCsrfCookie()
    if (csrf) headers.set('X-CSRF-Token', csrf)
  }

  const resp = await fetch(url, { ...init, headers, credentials: 'include' })

  if (
    resp.status === 401 &&
    !headers.get(RETRY_HEADER) &&
    !isAuthBoundaryPath(url)
  ) {
    try {
      await refreshWithLock()
    } catch {
      // Refresh failed; surface the ORIGINAL 401 so callers see the
      // backend's real error message (e.g. "Invalid email or
      // password" on /auth/login) instead of a synthesized "session
      // expired" stand-in.
      await throwTypedError(resp)
    }
    const retryHeaders = new Headers(init.headers)
    retryHeaders.set(RETRY_HEADER, '1')
    const retryResp = await fetch(url, {
      ...init,
      headers: retryHeaders,
      credentials: 'include',
    })
    // Same 2xx invariant as the non-retry branch. Without this, a
    // replayed 401 (or any non-ok status after refresh) resolves the
    // caller's `await resp.json()` with the error envelope and
    // silently corrupts typed reads (e.g. SessionUser cast in
    // currentUserQueryOptions).
    if (!retryResp.ok) await throwTypedError(retryResp)
    return retryResp
  }

  if (!resp.ok) await throwTypedError(resp)
  return resp
}
