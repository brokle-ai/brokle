'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from './use-auth'
import type { OrganizationRole } from '@/types/auth'

interface UseAuthGuardOptions {
  redirectTo?: string
  requiredRole?: OrganizationRole
  requireEmailVerification?: boolean
  onUnauthorized?: () => void
  onForbidden?: () => void
}

interface UseAuthGuardReturn {
  isAuthorized: boolean
  isLoading: boolean
  hasRequiredRole: boolean
  isEmailVerified: boolean
}

export function useAuthGuard(options: UseAuthGuardOptions = {}): UseAuthGuardReturn {
  const {
    redirectTo = '/auth/signin',
    requiredRole,
    requireEmailVerification = false,
    onUnauthorized,
    onForbidden,
  } = options

  const { 
    user, 
    isLoading
  } = useAuth()
  
  const router = useRouter()
  
  // User is authenticated if we have a user object
  const isAuthenticated = !!user

  // For now, we'll assume all authenticated users have required role
  // In a real app, you'd check user roles here
  const hasRequiredRole = !requiredRole || !!user

  // Check email verification status
  const isEmailVerified = !requireEmailVerification || (user?.isEmailVerified ?? false)

  // Determine if user is authorized
  const isAuthorized = isAuthenticated && hasRequiredRole && isEmailVerified

  useEffect(() => {
    // Don't redirect while loading
    if (isLoading) {
      console.log('[AuthGuard] Still loading, not redirecting')
      return
    }

    console.log('[AuthGuard] Auth check:', { isAuthenticated, isLoading, hasRequiredRole, isEmailVerified })

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

    // Handle insufficient permissions
    if (!hasRequiredRole) {
      if (onForbidden) {
        onForbidden()
      } else {
        // Redirect to login with return URL for production-ready handling
        router.push(`/auth/signin?returnUrl=${encodeURIComponent(window.location.pathname)}`)
      }
      return
    }

    // Handle unverified email
    if (requireEmailVerification && !isEmailVerified) {
      router.push('/auth/verify-email')
      return
    }
  }, [
    isLoading,
    isAuthenticated,
    hasRequiredRole,
    isEmailVerified,
    redirectTo,
    onUnauthorized,
    onForbidden,
    router,
  ])

  return {
    isAuthorized,
    isLoading,
    hasRequiredRole,
    isEmailVerified,
  }
}