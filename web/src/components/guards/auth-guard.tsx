'use client'

import React, { ReactNode } from 'react'
import { useAuthGuard } from '@/hooks/auth/use-auth-guard'
import { LoadingSpinner } from './loading-spinner'
import { UnauthorizedFallback } from './unauthorized-fallback'

interface AuthGuardProps {
  children: ReactNode
  requireEmailVerification?: boolean
  fallback?: ReactNode
  loadingFallback?: ReactNode
  unauthorizedFallback?: ReactNode
  redirectTo?: string
  onUnauthorized?: () => void
}

/**
 * AuthGuard - Simple authentication verification component
 * 
 * Only checks if user is logged in and optionally email verified.
 * No role-based access control - that will be handled by backend permissions.
 */
export function AuthGuard({
  children,
  requireEmailVerification = false,
  fallback,
  loadingFallback,
  unauthorizedFallback,
  redirectTo,
  onUnauthorized,
}: AuthGuardProps) {
  const {
    isAuthorized,
    isLoading,
    isEmailVerified,
  } = useAuthGuard({
    requireEmailVerification,
    redirectTo,
    onUnauthorized,
  })

  // Show loading state
  if (isLoading) {
    if (loadingFallback) {
      return <>{loadingFallback}</>
    }
    return <LoadingSpinner />
  }

  // Show unauthorized state (not logged in or email not verified)
  if (!isLoading && !isAuthorized) {
    // Check if it's email verification issue specifically
    if (requireEmailVerification && !isEmailVerified && unauthorizedFallback) {
      return <>{unauthorizedFallback}</>
    }

    // General fallback
    if (fallback) {
      return <>{fallback}</>
    }

    // Default unauthorized fallback
    return <UnauthorizedFallback />
  }

  // User is authorized, render children
  return <>{children}</>
}

/**
 * VerifiedGuard - Convenience component for email verification requirement
 */
export function VerifiedGuard({ children, ...props }: Omit<AuthGuardProps, 'requireEmailVerification'>) {
  return (
    <AuthGuard {...props} requireEmailVerification>
      {children}
    </AuthGuard>
  )
}