'use client'

import { Suspense, useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { exchangeLoginSession } from '@/lib/api/services/auth'

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
      setTimeout(() => router.push('/auth/signin'), 2000)
      return
    }

    // Exchange session for tokens
    exchangeLoginSession(session)
      .then((tokenData) => {
        const accessToken = tokenData.access_token
        const refreshToken = tokenData.refresh_token
        const expiresIn = tokenData.expires_in || 900

        if (accessToken && refreshToken) {
          // Store in localStorage
          localStorage.setItem('access_token', accessToken)
          localStorage.setItem('refresh_token', refreshToken)
          localStorage.setItem('expires_at', String(Date.now() + expiresIn * 1000))

          // Set cookie for SSR/middleware
          document.cookie = `access_token=${accessToken}; path=/; max-age=${expiresIn}; SameSite=Strict`

          // Redirect to dashboard
          window.location.href = '/'
        } else {
          setError('Invalid session data. Please try again.')
          setTimeout(() => router.push('/auth/signin'), 2000)
        }
      })
      .catch((err) => {
        console.error('Failed to exchange tokens:', err)
        setError('Failed to complete authentication. Please try again.')
        setTimeout(() => router.push('/auth/signin'), 3000)
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
