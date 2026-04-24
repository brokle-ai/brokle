import { NextResponse, type NextRequest } from 'next/server'

import { getBackendBaseURL } from '@/lib/env/backend-url'

// Next.js 16 file convention: `proxy.ts` (formerly `middleware.ts`).
// Renamed in 16.0 to clarify the network-boundary role and decouple
// the name from Express-style middleware semantics. Default runtime
// is Node.js (the deprecated middleware.ts ran on Edge).
//
// Responsibilities:
//
//   1. Cookie-presence route guard. Requests with zero session
//      cookies redirect to /signin, preserving the originally
//      requested URL as `redirect` so the signin page can bounce
//      the user back after authentication.
//
//   2. Silent SSR refresh. When the short-lived `access_token`
//      cookie has been evicted by its own Max-Age but the longer-
//      lived `refresh_token` cookie is still present, the proxy
//      calls the Go backend's /api/v1/auth/refresh endpoint,
//      forwards the resulting Set-Cookie headers to the browser,
//      and rewrites the inbound request headers so the Server
//      Component sees the fresh access_token via next/headers
//      `cookies()` on the very next `await` inside the Dashboard
//      layout DAL.
//
//      This is the canonical Next.js 16 refresh site. Per the
//      official `cookies()` API reference: "Setting cookies is not
//      supported during Server Component rendering. To modify
//      cookies, invoke a Server Function from the client or use a
//      Route Handler." Middleware satisfies the constraint because
//      NextResponse.next() runs before the Server Component stream
//      begins, and its `.headers.append('set-cookie', …)` is an
//      actual HTTP response header that reaches the browser on the
//      same round-trip.
//
//      Reference architecture: Auth.js v5, Clerk, Supabase —
//      all drive SSR refresh from middleware for the same reason.
//
// This is an OPTIMISTIC gate — the Go backend remains the authority
// on token validity. If the access_token is rejected server-side
// after the proxy admits the request (blacklist, signature, clock
// skew), the DAL still redirects to /signin. That branch is
// unrecoverable via refresh and should terminate the session.

const ACCESS_TOKEN_COOKIE = 'access_token'
const REFRESH_TOKEN_COOKIE = 'refresh_token'

// Pathname header stamped on every forwarded request so Server
// Components (specifically the DAL at web/src/lib/auth/dal.ts) can
// read the current URL — Next.js 16 does not expose the incoming
// URL to Server Components out of the box. The DAL uses this to
// build the `return` param on its silent-refresh redirect.
const PATHNAME_HEADER = 'x-brokle-pathname'

// Headers forwarded to the backend on the outbound /refresh call so
// audit logs attribute the refresh to the real end-user, not the
// Next.js pod. Same set the DAL's server-fetch forwards — keep in
// lockstep.
const FORWARDED_HEADERS = [
  'x-forwarded-for',
  'x-real-ip',
  'true-client-ip',
  'x-forwarded-proto',
  'x-forwarded-host',
  'user-agent',
] as const

export async function proxy(req: NextRequest): Promise<NextResponse> {
  const hasAccess = req.cookies.has(ACCESS_TOKEN_COOKIE)
  const hasRefresh = req.cookies.has(REFRESH_TOKEN_COOKIE)

  // No session at all — straight to /signin.
  if (!hasAccess && !hasRefresh) {
    return redirectToSignin(req)
  }

  // Access present (with or without refresh) — pass through. The Go
  // backend is the authority on validity; a 401 after this point
  // is recoverable iff the refresh cookie is still valid, so the
  // DAL may redirect to /api/auth/silent-refresh and try one more
  // time before surrendering to /signin.
  if (hasAccess) {
    return NextResponse.next({
      request: { headers: stampPathname(req) },
    })
  }

  // Access absent, refresh present — the short-lived cookie was
  // evicted by Max-Age. Mint a fresh pair server-to-server before
  // rendering begins.
  const refreshed = await attemptRefresh(req)
  if (!refreshed) {
    return redirectToSignin(req, 'expired')
  }
  return refreshed
}

