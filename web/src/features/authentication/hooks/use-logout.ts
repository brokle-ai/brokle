'use client'

import { useState } from 'react'
import { useAuthStore } from '../stores/auth-store'

/**
 * Hook for logout functionality (lightweight version for internal use)
 *
 * Note: For full logout with events, cache clearing, and redirect,
 * use useLogoutMutation from use-auth-queries.ts instead.
 */
export function useLogout() {
  const [isLoading, setIsLoading] = useState(false)
  const logout = useAuthStore(state => state.logout)

  const handleLogout = async () => {
    setIsLoading(true)
    try {
      await logout()
    } finally {
      setIsLoading(false)
    }
  }

  return {
    logout: handleLogout,
    isLoading,
  }
}
