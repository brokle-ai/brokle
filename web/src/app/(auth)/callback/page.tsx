'use client'

import { Suspense, useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { exchangeLoginSession } from '@/features/authentication'
import { ROUTES } from '@/lib/routes'

// OAuth callback page for handling token exchange after OAuth login
function OAuthCallbackContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const session = searchParams.get('session')
    const type = searchParams.get('type') // 'login' (for existing users)

    if (!session) {
      setError('No session found. Please try again.')
      setTimeout(() => router.push(ROUTES.SIGNIN), 2000)
      return
    }

    // Exchange session for tokens (tokens set in httpOnly cookies automatically)
    exchangeLoginSession(session)
      .then(async (response) => {
        // Backend sets httpOnly cookies automatically
        // Response contains: { user, expires_at, expires_in }

        if (response && response.user) {
          // Initialize auth store BEFORE redirecting (prevents loading flash)
          const { useAuthStore } = await import('@/features/authentication')

          // Map user response to User type
          const user = {
            id: response.user.id,
            email: response.user.email,
            firstName: response.user.first_name,
            lastName: response.user.last_name,
            name: `${response.user.first_name} ${response.user.last_name}`.trim(),
            role: 'user',
            organizationId: '',
            defaultOrganizationId: response.user.default_organization_id,
            projects: [],
            createdAt: response.user.created_at,
            updatedAt: response.user.created_at,
            isEmailVerified: response.user.is_email_verified,
          }

          // Set auth state manually
          const store = useAuthStore.getState()
          store.setUser(user)
          store.setLoading(false)

          // Set expiry metadata and start timer
          useAuthStore.setState({
            expiresAt: response.expires_at,
            expiresIn: response.expires_in,
            isAuthenticated: true,
          })

          store.startRefreshTimer()

          // Small delay to ensure state fully persists before redirect
          await new Promise(resolve => setTimeout(resolve, 50))

          // Now redirect (auth state fully settled)
          window.location.href = '/'
        } else {
          setError('Invalid session data. Please try again.')
          setTimeout(() => router.push(ROUTES.SIGNIN), 2000)
        }
      })
      .catch((err) => {
        console.error('Failed to exchange tokens:', err)
        setError('Failed to complete authentication. Please try again.')
        setTimeout(() => router.push(ROUTES.SIGNIN), 3000)
      })
  }, [searchParams, router])

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-4">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <p className="mt-4 text-sm text-muted-foreground">Completing authentication...</p>
    </div>
  )
}

export default function OAuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <OAuthCallbackContent />
    </Suspense>
  )
}
