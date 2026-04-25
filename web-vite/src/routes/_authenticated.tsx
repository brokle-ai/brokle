import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { currentUserQueryOptions } from '@/features/auth/queries'
import { AuthenticationError } from '@/lib/api/errors'
import { resetSession } from '@/lib/auth/session'
import { useAuthStore } from '@/stores/auth-store'

// Pathless layout that gates every dashboard route. Backend owns
// session validity (CLAUDE.md gotcha #22), so the guard asks
// `/v1/users/me` rather than trusting a local snapshot.
//
// Cache strategy: `currentUserQueryOptions` has `staleTime: 5min` +
// `retry: false`, so in-tab navigations hit the cache (no extra round
// trip). Cold reloads pay one `/me` request and seed the Zustand store
// for the rest of the session.
//
// `context.auth` from main.tsx is now a fast-path UI hint (header
// avatar, sidebar), not a security gate. The gate is this beforeLoad.
export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ context, location }) => {
    try {
      const user = await context.queryClient.ensureQueryData(
        currentUserQueryOptions(),
      )
      // Always mirror the latest /me into the store. Backend is
      // authority on every field (default_organization_id, first_name,
      // is_email_verified, …); skipping on matching ids would pin them
      // to the cold-load snapshot. `ensureQueryData` returns the cached
      // reference unchanged when the query is fresh, so this is a
      // Zustand no-op for warm in-tab navigations (selectors compare
      // user references via Object.is).
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
