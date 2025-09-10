'use client'

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { 
  login as apiLogin,
  signup as apiSignup,
  logout as apiLogout,
  getCurrentUser,
  updateProfile as apiUpdateProfile,
  changePassword as apiChangePassword
} from '@/lib/api'
import type { 
  User, 
  LoginCredentials, 
  SignUpCredentials,
  AuthResponse 
} from '@/types/auth'
import { BrokleAPIError as APIError } from '@/lib/api/core/types'

export interface AuthContextValue {
  // State
  user: User | null
  accessToken: string | null
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

// Simple token storage utilities
const TokenStorage = {
  setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => {
    if (typeof window === 'undefined') return
    
    // Store in memory (React state) and localStorage backup
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
    localStorage.setItem('expires_at', String(Date.now() + expiresIn * 1000))
    
    // Set cookie for middleware
    document.cookie = `access_token=${accessToken}; path=/; max-age=${expiresIn}; SameSite=Strict`
  },
  
  getAccessToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('access_token')
  },
  
  getRefreshToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('refresh_token')
  },
  
  isTokenExpired: (): boolean => {
    if (typeof window === 'undefined') return true
    const expiresAt = localStorage.getItem('expires_at')
    if (!expiresAt) return true
    return Date.now() >= parseInt(expiresAt) - 30000 // 30s buffer
  },
  
  clear: () => {
    if (typeof window === 'undefined') return
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('expires_at')
    localStorage.removeItem('user')
    
    // Clear cookie
    document.cookie = 'access_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT'
  },
  
  setUser: (user: User) => {
    if (typeof window === 'undefined') return
    localStorage.setItem('user', JSON.stringify(user))
  },
  
  getUser: (): User | null => {
    if (typeof window === 'undefined') return null
    const userStr = localStorage.getItem('user')
    try {
      return userStr ? JSON.parse(userStr) : null
    } catch {
      return null
    }
  }
}

export function AuthProvider({ children, serverUser }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(serverUser || null)
  const [accessToken, setAccessToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(!serverUser)
  const [error, setError] = useState<string | null>(null)

  const router = useRouter()

  // Initialize auth state on mount
  useEffect(() => {
    if (serverUser) {
      // Server-provided user, just load token from storage
      const storedToken = TokenStorage.getAccessToken()
      if (storedToken && !TokenStorage.isTokenExpired()) {
        setAccessToken(storedToken)
      }
      setIsLoading(false)
      return
    }
    
    initializeAuth()
  }, [serverUser])

  const initializeAuth = async () => {
    try {
      setIsLoading(true)
      
      const storedToken = TokenStorage.getAccessToken()
      const storedUser = TokenStorage.getUser()
      
      if (storedToken && storedUser && !TokenStorage.isTokenExpired()) {
        // Token is valid, verify with server
        try {
          const currentUser = await getCurrentUser()
          setUser(currentUser)
          setAccessToken(storedToken)
          TokenStorage.setUser(currentUser)
        } catch {
          // Token invalid, clear storage
          await clearAuthState()
        }
      } else if (TokenStorage.isTokenExpired() && TokenStorage.getRefreshToken()) {
        // Token expired, try refresh
        try {
          await refreshToken()
        } catch {
          await clearAuthState()
        }
      } else {
        // No valid auth state
        await clearAuthState()
      }
    } catch (error) {
      console.error('[AuthContext] Init failed:', error)
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
      
      // Store tokens and user
      TokenStorage.setTokens(response.accessToken, response.refreshToken, response.expiresIn)
      TokenStorage.setUser(response.user)
      
      // Update state
      setUser(response.user)
      setAccessToken(response.accessToken)
      
      return response
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
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
      
      // Store tokens and user
      TokenStorage.setTokens(response.accessToken, response.refreshToken, response.expiresIn)
      TokenStorage.setUser(response.user)
      
      // Update state
      setUser(response.user)
      setAccessToken(response.accessToken)
      
      return response
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const logout = async (): Promise<void> => {
    try {
      setIsLoading(true)
      
      // Call logout API
      await apiLogout()
      
      // Clear state
      await clearAuthState()
      
      // Redirect to login
      router.push('/auth/signin')
      
    } catch (error) {
      console.error('[AuthContext] Logout failed:', error)
      // Clear state even if API fails
      await clearAuthState()
      router.push('/auth/signin')
    } finally {
      setIsLoading(false)
    }
  }

  const refreshToken = async (): Promise<void> => {
    // Simplified refresh - for now just clear and redirect
    // In production, implement proper refresh token flow
    await clearAuthState()
    router.push('/auth/signin')
  }

  const updateUser = async (data: Partial<User>): Promise<User> => {
    try {
      const updatedUser = await apiUpdateProfile(data)
      
      setUser(updatedUser)
      TokenStorage.setUser(updatedUser)
      
      return updatedUser
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setError(errorMessage)
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
      throw error
    }
  }

  const clearError = () => {
    setError(null)
  }

  // Private helper methods
  const clearAuthState = async (): Promise<void> => {
    setUser(null)
    setAccessToken(null)
    setError(null)
    TokenStorage.clear()
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
    accessToken,
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