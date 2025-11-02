'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { useCurrentOrganization } from '@/hooks/api/use-auth-queries'
import { PageLoader } from '@/components/shared/loading'

// Session storage key for redirect tracking
const REDIRECT_KEY = 'org-redirect-initiated'

// Helper to build organization URL with composite slug
function buildOrgUrl(orgName: string, orgId: string): string {
  const slug = orgName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return `/organizations/${slug}-${orgId}`
}

export default function RootPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const { data: organization, isLoading: orgLoading, error: orgError } = useCurrentOrganization()
  const [isRedirecting, setIsRedirecting] = useState(false)

  const redirectToAppropriateLocation = useCallback(() => {
    // Prevent duplicate calls using session storage
    if (sessionStorage.getItem(REDIRECT_KEY)) return
    sessionStorage.setItem(REDIRECT_KEY, 'true')
    setIsRedirecting(true)

    if (organization) {
      // User has organization - redirect to it
      const orgUrl = buildOrgUrl(organization.name, organization.id)
      router.push(orgUrl)
    } else {
      // No organization found - redirect to creation wizard
      router.push('/organizations/create')
    }
  }, [organization, router])

  useEffect(() => {
    if (authLoading || orgLoading) return

    if (!user) {
      router.push('/auth/signin')
      return
    }

    // Handle organization fetch error - likely means no org exists
    if (orgError) {
      sessionStorage.setItem(REDIRECT_KEY, 'true')
      setIsRedirecting(true)
      router.push('/organizations/create')
      return
    }

    // Trigger redirect once organization data is loaded
    if (!isRedirecting) {
      redirectToAppropriateLocation()
    }
  }, [authLoading, orgLoading, orgError, user, router, redirectToAppropriateLocation, isRedirecting])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      sessionStorage.removeItem(REDIRECT_KEY)
    }
  }, [])

  if (authLoading || orgLoading || isRedirecting) {
    return <PageLoader message="Loading your workspace..." />
  }

  if (!user) {
    return null // Will redirect to signin
  }

  return <PageLoader message="Loading your workspace..." />
}
