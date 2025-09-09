'use client'

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { getTokenManager } from '@/lib/auth/token-manager'
import { getSessionSync } from '@/lib/auth/session-sync'
import { SecureStorage } from '@/lib/auth/storage'
import { AUTH_CONSTANTS } from '@/lib/auth/constants'
import { 
  getCurrentUser,
  getCurrentOrganization,
  login as apiLogin,
  signup as apiSignup,
  logout as apiLogout,
  updateProfile as apiUpdateProfile,
  changePassword as apiChangePassword
} from '@/lib/api'
import type { 
  User, 
  Organization, 
  LoginCredentials, 
  SignUpCredentials,
  AuthResponse 
} from '@/types/auth'
import { BrokleAPIError as APIError } from '@/lib/api/core/types'

export interface AuthContextValue {
  // State
  user: User | null
  organization: Organization | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  login: (credentials: LoginCredentials) => Promise<AuthResponse>
  signup: (credentials: SignUpCredentials) => Promise<AuthResponse>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
  clearError: () => void
  
  // User management
  updateUser: (data: Partial<User>) => Promise<User>
  changePassword: (currentPassword: string, newPassword: string) => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
  serverUser?: User | null
}

export function AuthProvider({ children, serverUser }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(serverUser || null)
  const [organization, setOrganization] = useState<Organization | null>(null)
  const [isLoading, setIsLoading] = useState(!serverUser) // If we have server user, we're not loading
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(!!serverUser)

  const router = useRouter()
  const tokenManager = getTokenManager()
  const sessionSync = getSessionSync()

  // Initialize auth state on mount
  useEffect(() => {
    // Skip client-side initialization if we already have server state
    if (serverUser) {
      // We already have server-validated user, just need to set up token manager
      const storedAccessToken = SecureStorage.getAccessToken()
      const storedExpiresAt = SecureStorage.getExpiresAt()
      const storedRefreshToken = SecureStorage.getRefreshToken()
      
      if (storedAccessToken && storedExpiresAt && storedRefreshToken) {
        tokenManager.setTokens({
          accessToken: storedAccessToken,
          refreshToken: storedRefreshToken,
          expiresIn: Math.floor((storedExpiresAt - Date.now()) / 1000),
          tokenType: 'Bearer'
        })
      }
      
      setIsLoading(false)
      return
    }
    
    // No server state, run full client initialization
    initializeAuth()
  }, [serverUser])

  // Setup session sync listeners
  useEffect(() => {
    const unsubscribeLogin = sessionSync.on('LOGIN', (event) => {
      if (event.payload?.user) {
        setUser(event.payload.user)
        setIsAuthenticated(true)
      }
    })

    const unsubscribeLogout = sessionSync.on('LOGOUT', () => {
      setUser(null)
      setOrganization(null)
      setIsAuthenticated(false)
    })

    const unsubscribeSessionExpired = sessionSync.on('SESSION_EXPIRED', () => {
      handleSessionExpired()
    })

    const unsubscribeUserUpdate = sessionSync.on('USER_UPDATED', (event) => {
      if (event.payload?.user) {
        setUser(event.payload.user)
      }
    })

    return () => {
      unsubscribeLogin()
      unsubscribeLogout()
      unsubscribeSessionExpired()
      unsubscribeUserUpdate()
    }
  }, [])

  /**
   * Initialize authentication state on client-side
   * 
   * This function handles the race condition where TokenManager might not have loaded
   * tokens from localStorage yet during app initialization, but the tokens exist.
   * 
   * The solution: Check localStorage directly for tokens and reload them into TokenManager
   * if they exist but TokenManager hasn't loaded them yet. This prevents false negatives
   * that would cause unnecessary token clearing.
   */
  const initializeAuth = async () => {
    try {
      setIsLoading(true)
      
      // Check if we have stored tokens and user
      const storedUser = SecureStorage.getUser()
      const storedAccessToken = SecureStorage.getAccessToken()
      const storedExpiresAt = SecureStorage.getExpiresAt()
      const hasToken = tokenManager.isAuthenticated()

      // If we have stored user and stored tokens (even if TokenManager hasn't loaded them yet)
      if (storedUser && storedAccessToken && storedExpiresAt) {
        // Check if stored token is not expired
        if (Date.now() < storedExpiresAt - AUTH_CONSTANTS.TOKEN_EXPIRY_BUFFER) {
          // If TokenManager doesn't have the token yet, reload it
          if (!hasToken) {
            const storedRefreshToken = SecureStorage.getRefreshToken()
            if (storedRefreshToken) {
              tokenManager.setTokens({
                accessToken: storedAccessToken,
                refreshToken: storedRefreshToken,
                expiresIn: Math.floor((storedExpiresAt - Date.now()) / 1000),
                tokenType: 'Bearer'
              })
            }
          }
          
          try {
            const currentUser = await getCurrentUser()
            const currentOrg = await getCurrentOrganization()
            
            setUser(currentUser)
            setOrganization(currentOrg)
            setIsAuthenticated(true)
            
            // Update stored user data if different
            if (JSON.stringify(currentUser) !== JSON.stringify(storedUser)) {
              SecureStorage.setUser(currentUser)
            }
          } catch (error) {
            // Token is invalid or API call failed, clear everything
            await clearAuthState()
          }
        } else {
          // Token is expired, clear auth state
          await clearAuthState()
        }
      } else {
        // No valid auth state
        await clearAuthState()
      }
    } catch (error) {
      // Auth initialization failed, clear state and proceed
      await clearAuthState()
    } finally {
      setIsLoading(false)
    }
  }

  const login = async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      setIsLoading(true)
      setError(null)

      const response = await apiLogin(credentials)
      
      // Update state
      setUser(response.user)
      setOrganization(response.organization)
      setIsAuthenticated(true)
      
      // Store user data
      SecureStorage.setUser(response.user)
      
      // Broadcast to other tabs
      sessionSync.broadcastLogin(response.user)
      
      return response
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
      console.error('[AuthContext] Login failed:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const signup = async (credentials: SignUpCredentials): Promise<AuthResponse> => {
    try {
      setIsLoading(true)
      setError(null)

      const response = await apiSignup(credentials)
      
      // Update state
      setUser(response.user)
      setOrganization(response.organization)
      setIsAuthenticated(true)
      
      // Store user data
      SecureStorage.setUser(response.user)
      
      // Broadcast to other tabs
      sessionSync.broadcastLogin(response.user)
      
      return response
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
      console.error('[AuthContext] Signup failed:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const logout = async (): Promise<void> => {
    try {
      setIsLoading(true)
      
      // Call logout API (this will clear tokens via tokenManager)
      await apiLogout()
      
      // Clear local state
      await clearAuthState()
      
      // Broadcast to other tabs
      sessionSync.broadcastLogout()
      
      // Redirect to login
      router.push('/auth/signin')
      
    } catch (error) {
      console.error('[AuthContext] Logout failed:', error)
      // Clear local state even if API call fails
      await clearAuthState()
      router.push('/auth/signin')
    } finally {
      setIsLoading(false)
    }
  }

  const refreshToken = async (): Promise<void> => {
    try {
      await tokenManager.refreshAccessToken()
    } catch (error) {
      console.error('[AuthContext] Token refresh failed:', error)
      await handleSessionExpired()
      throw error
    }
  }

  const updateUser = async (data: Partial<User>): Promise<User> => {
    try {
      const updatedUser = await apiUpdateProfile(data)
      
      setUser(updatedUser)
      SecureStorage.setUser(updatedUser)
      
      // Broadcast to other tabs
      sessionSync.broadcastUserUpdate(updatedUser)
      
      return updatedUser
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
      console.error('[AuthContext] User update failed:', error)
      throw error
    }
  }

  const changePassword = async (
    currentPassword: string, 
    newPassword: string
  ): Promise<void> => {
    try {
      await apiChangePassword(currentPassword, newPassword)
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
      console.error('[AuthContext] Password change failed:', error)
      throw error
    }
  }

  const clearError = () => {
    setError(null)
  }

  // Private helper methods
  const clearAuthState = async (): Promise<void> => {
    setUser(null)
    setOrganization(null)
    setIsAuthenticated(false)
    setError(null)
    tokenManager.clearTokens()
    SecureStorage.clearAllAuthData()
  }

  const handleSessionExpired = async (): Promise<void> => {
    await clearAuthState()
    setError('Your session has expired. Please login again.')
    router.push('/auth/signin')
  }

  const getErrorMessage = (error: unknown): string => {
    if (error instanceof APIError) {
      return error.message
    }
    if (error instanceof Error) {
      return error.message
    }
    return 'An unexpected error occurred'
  }

  const value: AuthContextValue = {
    // State
    user,
    organization,
    isAuthenticated,
    isLoading,
    error,

    // Actions
    login,
    signup,
    logout,
    refreshToken,
    clearError,

    // User management
    updateUser,
    changePassword,
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

// Hook to use the auth context
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

// Export for convenience
export { AuthContext }