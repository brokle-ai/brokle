'use client'

import { Suspense } from 'react'
import { Loader2 } from 'lucide-react'
import { AuthErrorBoundary } from './auth-error-boundary'

function AuthFormFallback() {
  return (
    <div className="flex items-center justify-center p-8">
      <Loader2 className="h-6 w-6 animate-spin" />
    </div>
  )
}

export function AuthFormWrapper({ children }: { children: React.ReactNode }) {
  return (
    <AuthErrorBoundary
      fallbackTitle="Sign In Error"
      fallbackDescription="Something went wrong while loading the sign-in form. Please try refreshing the page."
    >
      <Suspense fallback={<AuthFormFallback />}>
        {children}
      </Suspense>
    </AuthErrorBoundary>
  )
}