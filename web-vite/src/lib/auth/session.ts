import type { QueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth-store'

// Single primitive for crossing the auth boundary. Every login,
// logout, and terminal-401 path goes through this so the cache-reset
// invariant cannot drift between call sites.
//
// Order is load-bearing:
//   1. Drop the Zustand mirror first so any concurrent render cannot
//      observe a post-clear cache miss against the previous user.
//   2. Wipe the React Query cache. Cancels in-flight queries and
//      removes every cached entry — required for cross-account
//      isolation because most query keys are org-/project-scoped, not
//      user-scoped (TanStack Query Discussion #7839).
//
// Callers seed the new session AFTER calling this — `setUser` then
// `setQueryData(authKeys.currentUser(), user)` — and finally navigate.
export function resetSession(queryClient: QueryClient): void {
  useAuthStore.getState().expireSession()
  queryClient.clear()
}
