import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { User, Organization, Project, ApiKey, LoginCredentials, AuthResponse } from '../types'
import * as authApi from '../api/auth-api'
import { BrokleAPIError } from '@/lib/api/core/types'
import { BrokleAPIClient } from '@/lib/api/core/client'

/**
 * Payload accepted by `hydrate()` — the Phase 2 bootstrap path that
 * receives server-fetched user data and skips the legacy
 * client-side `initializeAuth()` round-trip.
 *
 * `expiresAt`/`expiresIn` are optional: server-side bootstrap from
 * the rich /v1/users/me endpoint does not include token expiry
 * metadata. The proactive refresh timer starts after the first
 * reactive refresh completes (which sets these fields). Until then
 * refresh is reactive on 401 — no functional regression, just one
 * extra round-trip per token-expiry.
 */
export interface HydratePayload {
  user: User
  organization?: Organization | null
  expiresAt?: number | null
  expiresIn?: number | null
}

// Create dedicated client instance for auth operations
const client = new BrokleAPIClient('/api')

export interface AuthState {
  // User state
  user: User | null
  organization: Organization | null
  currentProject: Project | null

  // Token metadata (NOT the actual tokens - those are in httpOnly cookies)
  expiresAt: number | null  // Unix timestamp in milliseconds
  expiresIn: number | null  // Duration in milliseconds

  // UI state
  isAuthenticated: boolean
  isLoading: boolean
  isRefreshing: boolean
  error: string | null

  // Auto-refresh timer (cross-platform type)
  refreshTimerId: ReturnType<typeof setTimeout> | null
  refreshPromise: Promise<void> | null

  // API keys
  apiKeys: ApiKey[]

  // Actions
  login: (credentials: LoginCredentials) => Promise<AuthResponse>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
  initializeAuth: () => Promise<void>
  hydrate: (payload: HydratePayload) => void
  startRefreshTimer: () => void
  stopRefreshTimer: () => void
  clearAuth: () => void

