'use client'

import { useEffect, useCallback } from 'react'
import { useAuth } from './use-auth'
import { getTokenManager } from '@/lib/auth/token-manager'
import { AUTH_CONSTANTS } from '@/lib/auth/constants'

interface UseTokenRefreshOptions {
  enabled?: boolean
  onRefreshSuccess?: () => void
  onRefreshError?: (error: Error) => void
}

export function useTokenRefresh(options: UseTokenRefreshOptions = {}) {
  const { enabled = true, onRefreshSuccess, onRefreshError } = options
  const { isAuthenticated, refreshToken } = useAuth()
  const tokenManager = getTokenManager()

  const checkAndRefreshToken = useCallback(async () => {
    if (!isAuthenticated || !enabled) return

    try {
      const timeLeft = tokenManager.getTokenTimeLeft()
      
      // Refresh token if it expires within the threshold
      if (timeLeft <= AUTH_CONSTANTS.TOKEN_REFRESH_THRESHOLD) {
        console.debug('[TokenRefresh] Token expiring soon, refreshing...')
        await refreshToken()
        onRefreshSuccess?.()
      }
    } catch (error) {
      console.error('[TokenRefresh] Auto-refresh failed:', error)
      onRefreshError?.(error as Error)
    }
  }, [isAuthenticated, enabled, refreshToken, onRefreshSuccess, onRefreshError, tokenManager])

  // Set up periodic token refresh check
  useEffect(() => {
    if (!enabled || !isAuthenticated) return

    // Check immediately
    checkAndRefreshToken()

    // Set up interval
    const interval = setInterval(
      checkAndRefreshToken,
      AUTH_CONSTANTS.TOKEN_REFRESH_INTERVAL
    )

    return () => clearInterval(interval)
  }, [enabled, isAuthenticated, checkAndRefreshToken])

  // Also check when tab becomes visible (user returns to tab)
  useEffect(() => {
    if (!enabled || !isAuthenticated) return

    const handleVisibilityChange = () => {
      if (!document.hidden) {
        checkAndRefreshToken()
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [enabled, isAuthenticated, checkAndRefreshToken])

  // Manual refresh function
  const manualRefresh = useCallback(async () => {
    if (!isAuthenticated) return

    try {
      await refreshToken()
      onRefreshSuccess?.()
    } catch (error) {
      onRefreshError?.(error as Error)
      throw error
    }
  }, [isAuthenticated, refreshToken, onRefreshSuccess, onRefreshError])

  return {
    manualRefresh,
    tokenTimeLeft: tokenManager.getTokenTimeLeft(),
    isTokenValid: tokenManager.isAuthenticated(),
  }
}