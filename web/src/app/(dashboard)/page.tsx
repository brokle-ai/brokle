'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/features/authentication'
import { useCurrentOrganization } from '@/features/authentication'
import { CreateOrganizationDialog } from '@/features/organizations'
import { PageLoader } from '@/components/shared/loading'
import { buildOrgUrl } from '@/lib/utils/slug-utils'
import { Button } from '@/components/ui/button'
import { Plus, Building2 } from 'lucide-react'

export default function RootPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const {
    data: organization,
    isLoading: orgLoading,
    error: orgError,
  } = useCurrentOrganization()
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)

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
    }
    // If no organization, stay on this page and show empty state
  }, [organization, router])

  useEffect(() => {
    if (authLoading || orgLoading) return

    if (!user) {
      router.push('/auth/signin')
      return
    }

    // Handle organization fetch error - user likely has no org, show empty state
    if (orgError) {
      redirectInitiatedRef.current = true
      setIsRedirecting(false) // Don't show redirecting loader
      return
    }

    // If we previously had an error but now have org data, reset the flag
    if (organization && redirectInitiatedRef.current && !isRedirecting) {
      redirectInitiatedRef.current = false
    }

    // Trigger redirect once organization data is loaded
    if (!isRedirecting && organization) {
      redirectToAppropriateLocation()
    }
  }, [
    authLoading,
    orgLoading,
    orgError,
    user,
    organization,
    router,
    redirectToAppropriateLocation,
    isRedirecting,
  ])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      redirectInitiatedRef.current = false
    }
  }, [])

  if (authLoading || orgLoading) {
    return <PageLoader message="Loading your workspace..." />
  }

  if (!user) {
    return null // Will redirect to signin
  }

  // User has organization - redirecting
  if (organization && isRedirecting) {
    return <PageLoader message="Loading your workspace..." />
  }

  // No organization - show empty state with creation dialog
  if (!organization || orgError) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="flex flex-col items-center justify-center text-center max-w-md">
          <div className="mb-6 rounded-full bg-muted p-6">
            <Building2 className="h-16 w-16 text-muted-foreground" />
          </div>
          <h1 className="text-2xl font-bold mb-3">Welcome to Brokle</h1>
          <p className="text-muted-foreground mb-8">
            Create your first organization to start managing AI projects and
            team members.
          </p>

          <Button size="lg" onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 h-5 w-5" />
            Create Your First Organization
          </Button>

          <CreateOrganizationDialog
            open={dialogOpen}
            onOpenChange={setDialogOpen}
          />
        </div>
      </div>
    )
  }

  return <PageLoader message="Loading your workspace..." />
}
