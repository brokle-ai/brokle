import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { type ReactNode } from 'react'

import { server } from '@/mocks/server'
import { useAuthStore } from '@/stores/auth-store'
import { useCurrentUser } from '../hooks'
import { authKeys } from '../queries'

// Tests cover the regression where TanStack Query's documented
// stale-while-revalidate state ({ data: <stale>, error: <new> })
// caused useCurrentUser to re-seed the store and leave consumers
// reading the cached user after a /me 401. The hook must (a) flip
// the store via notifySessionEnded and (b) mask `data` to undefined
// so direct consumers (e.g. /accept-invite) fall into their
// unauthenticated branch.

const MOCK_USER = {
  id: 'usr_test_0000000000000000000000',
  email: 'a@b.test',
  first_name: 'A',
  last_name: 'B',
  default_organization_id: 'org_test_0000000000000000000000',
  is_email_verified: true,
  created_at: '2026-01-01T00:00:00Z',
} as const

function makeWrapper() {
  // retry: false so AuthenticationError surfaces on the first 401
  // exactly like in production (currentUserQueryOptions sets it).
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  }
  return { qc, Wrapper }
}

beforeEach(() => {
  // Reset Zustand store between tests (no built-in reset; this is the
  // documented pattern from zustand docs).
  useAuthStore.setState({ user: null, isAuthenticated: false })
})

describe('useCurrentUser', () => {
  it('seeds the auth store on a successful /me', async () => {
    server.use(
      http.get('*/api/v1/users/me', () => HttpResponse.json(MOCK_USER)),
    )
    const { qc, Wrapper } = makeWrapper()
    const { result } = renderHook(() => useCurrentUser(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.data).toBeTruthy())
    expect(result.current.data?.id).toBe(MOCK_USER.id)
    expect(useAuthStore.getState().user?.id).toBe(MOCK_USER.id)
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
    qc.clear()
  })

  it('clears the store and masks data when /me refetches 401 after a prior success', async () => {
    server.use(
      http.get('*/api/v1/users/me', () => HttpResponse.json(MOCK_USER)),
    )
    const { qc, Wrapper } = makeWrapper()
    const { result } = renderHook(() => useCurrentUser(), { wrapper: Wrapper })

    // First fetch: success — store gets populated.
    await waitFor(() => expect(result.current.data?.id).toBe(MOCK_USER.id))
    expect(useAuthStore.getState().isAuthenticated).toBe(true)

    // Session expires server-side. The next /me returns 401.
    server.use(
      http.get('*/api/v1/users/me', () =>
        HttpResponse.json(
          { error: { type: 'authentication_error', message: 'Session expired' } },
          { status: 401 },
        ),
      ),
    )

    // Force a refetch — equivalent to the staleTime expiring on a
    // navigation that revisits /me.
    await act(async () => {
      await qc.refetchQueries({ queryKey: authKeys.currentUser() })
    })

    await waitFor(() => expect(result.current.error).toBeTruthy())
    // Regression assertions:
    //   1. Store is cleared (notifySessionEnded fired).
    //   2. Data is masked to undefined so consumers don't see a
    //      cached user even though TanStack still holds one.
    expect(useAuthStore.getState().user).toBeNull()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(result.current.data).toBeUndefined()
    qc.clear()
  })

  it('keeps cached data and store on a non-auth refetch failure (5xx)', async () => {
    server.use(
      http.get('*/api/v1/users/me', () => HttpResponse.json(MOCK_USER)),
    )
    const { qc, Wrapper } = makeWrapper()
    const { result } = renderHook(() => useCurrentUser(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.data?.id).toBe(MOCK_USER.id))

    // Transient server failure — must NOT sign the user out.
    server.use(
      http.get('*/api/v1/users/me', () =>
        HttpResponse.json(
          { error: { type: 'api_error', message: 'Internal server error' } },
          { status: 500 },
        ),
      ),
    )
    await act(async () => {
      await qc.refetchQueries({ queryKey: authKeys.currentUser() })
    })

    await waitFor(() => expect(result.current.error).toBeTruthy())
    // Cached data still surfaced; store still authenticated.
    expect(result.current.data?.id).toBe(MOCK_USER.id)
    expect(useAuthStore.getState().user?.id).toBe(MOCK_USER.id)
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
    qc.clear()
  })

  it('does not seed the store on the very first /me 401 (no prior success)', async () => {
    server.use(
      http.get('*/api/v1/users/me', () =>
        HttpResponse.json(
          { error: { type: 'authentication_error', message: 'Unauthenticated' } },
          { status: 401 },
        ),
      ),
    )
    const { qc, Wrapper } = makeWrapper()
    const { result } = renderHook(() => useCurrentUser(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.error).toBeTruthy())
    expect(result.current.data).toBeUndefined()
    expect(useAuthStore.getState().user).toBeNull()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    qc.clear()
  })
})
