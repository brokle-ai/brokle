'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { useCurrentOrganization } from '@/hooks/api/use-auth-queries'
import { PageLoader } from '@/components/shared/loading'
import { buildOrgUrl } from '@/lib/utils/slug-utils'

export default function RootPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const { data: organization, isLoading: orgLoading, error: orgError } = useCurrentOrganization()
  const [isRedirecting, setIsRedirecting] = useState(false)

  // Use ref instead of session storage to prevent race conditions
  const redirectInitiatedRef = useRef(false)

  const redirectToAppropriateLocation = useCallback(() => {
    // Prevent duplicate calls using ref
    if (redirectInitiatedRef.current) return
    redirectInitiatedRef.current = true
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
      redirectInitiatedRef.current = true
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
      redirectInitiatedRef.current = false
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
