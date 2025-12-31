'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from './use-auth'
import { ROUTES } from '@/lib/routes'

interface UseAuthGuardOptions {
  redirectTo?: string
  requireEmailVerification?: boolean
  onUnauthorized?: () => void
}

interface UseAuthGuardReturn {
  isAuthorized: boolean
  isLoading: boolean
  isEmailVerified: boolean
}

/**
 * useAuthGuard - Simple authentication verification hook
 * 
 * Only handles basic authentication (login) and email verification.
 * No role-based logic - permissions will be handled by backend.
 */
export function useAuthGuard(options: UseAuthGuardOptions = {}): UseAuthGuardReturn {
  const {
    redirectTo = ROUTES.SIGNIN,
    requireEmailVerification = false,
    onUnauthorized,
  } = options

  const { 
    user, 
    isLoading
  } = useAuth()
  
  const router = useRouter()
  
  // User is authenticated if we have a user object
  const isAuthenticated = !!user

  // Check email verification status
  const isEmailVerified = !requireEmailVerification || (user?.isEmailVerified ?? false)

  // Determine if user is authorized (authenticated and optionally email verified)
  const isAuthorized = isAuthenticated && isEmailVerified

  useEffect(() => {
    // Don't redirect while loading
    if (isLoading) {
      console.log('[AuthGuard] Still loading, not redirecting')
      return
    }

    console.log('[AuthGuard] Auth check:', { isAuthenticated, isLoading, isEmailVerified })

    // Handle unauthenticated users
    if (!isAuthenticated) {
      console.log('[AuthGuard] User not authenticated, redirecting to signin')
      if (onUnauthorized) {
        onUnauthorized()
      } else {
        const currentPath = window.location.pathname
        const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(currentPath)}`
        console.log('[AuthGuard] Redirecting to:', redirectUrl)
        router.push(redirectUrl)
      }
      return
    }

    // Handle unverified email
    if (requireEmailVerification && !isEmailVerified) {
      console.log('[AuthGuard] Email verification required, redirecting')
      router.push(ROUTES.VERIFY_EMAIL)
      return
    }
  }, [
    isLoading,
    isAuthenticated,
    isEmailVerified,
    redirectTo,
    onUnauthorized,
    router,
    requireEmailVerification
  ])

  return {
    isAuthorized,
    isLoading,
    isEmailVerified,
  }
}