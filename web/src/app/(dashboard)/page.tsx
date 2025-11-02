'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { OrganizationSelector } from '@/components/organization/organization-selector'
import { PageLoader } from '@/components/shared/loading'

export default function RootPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()

  // Derive authentication state from user presence
  const isAuthenticated = !!user

  useEffect(() => {
    if (authLoading) return

    if (!isAuthenticated) {
      router.push('/auth/signin')
      return
    }

    // No onboarding check needed - completed during signup
  }, [authLoading, isAuthenticated, user, router])

  if (authLoading) {
    return <PageLoader message="Loading your workspace..." />
  }

  if (!isAuthenticated) {
    return null // Will redirect to signin
  }

  return (
    <div className="flex h-screen items-center justify-center p-6">
      <div className="w-full max-w-6xl">
        <OrganizationSelector />
      </div>
    </div>
  )
}
