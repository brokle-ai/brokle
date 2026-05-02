import type { QueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth-store'

// Two primitives for crossing the auth boundary, picked by call site:
//
//   resetSession(qc)        — full reset INCLUDING queryClient.clear().
//                             Use from sites where observers are about
//                             to unmount (login success, logout, route-
//                             guard catch path → throw redirect).
//
//   notifySessionEnded()    — auth store + registered store callbacks,
//                             but NO queryClient.clear(). Use from
//                             observer hooks (`useCurrentUser`) that
//                             fire on AuthenticationError. qc.clear()
//                             would remove the observer's active query,
//                             causing React Query to immediately
//                             recreate it and re-fire the failed
//                             request in a loop on routes that stay
//                             mounted while unauthenticated (e.g.
//                             /accept-invite).
//
// Both invoke the registered Zustand store reset callbacks (e.g.
// `usePlaygroundStore.clearAll`) so account-scoped non-query state is
// wiped uniformly. Preferences-only stores (UI theme/font/sidebar)
// MUST NOT register so prefs survive logout.
//
// Pattern: https://zustand.docs.pmnd.rs/guides/how-to-reset-state
const sessionResetCallbacks = new Set<() => void>()

export function registerSessionReset(callback: () => void): void {
  sessionResetCallbacks.add(callback)
}

function runRegisteredCallbacks(): void {
  for (const cb of sessionResetCallbacks) {
    try {
      cb()
    } catch (err) {
      // Per-store reset failure must not block the auth boundary;
      // every other registered reset still runs. Log for diagnostics.
      console.error('[session] reset callback failed', err)
    }
  }
}

export function notifySessionEnded(): void {
  useAuthStore.getState().expireSession()
  runRegisteredCallbacks()
}

export function resetSession(queryClient: QueryClient): void {
  useAuthStore.getState().expireSession()
  queryClient.clear()
  runRegisteredCallbacks()
}
