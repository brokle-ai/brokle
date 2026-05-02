import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { currentUserQueryOptions } from '@/features/auth/queries'
import { AuthenticationError } from '@/lib/api/errors'
import { resetSession } from '@/lib/auth/session'
import { useAuthStore } from '@/stores/auth-store'

// Pathless layout that gates every dashboard route. Backend owns
// session validity (CLAUDE.md gotcha #22), so the guard asks
// `/v1/users/me` rather than trusting a local snapshot.
//
// Cache strategy: `currentUserQueryOptions` has a short `staleTime`
// (30s). The gate uses `fetchQuery` (NOT `ensureQueryData`) so it
// honours that staleness: cache fresh → return cached, no /me round
// trip; cache stale (>30s) or missing → refetch. `ensureQueryData`
// returns cached data regardless of staleness, which would let an
// expired backend session keep admitting navigations indefinitely
// after the first successful /me.
//
// Bounded admission window: 30s is long enough for typical click
// cadence (no /me storm on rapid navigation), short enough that a
// session-ended state surfaces within a half-minute on the next
// nav.
//
// `context.auth` from main.tsx is a fast-path UI hint (header avatar,
// sidebar), not a security gate. The gate is this beforeLoad.
export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ context, location }) => {
    try {
      const user = await context.queryClient.fetchQuery(
        currentUserQueryOptions(),
      )
      // Always mirror the latest /me into the store. Backend is
      // authority on every field (default_organization_id,
      // first_name, is_email_verified, …); on cache hits this writes
      // the same reference (Zustand no-op for selectors via Object.is).
      useAuthStore.getState().setUser(user)
    } catch (err) {
      if (err instanceof AuthenticationError) {
        // Drop the cache too — the previous session's org/project
        // queries must not survive into the redirect or the eventual
        // re-login.
        resetSession(context.queryClient)
        // `session: 'expired'` wires through to SignInToastHandler so
        // the destination /signin page shows a toast explaining why
        // the user landed there. This is the only signal — the form
        // itself does not track post-login state.
        throw redirect({
          to: '/signin',
          search: { redirect: location.href, session: 'expired' },
        })
      }
      throw err
    }
  },
  component: () => <Outlet />,
})
