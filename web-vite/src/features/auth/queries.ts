import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { SessionUser } from '@/stores/auth-store'

export const authKeys = {
  all: ['auth'] as const,
  currentUser: () => [...authKeys.all, 'currentUser'] as const,
} as const

// `staleTime: 30s` is the bounded admission window for the auth
// gate (`_authenticated.beforeLoad` via `ensureQueryData`). Short
// enough that a session-ended state surfaces within a half-minute
// on the next navigation; long enough that rapid intra-tree clicks
// hit the cache and don't fire `/me` on every page change. Same
// cache is read by component-side `useQuery(currentUserQueryOptions())`
// (header avatar, sidebar) — staleness is acceptable there too.
export const currentUserQueryOptions = () =>
  queryOptions({
    queryKey: authKeys.currentUser(),
    queryFn: async () => {
      const resp = await rawFetch('/api/v1/users/me', { method: 'GET' })
      return (await resp.json()) as SessionUser
    },
    staleTime: 30 * 1000,
    retry: false,
  })
