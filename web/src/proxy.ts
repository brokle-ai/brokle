import { NextResponse, type NextRequest } from 'next/server'

// Next.js 16 file convention: `proxy.ts` (formerly `middleware.ts`).
// Renamed in 16.0 to clarify the network-boundary role and decouple
// the name from Express-style middleware semantics. Default runtime
// is Node.js (the deprecated middleware.ts ran on Edge).
//
// Cookie names mirror what the backend sets in pkg/cookies / handlers/auth.
// Both names are checked because the proxy should let the access-token
// flow through (handler decides validity) AND the refresh flow through
// (refresh handler decides validity). Only when BOTH are absent do we
// know there is no session at all and a redirect is unambiguously
// correct.
//
// This is an OPTIMISTIC check — proxy is the route guard, not the
// security perimeter. The Go backend remains the authority on token
// validity. Per Next.js auth guidance: proxy redirects fast at the
// network boundary to avoid loading flashes; backend rejects invalid
// sessions on the very next request.
//
// Reference architecture: NextAuth v5, Clerk, Supabase, Stack Auth —
// all converge on this shape (cookie-presence check + edge/network
// redirect + DAL on the server side for actual verification).
const ACCESS_TOKEN_COOKIE = 'access_token'
const REFRESH_TOKEN_COOKIE = 'refresh_token'

export function proxy(req: NextRequest) {
  const hasSession =
    req.cookies.has(ACCESS_TOKEN_COOKIE) ||
    req.cookies.has(REFRESH_TOKEN_COOKIE)

  if (hasSession) {
    return NextResponse.next()
  }

  // Preserve the originally-requested URL so the signin page can
  // redirect back after successful login. Matches the existing
  // signinWithStatus('redirect=<url>') convention at lib/routes.ts:32.
  const signinUrl = new URL('/signin', req.url)
  signinUrl.searchParams.set(
    'redirect',
    req.nextUrl.pathname + req.nextUrl.search
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
// Anything else (i.e. /, /dashboard, /projects/…) requires a session
// cookie or gets redirected.
export const config = {
  matcher: [
    '/((?!signin|signup|forgot-password|reset-password|verify-email|callback|accept-invite|api|_next|v1|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|woff|woff2|ttf)$).*)',
  ],
}
