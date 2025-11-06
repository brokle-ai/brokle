'use client'

import { useSearchParams } from 'next/navigation'
import { useEffect, useRef } from 'react'
import { toast } from 'sonner'

/**
 * Client component that shows toast notifications on signin page
 * based on query parameters (logout, session expiry, cross-tab sync)
 *
 * Keeps signin page as server component (preserves metadata exports)
 */
export function SignInToastHandler() {
  const searchParams = useSearchParams()
  const hasShownToast = useRef(false)  // Prevent duplicate toasts

  useEffect(() => {
    // Only show toast once
    if (hasShownToast.current) return

    const logoutParam = searchParams.get('logout')
    const sessionParam = searchParams.get('session')

    // Only proceed if we have a param to handle
    if (!logoutParam && !sessionParam) return

    // Mark as shown
    hasShownToast.current = true

    // Show appropriate toast based on redirect reason
    if (logoutParam === 'success') {
      toast.success('Logged out successfully', {
        description: 'You have been securely logged out.',
      })
    } else if (sessionParam === 'ended') {
      toast.info('Session ended', {
        description: 'You have been logged out from another tab.',
      })
    } else if (sessionParam === 'expired') {
      toast.error('Session expired', {
        description: 'Your session has expired. Please log in again.',
      })
    } else if (logoutParam === 'error') {
      toast.warning('Logged out locally', {
        description: 'Session cleared locally.',
      })
    }

    // Clean up logout/session params while preserving others (e.g., ?redirect=/dashboard)
    if (logoutParam || sessionParam) {
      const url = new URL(window.location.href)
      url.searchParams.delete('logout')
      url.searchParams.delete('session')

      // Preserve pathname and remaining search params
      const cleanUrl = url.pathname + url.search
      window.history.replaceState({}, '', cleanUrl)
    }
  }, [searchParams])

  return null  // Side effects only
}
