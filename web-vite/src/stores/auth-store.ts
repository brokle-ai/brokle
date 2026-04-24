import { create } from 'zustand'

// Minimal in-memory mirror of the backend-owned session state.
//
// Invariants:
//   - The backend is the authority for authentication (CLAUDE.md gotcha
//     #22). This store is the presence-only hint that lets the router
//     guard decide redirects BEFORE the network round-trip.
//   - Never decode JWTs here. Never cache tokens here. Cookies are
//     httpOnly; the client reads `document.cookie` only for the
//     non-httpOnly `csrf_token`.
//   - `user` is whatever the latest `/v1/users/me` response returned.
//     On 401 the auth middleware clears it via `expireSession()`.

export interface SessionUser {
  id: string
  email: string
  first_name: string
  last_name: string
  default_organization_id?: string
  is_email_verified: boolean
  created_at: string
}

interface AuthState {
  user: SessionUser | null
  isAuthenticated: boolean
  setUser: (user: SessionUser | null) => void
  expireSession: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  setUser: (user) => set({ user, isAuthenticated: user !== null }),
  expireSession: () => set({ user: null, isAuthenticated: false }),
}))

// Frozen snapshot for TanStack Router context. Reading the store from
// inside `beforeLoad` is safe but couples the router to Zustand; the
// snapshot indirection keeps `beforeLoad` a pure function of its
// inputs.
export function getAuthSnapshot(): Pick<AuthState, 'user' | 'isAuthenticated'> {
  const { user, isAuthenticated } = useAuthStore.getState()
  return { user, isAuthenticated }
}
