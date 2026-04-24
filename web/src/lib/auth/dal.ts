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
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { ROUTES } from '@/lib/routes'
import { fetchBackend } from './server-fetch'
import { mapEnhancedUserProfile, type MappedProfile } from './profile-mapper'
import type { EnhancedUserProfileResponse } from '@/types/api-responses'

const ACCESS_TOKEN_COOKIE = 'access_token'
const REFRESH_TOKEN_COOKIE = 'refresh_token'

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
 * from the backend in a single call. Redirects to /signin on 401.
 *
 * Mirrors the shape WorkspaceProvider produces client-side, so the
 * server-fetched payload can be passed straight into the workspace
 * context's `initialData` slot without re-mapping on the client.
 */
export const getCurrentUser = cache(async (): Promise<MappedProfile> => {
  await verifySession()

  const res = await fetchBackend('/api/v1/users/me')
  if (res.status === 401) {
    // Cookie was present (verifySession passed) but invalid or
    // expired server-side. Cleanest UX is to redirect to signin
    // with a status hint so the page can show "Session expired".
    redirect(`${ROUTES.SIGNIN}?status=expired`)
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
