import { useAuthStore } from '../stores/auth-store'

/**
 * useAuth hook - Simple interface to Zustand auth store
 *
 * Replaces the old AuthContext with Zustand-based state management
 * using httpOnly cookies for security
 */
export function useAuth() {
  const user = useAuthStore(state => state.user)
  const organization = useAuthStore(state => state.organization)
  const isLoading = useAuthStore(state => state.isLoading)
  const isAuthenticated = useAuthStore(state => state.isAuthenticated)
  const error = useAuthStore(state => state.error)
  const login = useAuthStore(state => state.login)
  const logout = useAuthStore(state => state.logout)

  return {
    user,
    organization,
    isLoading,
    isAuthenticated,
    error,
    login,
    logout,
  }
}