  // Legacy actions (for compatibility)
  setUser: (user: User | null) => void
  setOrganization: (organization: Organization | null) => void
  setCurrentProject: (project: Project | null) => void
  setApiKeys: (apiKeys: ApiKey[]) => void
  setLoading: (loading: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  devtools(
    (set, get) => ({
      // Initial state
      user: null,
      organization: null,
      currentProject: null,
      expiresAt: null,
      expiresIn: null,
      isAuthenticated: false,
      isLoading: false,
      isRefreshing: false,
      error: null,
      refreshTimerId: null,
      refreshPromise: null,
      apiKeys: [],

      // Login action
      login: async (credentials) => {
        if (process.env.NODE_ENV === 'development') {
          console.debug('[AuthStore] Login started')
        }
        set({ isLoading: true, error: null })

        try {
          // Call auth service (sets httpOnly cookies, returns metadata)
          if (process.env.NODE_ENV === 'development') {
            console.debug('[AuthStore] Calling authApi.login...')
          }
          const response = await authApi.login(credentials)

          if (process.env.NODE_ENV === 'development') {
            console.debug('[AuthStore] authApi.login returned:', {
              hasResponse: !!response,
              hasUser: !!response?.user,
              hasOrg: !!response?.organization,
              hasExpiresAt: !!response?.expiresAt,
              response
            })
          }

          // Defensive check
          if (!response) {
            console.error('[AuthStore] Response is undefined')
            throw new Error('Login response is undefined')
          }

          if (!response.user) {
            console.error('[AuthStore] Response missing user:', response)
            throw new Error('Login response missing user data')
          }

          if (!response.organization) {
            console.error('[AuthStore] Response missing organization:', response)
            throw new Error('Login response missing organization data')
          }

          console.debug('[AuthStore] Setting auth state...')
          set({
            user: response.user,
            organization: response.organization,
            expiresAt: response.expiresAt,  // Milliseconds
            expiresIn: response.expiresIn,  // Milliseconds
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          console.debug('[AuthStore] Auth state set successfully')

          // Start auto-refresh timer
          get().startRefreshTimer()
          console.debug('[AuthStore] Login complete')

          // Return the response for mutation hooks
          return response
        } catch (error) {
          console.error('[AuthStore] Login error:', error)
          const errorMessage = error instanceof BrokleAPIError
            ? error.message
            : 'Login failed'

          set({
            isLoading: false,
            error: errorMessage
          })
          throw error
        }
      },

      // Logout action
      logout: async () => {
        try {
          // Call backend to clear cookies
          await authApi.logout()
        } catch (error) {
          console.error('[Auth] Logout error:', error)
        } finally {
          get().clearAuth()

          // Signal other tabs
          if (typeof window !== 'undefined') {
            localStorage.setItem('logout_signal', Date.now().toString())

            // Delay removal to ensure other tabs detect the signal
            setTimeout(() => {
              localStorage.removeItem('logout_signal')
            }, 100)
          }
        }
      },

      // Token refresh with shared promise pattern
      refreshToken: async () => {
        // Return existing promise if refresh in progress
        const existing = get().refreshPromise
        if (get().isRefreshing && existing) {
          return existing
        }

        // Create new refresh promise
        const refreshPromise = (async () => {
          set({ isRefreshing: true, error: null })

          try {
            // Call refresh endpoint (cookies sent automatically)
            // skipRefreshInterceptor prevents recursive refresh loops
            const response = await client.post<{
              expires_at: number
              expires_in: number
            }>(
              '/v1/auth/refresh',
              {},
              { skipRefreshInterceptor: true }
            )

            set({
              expiresAt: response.expires_at,
              expiresIn: response.expires_in,
              isRefreshing: false,
              refreshPromise: null,
            })

            // Restart timer with new expiry
            get().startRefreshTimer()

            if (process.env.NODE_ENV === 'development') {
              console.debug('[Auth] Token refreshed successfully')
            }
          } catch (error) {
            set({
              isRefreshing: false,
              refreshPromise: null
            })

            // Any 401 from /v1/auth/refresh means the refresh path is
            // no longer viable — cookie missing, invalid, expired, or
            // revoked. Cleanup is identical regardless of which: clear
            // local auth state and signal the providers-level handler
            // to redirect to /signin.
            //
            // Previously this only fired for `error.code === 'REFRESH_EXPIRED'`,
            // which missed the missing-cookie case (the most common
            // first-load scenario) and left the dashboard wedged on
            // "Loading workspace…" because no event was dispatched.
            const isAuthFailure =
              error instanceof BrokleAPIError && error.statusCode === 401

            if (isAuthFailure) {
              if (process.env.NODE_ENV === 'development') {
                console.debug('[Auth] Refresh failed, clearing session', {
                  code: error.code,
                  statusCode: error.statusCode,
                })
              }
              get().clearAuth()

              if (typeof window !== 'undefined') {
                window.dispatchEvent(new CustomEvent('auth:session-expired'))
              }
            }

            throw error
          }
        })()

        set({ refreshPromise })
        return refreshPromise
      },

      // Start auto-refresh timer (1 minute before expiry)
      startRefreshTimer: () => {
        const { expiresAt, refreshTimerId } = get()

        // Clear existing timer
        if (refreshTimerId) {
          clearTimeout(refreshTimerId)
        }

        if (!expiresAt) return

        // Calculate time until refresh (all in milliseconds)
        const now = Date.now()
        const refreshTime = expiresAt - 60000  // 60 seconds before expiry
        const timeUntilRefresh = refreshTime - now

        if (process.env.NODE_ENV === 'development') {
          console.debug('[Auth] Refresh timer:', {
            now,
            expiresAt,
            timeUntilRefresh,
            refreshIn: `${Math.floor(timeUntilRefresh / 1000)}s`,
          })
        }

        if (timeUntilRefresh > 0) {
          const timerId = setTimeout(() => {
            if (process.env.NODE_ENV === 'development') {
              console.debug('[Auth] Auto-refreshing token...')
            }
            get().refreshToken().catch(console.error)
          }, timeUntilRefresh)

          set({ refreshTimerId: timerId })
        } else {
          // Token already expired or expires very soon - refresh immediately
          if (process.env.NODE_ENV === 'development') {
            console.debug('[Auth] Token expired, refreshing immediately')
          }
          get().refreshToken().catch(console.error)
        }
      },

      // Stop auto-refresh timer
      stopRefreshTimer: () => {
        const { refreshTimerId } = get()
        if (refreshTimerId) {
          clearTimeout(refreshTimerId)
          set({ refreshTimerId: null })
        }
      },

      // Hydrate from server-fetched bootstrap data (Phase 2 path).
      //
      // Called once per cold load from DashboardLayoutClient with the
      // payload the Server Component DAL fetched. Idempotent:
      // re-calling with the same user is a no-op (the lazy useState
      // initializer in DashboardLayoutClient guarantees this anyway,
      // but the guard makes the function safe to call from elsewhere).
      //
      // Replaces the legacy initializeAuth() round-trip. If
      // expiresAt is provided, starts the proactive refresh timer;
      // otherwise the timer kicks in after the first reactive refresh.
      hydrate: (payload) => {
        const current = get()
        if (current.isAuthenticated && current.user?.id === payload.user.id) {
          // Already hydrated for this user (StrictMode double-render,
          // hot reload, etc.). No-op.
          return
        }

        set({
          user: payload.user,
          organization: payload.organization ?? null,
          expiresAt: payload.expiresAt ?? null,
          expiresIn: payload.expiresIn ?? null,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        })

        if (payload.expiresAt) {
          get().startRefreshTimer()
        }
      },

      // Clear all auth state
      clearAuth: () => {
        get().stopRefreshTimer()

        set({
          user: null,
          organization: null,
          currentProject: null,
          expiresAt: null,
          expiresIn: null,
          isAuthenticated: false,
          isLoading: false,
          refreshTimerId: null,
          refreshPromise: null,
          error: null,
        })

        // Clear persisted context data (privacy - prevent email leak on shared computers)
        if (typeof window !== 'undefined') {
          localStorage.removeItem('brokle_last_context')
        }
      },

      // Initialize auth on app load
      initializeAuth: async () => {
        set({ isLoading: true })

        try {
          // Call /me endpoint (cookies sent automatically)
          const response = await client.get<{
            user: any
            expires_at: number
            expires_in: number
          }>('/v1/auth/me')

          // Map user data
          const user: User = {
            id: response.user.id,
            email: response.user.email,
            firstName: response.user.first_name,
            lastName: response.user.last_name,
            name: `${response.user.first_name} ${response.user.last_name}`.trim(),
            role: 'user',
            organizationId: '',
            defaultOrganizationId: response.user.default_organization_id,
            projects: [],
            createdAt: response.user.created_at,
            updatedAt: response.user.updated_at,
            isEmailVerified: response.user.is_email_verified,
          }

          // Fetch organization (cookies sent automatically). Canonical
          // `{data, pagination}` list shape per CLAUDE.md gotcha #23.
          const orgResponse = await client.getPaginated<{
            id: string
            name: string
            subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
            created_at: string
            updated_at: string
          }>('/v1/organizations')

          const firstOrg = orgResponse.data.length > 0 ? orgResponse.data[0] : null
          let organization: Organization | null = null

          if (firstOrg) {
            organization = {
              id: firstOrg.id,
              name: firstOrg.name,
              plan: firstOrg.subscription_plan,
              members: [],
              apiKeys: [],
              usage: {
                traces_this_month: 0,
                observed_cost_this_month: 0,
                models_observed: 0,
              },
              createdAt: firstOrg.created_at,
              updatedAt: firstOrg.updated_at,
            }
          }

          set({
            user,
            organization,
            expiresAt: response.expires_at,
            expiresIn: response.expires_in,
            isAuthenticated: true,
            isLoading: false,
          })

          // Start auto-refresh timer
          get().startRefreshTimer()
        } catch (error) {
          // Distinguish error types for better UX
          set({
            isLoading: false,
            isAuthenticated: false
          })

          if (error instanceof BrokleAPIError) {
            if (error.statusCode === 401) {
              // Auth error - expected when not logged in
              if (process.env.NODE_ENV === 'development') {
                console.debug('[Auth] Not authenticated (expected)')
              }
            } else if (error.isNetworkError()) {
              // Network error - user should be aware
              set({ error: 'Network error - please check your connection' })
              console.error('[Auth] Network error during initialization:', error)
            } else {
              // Other unexpected error
              set({ error: 'Authentication system error' })
              console.error('[Auth] Unexpected error during initialization:', error)
            }
          } else {
            // Non-API error
            console.error('[Auth] Unknown error during initialization:', error)
          }
        }
      },

      // Legacy actions (for compatibility with existing components)
      setUser: (user) => set({ user }),
      setOrganization: (organization) => set({ organization }),
      setCurrentProject: (project) => set({ currentProject: project }),
      setApiKeys: (apiKeys) => set({ apiKeys }),
      setLoading: (loading) => set({ isLoading: loading }),
    }),
    {
      name: 'auth-store',
    }
  )
)