// attemptRefresh performs one POST to the backend refresh endpoint
// using the inbound cookie jar. On 200 it returns a NextResponse.next()
// that both (a) emits Set-Cookie headers to the browser and (b)
// rewrites the inbound Cookie header so the Server Component render
// sees the fresh access_token via `cookies()`.
//
// Returns null on refresh failure so the caller can redirect cleanly.
// A null return never tries to set cookies — the backend's 401
// response already carries clear-cookie Set-Cookie headers, but we
// deliberately don't forward them here: the redirect response the
// caller returns is terminal; the browser will simply overwrite the
// refresh cookie on the next signin.
async function attemptRefresh(
  req: NextRequest,
): Promise<NextResponse | null> {
  const cookieHeader = req.cookies
    .getAll()
    .map(c => `${c.name}=${c.value}`)
    .join('; ')

  const outboundHeaders: Record<string, string> = {
    cookie: cookieHeader,
    'content-type': 'application/json',
    accept: 'application/json',
  }
  for (const name of FORWARDED_HEADERS) {
    const v = req.headers.get(name)
    if (v) outboundHeaders[name] = v
  }

  let res: Response
  try {
    res = await fetch(`${getBackendBaseURL()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: outboundHeaders,
      // Auth-bearing responses must never land in Next.js's fetch cache.
      cache: 'no-store',
    })
  } catch {
    // Network error / backend unreachable. Treat as refresh failure —
    // the alternative (let the request through with no session) would
    // cascade into a DAL redirect anyway, so short-circuit here.
    return null
  }

  if (!res.ok) {
    return null
  }

  // Build the downstream request's rewritten Cookie header so the
  // Server Component DAL sees the fresh access_token when it reads
  // `cookies()`. Without this the DAL would still forward the stale
  // (missing) access cookie on the same request and get 401'd.
  const setCookies = res.headers.getSetCookie()
  const newValues = parseSetCookieNames(setCookies)
  const rewrittenHeaders = stampPathname(req)
  rewrittenHeaders.set(
    'cookie',
    rebuildCookieHeader(req, newValues),
  )

  const response = NextResponse.next({
    request: { headers: rewrittenHeaders },
  })

  // Forward Set-Cookie headers verbatim to the browser. Node 20+
  // fetch returns them as a distinct array via getSetCookie(); each
  // entry is one full Set-Cookie header string (name=value + attrs).
  // NextResponse.headers.append preserves multiple Set-Cookie entries
  // as distinct headers on the wire — required by RFC 6265.
  for (const sc of setCookies) {
    response.headers.append('set-cookie', sc)
  }

  return response
}

// parseSetCookieNames extracts just the name→value pairs from a list
// of Set-Cookie headers. Attributes (Path, Max-Age, HttpOnly, etc.)
// flow through the browser verbatim via forwardSetCookies — this
// function only needs enough to rebuild the outbound Cookie header
// for the rewritten downstream request.
function parseSetCookieNames(headers: string[]): Map<string, string> {
  const out = new Map<string, string>()
  for (const h of headers) {
    const semi = h.indexOf(';')
    const pair = semi === -1 ? h : h.slice(0, semi)
    const eq = pair.indexOf('=')
    if (eq === -1) continue
    const name = pair.slice(0, eq).trim()
    const value = pair.slice(eq + 1).trim()
    if (name) out.set(name, value)
  }
  return out
}

// rebuildCookieHeader merges the inbound request's cookie jar with
// the fresh name→value pairs from the backend refresh response. New
// values override stale ones (covers the rotating refresh_token case
// where the server rolls the token on every refresh).
function rebuildCookieHeader(
  req: NextRequest,
  newValues: Map<string, string>,
): string {
  const jar = new Map<string, string>()
  for (const c of req.cookies.getAll()) jar.set(c.name, c.value)
  for (const [name, value] of newValues) jar.set(name, value)
  return [...jar.entries()].map(([n, v]) => `${n}=${v}`).join('; ')
}

// stampPathname returns a new Headers object derived from the
// inbound request with x-brokle-pathname set to the current URL.
// Used so Server Components can reconstruct the pathname via
// next/headers `headers()` — Next.js 16 does not expose the current
// URL natively.
function stampPathname(req: NextRequest): Headers {
  const headers = new Headers(req.headers)
  headers.set(PATHNAME_HEADER, req.nextUrl.pathname + req.nextUrl.search)
  return headers
}

function redirectToSignin(req: NextRequest, status?: string): NextResponse {
  const signinUrl = new URL('/signin', req.url)
  if (status) signinUrl.searchParams.set('status', status)
  signinUrl.searchParams.set(
    'redirect',
    req.nextUrl.pathname + req.nextUrl.search,
  )
  return NextResponse.redirect(signinUrl)
}

// The matcher excludes:
//  - Public auth pages under app/(auth)/ — signin, signup,
//    forgot-password, reset-password (reserved for the email-link
//    flow when the page ships), verify-email, callback (OAuth
//    provider callback, receives the authorization code BEFORE any
//    Brokle session exists), accept-invite (invitation acceptance
//    via an email-link token, invitees have no session yet).
//    These MUST stay in lockstep with the directories under
//    app/(auth)/. A drift here silently breaks a public entry flow —
//    e.g. an invitee gets 302'd to /signin with their token encoded
//    into the redirect param, and the real page never runs.
//    Asserted at test time by src/proxy.test.ts.
//  - /api/* and /v1/* — Next.js rewrites proxy these to the Go
//    backend, which has its own auth (RequireAuth / RequireSDKAuth).
//  - /_next/* — Next.js build output (RSC streams, JS, CSS).
//  - Common static asset extensions — favicons, fonts, images.
//
// Anything else (i.e. /, /dashboard, /projects/…) goes through the
// presence check (and conditional silent refresh) above.
export const config = {
  matcher: [
    '/((?!signin|signup|forgot-password|reset-password|verify-email|callback|accept-invite|api|_next|v1|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|woff|woff2|ttf)$).*)',
  ],
}
