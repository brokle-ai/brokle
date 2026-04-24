// Data Access Layer (DAL) — server-side authentication and user data.
//
// Pattern from the official Next.js authentication guide:
// https://nextjs.org/docs/app/guides/authentication#creating-a-data-access-layer-dal
//
// The DAL is the canonical Next.js 16 entry point for server-side
// session verification and user data fetching. Server Components,
// Server Actions, and Route Handlers call these helpers instead of
// re-implementing cookie reads + backend calls + redirect logic.
//
// Memoization: every export is wrapped in React's `cache()` so a
// single render pass making multiple calls (e.g. layout AND page
// both calling `getCurrentUser()`) hits the backend exactly once.
// `cache()` scope is per-request, so different users never see each
// other's data.
//
// Redirect-on-failure: 401 from the backend redirects via Next.js
// `redirect()`, which throws a special signal Next.js intercepts at
// the framework boundary. Callers don't need try/catch — a 401 from
// the DAL means the request never returns; it short-circuits with a
// 302 to /signin. proxy.ts catches the cold-load case before the
// request reaches a Server Component; the DAL catches the runtime
// case (cookie present but server-side validity check fails).
//
// Why no JWT-decode here: the backend Go service is the authority
// on token validity (signing key, blacklist, expiry). The DAL
// performs an actual API call, not optimistic decode. Cost: one
// network hop per render. Benefit: zero risk of trusting a forged
// or revoked token.

import 'server-only'

import { cache } from 'react'
import { cookies, headers } from 'next/headers'
import { redirect } from 'next/navigation'

import { ROUTES } from '@/lib/routes'
import { fetchBackend } from './server-fetch'
import { mapEnhancedUserProfile, type MappedProfile } from './profile-mapper'
import type { EnhancedUserProfileResponse } from '@/types/api-responses'

const ACCESS_TOKEN_COOKIE = 'access_token'
const REFRESH_TOKEN_COOKIE = 'refresh_token'

// Stamped on every dashboard request by proxy.ts so the DAL can
// reconstruct the current URL when building the return-path for a
// silent-refresh redirect. Missing = proxy didn't run (misconfigured
// matcher or direct SSR render outside the dashboard matcher); the
// DAL falls back to "/" which still works — the user just lands on
// the dashboard home rather than their originally-requested page.
const PATHNAME_HEADER = 'x-brokle-pathname'

// Loop-guard cookie set by the silent-refresh route handler
// (web/src/app/api/auth/silent-refresh/route.ts) on successful
// refresh, with a very short Max-Age (seconds). If /users/me STILL
// returns 401 while this cookie is present, a second refresh
// wouldn't help — redirect to /signin. After the Max-Age elapses
// the cookie is gone, so a genuine later 401 (the next 15-minute
// access-token rollover on the same URL) is treated as fresh and
// recoverable. A URL query-param would persist for the life of
// the navigation and block every future refresh from the same
// page — that's the regression we're avoiding.
const REFRESH_MARKER_COOKIE = 'brokle_refresh_attempt'

/**
 * Optimistic session check — verifies a session cookie exists.
 * Does NOT verify validity (that requires a backend call). Use this
 * for early redirects when you want to skip the backend round-trip
 * for the obvious "user has zero cookies" case.
 *
 * Throws via `redirect()` if no session cookie present.
 */
export const verifySession = cache(async (): Promise<{ hasSession: true }> => {
  const cookieStore = await cookies()
  const hasSession =
    cookieStore.has(ACCESS_TOKEN_COOKIE) ||
    cookieStore.has(REFRESH_TOKEN_COOKIE)

  if (!hasSession) {
    redirect(ROUTES.SIGNIN)
  }

  return { hasSession: true }
})

/**
 * Fetches the authenticated user's profile + organizations + projects
 * from the backend in a single call.
 *
 * On 401 the DAL tries to recover before giving up. The proxy handles
 * the common Max-Age eviction case (access cookie missing on entry);
 * this 401 branch covers the complementary class — access cookie
 * present but server-rejected (clock skew, admin revocation, JWT
 * key rotation, user-wide timestamp blacklist). Next.js 16 forbids
 * cookie mutation from Server Component rendering, so the DAL
 * delegates the refresh + Set-Cookie work to a Route Handler at
 * /api/auth/silent-refresh which redirects back here once cookies
 * are rotated.
 *
 * Mirrors the shape WorkspaceProvider produces client-side, so the
 * server-fetched payload can be passed straight into the workspace
 * context's `initialData` slot without re-mapping on the client.
 */
export const getCurrentUser = cache(async (): Promise<MappedProfile> => {
  await verifySession()

  const res = await fetchBackend('/api/v1/users/me')
  if (res.status === 401) {
    await handleUnauthorized()
  }
  if (!res.ok) {
    throw new Error(
      `dal.getCurrentUser: backend returned ${res.status} ${res.statusText}`,
    )
  }

  // Stripe/OpenAI-style contract: 2xx body IS the resource. No
  // envelope wrapper. res.ok is true here, so the body parses
  // straight into the EnhancedUserProfileResponse shape.
  const profile = (await res.json()) as EnhancedUserProfileResponse
  return mapEnhancedUserProfile(profile)
})

// handleUnauthorized is the /users/me-401 recovery policy:
//
//   1. If the `brokle_refresh_attempt` cookie is present, a refresh
//      completed within the last few seconds and /users/me is STILL
//      401. A second refresh wouldn't help — redirect to /signin to
//      avoid an infinite refresh loop. The cookie auto-expires
//      (Max-Age seconds), so a 401 on the *next* access-token
//      rollover on the same URL is treated as fresh and recoverable.
//
//   2. Else if the refresh_token cookie is gone, the refresh
//      endpoint would 401 too — skip the round-trip and go straight
//      to /signin.
//
//   3. Otherwise defer to the silent-refresh route handler, which
//      can set cookies. It rotates the session via the backend and
//      redirects back here.
//
// Always exits via `redirect()` — never returns. Callers treat this
// as `Promise<never>`.
async function handleUnauthorized(): Promise<never> {
  const [hdrs, cookieStore] = await Promise.all([headers(), cookies()])

  if (cookieStore.has(REFRESH_MARKER_COOKIE)) {
    redirect(`${ROUTES.SIGNIN}?status=expired`)
  }
  if (!cookieStore.has(REFRESH_TOKEN_COOKIE)) {
    redirect(`${ROUTES.SIGNIN}?status=expired`)
  }

  const pathname = hdrs.get(PATHNAME_HEADER) ?? '/'
  redirect(
    `/api/auth/silent-refresh?return=${encodeURIComponent(pathname)}`,
  )
}
