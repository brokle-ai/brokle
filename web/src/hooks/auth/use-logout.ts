'use client'

import { useState } from 'react'
import { useAuth } from './use-auth'
import { toast } from 'sonner'

interface UseLogoutReturn {
  logout: () => Promise<void>
  isLoading: boolean
}

export function useLogout(): UseLogoutReturn {
  const [isLoading, setIsLoading] = useState(false)
  
  const { logout: authLogout } = useAuth()

  const logout = async (): Promise<void> => {
    setIsLoading(true)

    try {
      await authLogout()
      
      // Show success message
      toast.success('Logged out successfully', {
        description: 'You have been securely logged out.',
      })
    } catch (err: any) {
      // Even if logout API fails, we still want to clear local state
      console.error('Logout error:', err)
      
      toast.warning('Logged out locally', {
        description: 'Session cleared locally. You may need to refresh other tabs.',
      })
    } finally {
      setIsLoading(false)
    }
  }

  return {
    logout,
    isLoading,
  }
}