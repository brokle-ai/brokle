'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { OrganizationSelector } from '@/components/organization/organization-selector'
import { Skeleton } from '@/components/ui/skeleton'
import { getCurrentUser } from '@/lib/api'

export default function RootPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const [checkingOnboarding, setCheckingOnboarding] = useState(true)

  // Derive authentication state from user presence
  const isAuthenticated = !!user

  useEffect(() => {
    if (authLoading) return

    if (!isAuthenticated) {
      router.push('/auth/signin')
      return
    }

    // Check onboarding status first
    const checkOnboardingStatus = async () => {
      try {
        setCheckingOnboarding(true)
        const user = await getCurrentUser()
        
        // If user hasn't completed onboarding, redirect to onboarding
        if (!user.onboardingCompleted) {
          router.push('/onboarding')
          return
        }
      } catch (error) {
        console.error('Error checking onboarding status:', error)
      } finally {
        setCheckingOnboarding(false)
      }
    }

    checkOnboardingStatus()
  }, [authLoading, isAuthenticated, router])

  if (authLoading || checkingOnboarding) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="space-y-4 text-center">
          <Skeleton className="h-8 w-64 mx-auto" />
          <Skeleton className="h-5 w-96 mx-auto" />
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 max-w-6xl">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-48 w-80" />
            ))}
          </div>
        </div>
      </div>
    )
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
