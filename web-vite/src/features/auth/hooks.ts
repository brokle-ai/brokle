import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { AuthenticationError } from '@/lib/api/errors'
import { currentUserQueryOptions } from './queries'

// Primary read hook for the authenticated user. Keeps the Zustand
// store in lockstep with the TanStack Query cache so route guards can
// read the store synchronously while the component tree reads via
// React Query (loading states, suspense, etc.).
export function useCurrentUser() {
  const query = useQuery(currentUserQueryOptions())
  const setUser = useAuthStore((s) => s.setUser)
  const expireSession = useAuthStore((s) => s.expireSession)

  useEffect(() => {
    if (query.data) {
      setUser(query.data)
    } else if (query.error instanceof AuthenticationError) {
      expireSession()
    }
  }, [query.data, query.error, setUser, expireSession])

  return query
}
