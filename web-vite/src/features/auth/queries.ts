import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { SessionUser } from '@/stores/auth-store'

export const authKeys = {
  all: ['auth'] as const,
  currentUser: () => [...authKeys.all, 'currentUser'] as const,
} as const

// `staleTime: 5min` matches the backend session TTL granularity and
// neutralises TanStack Router issue #3997 — `beforeLoad` +
// `ensureQueryData` now read from cache on every in-tab navigation,
// not the network.
export const currentUserQueryOptions = () =>
  queryOptions({
    queryKey: authKeys.currentUser(),
    queryFn: async () => {
      const resp = await rawFetch('/api/v1/users/me', { method: 'GET' })
      return (await resp.json()) as SessionUser
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
