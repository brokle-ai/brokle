// Server-side fetch helper for the Brokle Go backend.
//
// Forwards the incoming request's cookies + proxy/forwarding headers
// (read via Next.js' `cookies()` and `headers()` APIs) to the upstream
// backend over a same-process HTTP call. Used by the Data Access Layer
// (lib/auth/dal.ts) and any other Server Component / Route Handler /
// Server Action that needs to talk to the backend on behalf of the
// current user.
//
// Why a dedicated helper:
//   - Server-side fetches BYPASS the next.config.ts rewrites that
//     proxy /api/v1/* and /v1/* in the browser. Server-side calls
//     must use the absolute backend URL.
//   - httpOnly cookies are not visible to client JS, so the backend
//     would receive no auth context unless we manually forward the
//     cookie header from the Next.js request.
//   - cache: 'no-store' is mandatory for auth-bearing fetches —
//     Next.js' default fetch cache would otherwise cache the response
//     and serve another user's data.
//   - Forwarding X-Forwarded-For / X-Real-IP / User-Agent from the
//     incoming request preserves end-user attribution at the Go
//     backend's audit-log layer (httpctx.ClientIP / httpctx.UserAgent).
//     Without it, every SSR fetch logs as the Next.js peer IP and
//     the Node fetch user agent, which poisons observability data.
//     This is a correctness fix for logs/audit; rate limiting no
//     longer depends on IP for authenticated traffic (see Go-side
//     routes.go for the per-principal bucketing).
//
// `'server-only'` is a runtime safety net — if a client component
// imports this file, the build fails. The DAL is not safe to ship
// to the browser bundle.

import 'server-only'

import { cookies, headers } from 'next/headers'

const BACKEND_URL =
  process.env.BROKLE_API_PROXY_TARGET || 'http://localhost:8080'

// Headers forwarded from the incoming Next.js request to the Go
// backend. Pass-through only — Next.js does not expose the upstream
// socket IP to headers(), so we cannot RFC-7239 append our own hop.
// The Go backend's chimw.RealIP takes the first public IP in XFF,
// so pass-through of what the edge already set is correct and
// sufficient. X-Forwarded-Proto / X-Forwarded-Host are forwarded so
// request logs render the original scheme/host (observability win,
// no behavioural change).
const FORWARDED_HEADERS = [
  'x-forwarded-for',
  'x-real-ip',
  'true-client-ip',
  'x-forwarded-proto',
  'x-forwarded-host',
  'user-agent',
] as const

/**
 * Server-side fetch to the Brokle backend with the incoming request's
 * cookies + forwarding headers propagated. Returns the raw Response
 * so callers can branch on status (DAL maps 401 → redirect, 5xx →
 * throw, etc.).
 */
export async function fetchBackend(
  path: string,
  init: RequestInit = {},
): Promise<Response> {
  const [cookieStore, requestHeaders] = await Promise.all([
    cookies(),
    headers(),
  ])

  const cookieHeader = cookieStore
    .getAll()
    .map(c => `${c.name}=${c.value}`)
    .join('; ')

  const outgoing = new Headers(init.headers)
  if (cookieHeader && !outgoing.has('cookie')) {
    outgoing.set('cookie', cookieHeader)
  }
  if (!outgoing.has('accept')) {
    outgoing.set('accept', 'application/json')
  }
  for (const name of FORWARDED_HEADERS) {
    const value = requestHeaders.get(name)
    if (value && !outgoing.has(name)) {
      outgoing.set(name, value)
    }
  }

  return fetch(`${BACKEND_URL}${path}`, {
    ...init,
    headers: outgoing,
    cache: 'no-store',
  })
}
