'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { useOrganization } from '@/context/organization-context'
import { OrganizationSelector } from '@/components/organization/organization-selector'
import { Skeleton } from '@/components/ui/skeleton'
import { api } from '@/lib/api'

export default function RootPage() {
  const router = useRouter()
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const { 
    currentOrganization, 
    currentProject, 
    organizations, 
    isLoading: orgLoading
  } = useOrganization()
  const [checkingOnboarding, setCheckingOnboarding] = useState(true)

  useEffect(() => {
    if (authLoading || orgLoading) return

    if (!isAuthenticated) {
      router.push('/auth/signin')
      return
    }

    // Check onboarding status first
    const checkOnboardingStatus = async () => {
      try {
        setCheckingOnboarding(true)
        const user = await api.auth.getCurrentUser()
        
        // If user hasn't completed onboarding, redirect to onboarding
        if (!user.onboardingCompleted) {
          router.push('/onboarding')
          return
        }
        
        // If onboarding is completed, proceed with existing organization logic
        if (currentOrganization) {
          const url = currentProject 
            ? `/${currentOrganization.slug}/${currentProject.slug}`
            : `/${currentOrganization.slug}`
          
          router.replace(url)
          return
        }

        // If user has organizations but no current one, stay on selector
        // If user has no organizations, the OrganizationSelector will handle showing create flow
      } catch (error) {
        console.error('Error checking onboarding status:', error)
        // On error, proceed with existing logic
      } finally {
        setCheckingOnboarding(false)
      }
    }

    checkOnboardingStatus()
  }, [
    authLoading,
    orgLoading,
    isAuthenticated,
    currentOrganization,
    currentProject,
    organizations,
    router
  ])

  if (authLoading || orgLoading || checkingOnboarding) {
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
