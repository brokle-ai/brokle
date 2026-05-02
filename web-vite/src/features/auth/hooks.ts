import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { AuthenticationError } from '@/lib/api/errors'
import { notifySessionEnded } from '@/lib/auth/session'
import { currentUserQueryOptions } from './queries'

// Primary read hook for the authenticated user. Keeps the Zustand
// store in lockstep with the TanStack Query cache so route guards
// can read the store synchronously while the component tree reads
// via React Query (loading states, suspense, etc.).
//
// Branch order matters. TanStack Query v5 exposes `data` AND `error`
// simultaneously when a background refetch fails after a prior
// success — the documented stale-while-revalidate state
// (https://tanstack.com/query/latest/docs/framework/react/guides/migrating-to-v5).
// AuthenticationError must therefore be checked BEFORE `data`, or the
// cached payload re-seeds the store and the page continues acting as
// if the user is signed in. The bug pre-dated this comment because
// `staleTime` was effectively per-navigation; reducing it to 30s made
// the SWR state the common case, surfacing the latent ordering bug.
//
// The hook ALSO masks the cached `data` to undefined while an
// AuthenticationError is the latest result. The React Query cache
// itself stays populated (no eviction, no loop — see below), but
// consumers that read `data` directly (e.g. /accept-invite) get a
// coherent view: store cleared, data undefined, page falls into its
// unauthenticated branch.
//
// On AuthenticationError this calls `notifySessionEnded()` (NOT
// `resetSession(queryClient)`). The hook's observer is still mounted
// on routes that stay rendered while unauthenticated (e.g.
// /accept-invite); calling `queryClient.clear()` (or even
// `removeQueries` on this key) from inside the effect would drop the
// observer's own active query, React Query would immediately recreate
// it on the next render and re-fire the failing /me request — an
// infinite refetch loop. `notifySessionEnded` flips the auth store
// and runs registered store-reset callbacks (e.g.
// usePlaygroundStore.clearAll), so cross-account isolation is
// preserved. The full `resetSession(qc)` runs from sites that are
// about to unmount the observer (login/logout flows, the
// `_authenticated.beforeLoad` catch path that throws a redirect).
//
// Non-auth refetch failures (5xx, network glitch) are intentionally
// NOT treated as a session end. They leave the store and `data`
// intact so a transient failure doesn't sign the user out.
export function useCurrentUser() {
  const query = useQuery(currentUserQueryOptions())
  const setUser = useAuthStore((s) => s.setUser)
  const isAuthExpired = query.error instanceof AuthenticationError

  useEffect(() => {
    if (isAuthExpired) {
      notifySessionEnded()
      return
    }
    if (query.data) {
      setUser(query.data)
    }
  }, [query.data, isAuthExpired, setUser])

  if (isAuthExpired) {
    return { ...query, data: undefined }
  }
  return query
}
