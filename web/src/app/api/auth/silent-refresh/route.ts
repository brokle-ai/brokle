// Silent-refresh route handler.
//
// Used by the Data Access Layer (web/src/lib/auth/dal.ts) on any 401
// from /api/v1/users/me where a refresh_token cookie is still
// present. The DAL cannot set cookies itself — Next.js 16 forbids
// cookie mutation from Server Component rendering (cookies() API:
// "Setting cookies is not supported during Server Component
// rendering") — so it redirects the browser here. This route
// handler:
//
//   1. Reads the refresh_token from the inbound cookie jar.
//   2. POSTs to the Go backend's /api/v1/auth/refresh.
//   3. On 200: forwards the backend's Set-Cookie headers verbatim,
//      ALSO emits a short-lived `brokle_refresh_attempt` marker
//      cookie (Max-Age a few seconds), and redirects the browser
//      back to the original URL (carried on the `return` query
//      param). The DAL reads the marker cookie and loop-guards: if
//      /users/me STILL 401s within the marker's lifetime, a second
//      refresh wouldn't help, so redirect to /signin. The marker
//      auto-expires, so a later 401 (e.g. the next 15-minute
//      access-token rollover on the same URL) is treated as fresh
//      and recoverable.
//   4. On non-2xx / network error: forwards the backend's clear-
//      cookie Set-Cookie headers (if any) and redirects to
//      /signin?status=expired, preserving the original URL as the
//      `redirect` query param so signin can bounce back post-auth.
//
// This pattern is canonical for SSR refresh-on-401 (Auth.js v5,
// Clerk, Supabase — all use a dedicated refresh endpoint hit via
// server-initiated redirect). It covers the edge cases proxy.ts
// cannot reach: access_token cookie present but server-rejected
// (clock skew, blacklist, JWT key rotation).

import { NextResponse, type NextRequest } from 'next/server'

import { getBackendBaseURL } from '@/lib/env/backend-url'

const REFRESH_TOKEN_COOKIE = 'refresh_token'

// brokle_refresh_attempt is the loop-guard marker. Set on a
// successful refresh with a very short Max-Age so that if the very
// next /users/me on the return URL still 401s, the DAL terminates
// to /signin instead of re-entering this route (infinite refresh).
// Outside the short window the cookie has auto-expired, so a
// genuine later 401 (the next 15-minute access-token rollover) is
// treated as fresh and recoverable.
const REFRESH_MARKER_COOKIE = 'brokle_refresh_attempt'
const REFRESH_MARKER_MAX_AGE_SECONDS = 5

// Forwarded for backend audit-log attribution.
const FORWARDED_HEADERS = [
  'x-forwarded-for',
  'x-real-ip',
  'true-client-ip',
  'x-forwarded-proto',
  'x-forwarded-host',
  'user-agent',
] as const

export async function GET(req: NextRequest): Promise<NextResponse> {
  const rawReturn = req.nextUrl.searchParams.get('return') ?? '/'
  const safeReturn = sanitiseReturnPath(rawReturn)

  // No refresh cookie — no recoverable session. Skip the backend
  // round-trip and redirect straight to signin.
  if (!req.cookies.has(REFRESH_TOKEN_COOKIE)) {
    return signinRedirect(req, safeReturn)
  }

  let backendRes: Response
  try {
    backendRes = await fetch(`${getBackendBaseURL()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: buildBackendHeaders(req),
      cache: 'no-store',
    })
  } catch {
    // Network error / backend unreachable. Surface as expired —
    // the alternative (silent success) would strand the user on a
    // page that can't authenticate.
    return signinRedirect(req, safeReturn)
  }

  if (!backendRes.ok) {
    // Backend explicitly rejected refresh (401 with cleared cookies,
    // or 5xx). Forward any clearing Set-Cookie headers so the jar
    // matches the backend's view, then surface as expired.
    const response = signinRedirect(req, safeReturn)
    forwardSetCookies(backendRes, response)
    return response
  }

  // Refresh succeeded. Redirect back to the original path (no URL
  // pollution — the loop guard rides a short-lived cookie instead)
  // and forward the three fresh cookies verbatim plus the marker.
  //
  // All four cookies (the three from the backend plus the marker)
  // are emitted via headers.append('set-cookie', …). NextResponse's
  // `.cookies.set(…)` API goes through a separate internal store
  // that doesn't compose with raw Set-Cookie header appends — mixing
  // them drops the forwarded cookies on some Next.js versions
  // (observed on 16.x). Keep one code path.
  const destination = new URL(safeReturn, req.url)
  const response = NextResponse.redirect(destination)
  forwardSetCookies(backendRes, response)
  response.headers.append('set-cookie', buildMarkerCookie())
  return response
}

function buildMarkerCookie(): string {
  const parts = [
    `${REFRESH_MARKER_COOKIE}=1`,
    'Path=/',
    `Max-Age=${REFRESH_MARKER_MAX_AGE_SECONDS}`,
    'HttpOnly',
    'SameSite=Lax',
  ]
  if (process.env.NODE_ENV === 'production') parts.push('Secure')
  return parts.join('; ')
}

// sanitiseReturnPath accepts only same-origin relative paths to
// prevent open-redirect abuse via the `return` query param. Anything
// that isn't a single-slash-prefixed path is collapsed to "/".
function sanitiseReturnPath(raw: string): string {
  if (!raw.startsWith('/')) return '/'
  // Reject protocol-relative URLs ("//host/path") which some browsers
  // treat as absolute.
  if (raw.startsWith('//')) return '/'
  return raw
}

function buildBackendHeaders(req: NextRequest): Record<string, string> {
  const cookieHeader = req.cookies
    .getAll()
    .map(c => `${c.name}=${c.value}`)
    .join('; ')
  const out: Record<string, string> = {
    cookie: cookieHeader,
    'content-type': 'application/json',
    accept: 'application/json',
  }
  for (const name of FORWARDED_HEADERS) {
    const v = req.headers.get(name)
    if (v) out[name] = v
  }
  return out
}

function signinRedirect(req: NextRequest, returnPath: string): NextResponse {
  const url = new URL('/signin', req.url)
  url.searchParams.set('status', 'expired')
  url.searchParams.set('redirect', returnPath)
  return NextResponse.redirect(url)
}

function forwardSetCookies(from: Response, to: NextResponse): void {
  for (const sc of from.headers.getSetCookie()) {
    to.headers.append('set-cookie', sc)
  }
}
