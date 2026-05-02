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

// Public auth endpoints (no session required). Refresh-retry must be
// SKIPPED for these — their 401 means "wrong credentials" / "expired
// reset token" / "missing refresh cookie" / etc., not "session
// expired and refreshable." Mirrors the backend's
// `RegisterPublicRoutes` set in
// `internal/transport/http/handlers/auth/routes.go`.
//
// All OTHER /api/v1/auth/* paths are session-authenticated
// (`RegisterProtectedRoutes`: /auth/me, /auth/logout,
// /auth/change-password, /auth/profile, /auth/sessions[/...]) and
// MUST go through normal refresh-retry — an expired access token on
// those should rotate via the refresh cookie, not surface as a
// "session expired" forced re-login.
const PUBLIC_AUTH_PATHS = new Set<string>([
  '/api/v1/auth/login',
  '/api/v1/auth/signup',
  '/api/v1/auth/refresh',
  '/api/v1/auth/forgot-password',
  '/api/v1/auth/reset-password',
  '/api/v1/auth/google',
  '/api/v1/auth/google/callback',
  '/api/v1/auth/github',
  '/api/v1/auth/github/callback',
  '/api/v1/auth/complete-oauth-signup',
])

// Same idea but for paths with a dynamic suffix (e.g. session_id).
// Trailing slash is significant — prevents matching unrelated paths
// like /api/v1/auth/exchange-sessions.
const PUBLIC_AUTH_PREFIXES: readonly string[] = [
  '/api/v1/auth/exchange-session/',
]

export function isAuthBoundaryPath(url: string): boolean {
  try {
    // Derive the deployment's path prefix from the same API_URL
    // rawFetch uses to build request URLs, so they cannot drift.
    // Handles every shape: '' (relative), 'http://localhost:8080'
    // (absolute, no prefix), '/backend', 'https://host/proxy',
    // 'https://host/proxy/' (trailing slash stripped).
    const apiBase = new URL(
      getRuntimeConfig().API_URL || '/',
      window.location.origin,
    )
    const prefix = apiBase.pathname.replace(/\/$/, '')
    const pathname = url.startsWith('http')
      ? new URL(url).pathname
      : (url.split('?')[0] ?? '')

    // Strip the deployment prefix once so the allow-list is checked
    // against the backend's logical paths.
    const normalized =
      prefix && pathname.startsWith(prefix)
        ? pathname.slice(prefix.length)
        : pathname

    if (PUBLIC_AUTH_PATHS.has(normalized)) return true
    return PUBLIC_AUTH_PREFIXES.some((p) => normalized.startsWith(p))
  } catch {
    return false
  }
}

// Inject the current CSRF token from the cookie into mutation
// requests. Idempotent: safe to call multiple times on the same
// Headers object — the latest cookie value wins. This matters for
// retry paths after `refreshWithLock`, because /auth/refresh rotates
// the csrf_token cookie alongside access/refresh, so the retried
// request MUST re-read the cookie or the backend's double-submit
// check fails (header[old] ≠ cookie[new]).
function applyCsrfHeader(headers: Headers, method: string): void {
  if (!MUTATION_METHODS.has(method.toUpperCase())) return
  const csrf = readCsrfCookie()
  if (csrf) headers.set('X-CSRF-Token', csrf)
}

const csrfMiddleware: Middleware = {
  async onRequest({ request }) {
    applyCsrfHeader(request.headers, request.method)
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
    // Refresh rotated the csrf_token cookie; the preserved CSRF
    // header is now stale and the backend's double-submit check
    // would 403. Re-read the cookie on every retry.
    applyCsrfHeader(retry.headers, retry.method)
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
  applyCsrfHeader(headers, method)

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
    // Rebuild headers from the caller's init (NOT the augmented
    // `headers` above) and re-inject CSRF from the freshly-rotated
    // cookie. The previous request's CSRF token is stale because
    // /auth/refresh issues a new csrf_token cookie alongside
    // access/refresh; sending the old one would 403 the retry.
    const retryHeaders = new Headers(init.headers)
    applyCsrfHeader(retryHeaders, method)
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
