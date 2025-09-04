'use client'

import React, { ReactNode } from 'react'
import { useAuthGuard } from '@/hooks/auth/use-auth-guard'
import { LoadingSpinner } from './loading-spinner'
import { UnauthorizedFallback } from './unauthorized-fallback'
import { ForbiddenFallback } from './forbidden-fallback'
import type { OrganizationRole } from '@/types/auth'

interface AuthGuardProps {
  children: ReactNode
  requiredRole?: OrganizationRole
  requireEmailVerification?: boolean
  fallback?: ReactNode
  loadingFallback?: ReactNode
  unauthorizedFallback?: ReactNode
  forbiddenFallback?: ReactNode
  redirectTo?: string
  onUnauthorized?: () => void
  onForbidden?: () => void
}

export function AuthGuard({
  children,
  requiredRole,
  requireEmailVerification = false,
  fallback,
  loadingFallback,
  unauthorizedFallback,
  forbiddenFallback,
  redirectTo,
  onUnauthorized,
  onForbidden,
}: AuthGuardProps) {
  const {
    isAuthorized,
    isLoading,
    hasRequiredRole,
    isEmailVerified,
  } = useAuthGuard({
    requiredRole,
    requireEmailVerification,
    redirectTo,
    onUnauthorized,
    onForbidden,
  })

  // Show loading state
  if (isLoading) {
    if (loadingFallback) {
      return <>{loadingFallback}</>
    }
    return <LoadingSpinner />
  }

  // Show unauthorized state (not logged in)
  if (!isLoading && !isAuthorized) {
    // If we have specific fallbacks for different failure reasons
    if (!hasRequiredRole && forbiddenFallback) {
      return <>{forbiddenFallback}</>
    }

    if (!isEmailVerified && unauthorizedFallback) {
      return <>{unauthorizedFallback}</>
    }

    // General fallback
    if (fallback) {
      return <>{fallback}</>
    }

    // Default fallbacks based on the specific issue
    if (!hasRequiredRole) {
      return <ForbiddenFallback requiredRole={requiredRole} />
    }

    return <UnauthorizedFallback />
  }

  // User is authorized, render children
  return <>{children}</>
}

// Specialized guard components for common use cases
export function AdminGuard({ children, ...props }: Omit<AuthGuardProps, 'requiredRole'>) {
  return (
    <AuthGuard {...props} requiredRole="admin">
      {children}
    </AuthGuard>
  )
}

export function OwnerGuard({ children, ...props }: Omit<AuthGuardProps, 'requiredRole'>) {
  return (
    <AuthGuard {...props} requiredRole="owner">
      {children}
    </AuthGuard>
  )
}

export function DeveloperGuard({ children, ...props }: Omit<AuthGuardProps, 'requiredRole'>) {
  return (
    <AuthGuard {...props} requiredRole="developer">
      {children}
    </AuthGuard>
  )
}

export function VerifiedGuard({ children, ...props }: Omit<AuthGuardProps, 'requireEmailVerification'>) {
  return (
    <AuthGuard {...props} requireEmailVerification>
      {children}
    </AuthGuard>
  )
}