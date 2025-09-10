'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { SettingsView } from '@/views/settings-view'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/types/organization'

export default function OrganizationSettingsPage() {
  const params = useParams() as OrganizationParams
  const router = useRouter()
  const { 
    currentOrganization,
    isLoading,
    error,
    hasAccess,
    getUserRole
  } = useOrganization()

  useEffect(() => {
    if (isLoading) return

    if (!hasAccess(params.orgSlug)) {
      router.push('/')
      return
    }

    // Check if user has admin permissions
    const userRole = getUserRole(params.orgSlug)
    if (userRole !== 'owner' && userRole !== 'admin') {
      router.push(`/${params.orgSlug}`)
      return
    }
  }, [params.orgSlug, isLoading, hasAccess, getUserRole, router])

  if (isLoading) {
    return (
      <div className="space-y-6 p-6">
        <Skeleton className="h-8 w-64" />
        <div className="space-y-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} className="h-16" />
          ))}
        </div>
      </div>
    )
  }

  if (error || !currentOrganization) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Access Denied
          </h1>
          <p className="text-muted-foreground mb-4">
            You don't have permission to access organization settings.
          </p>
          <button 
            onClick={() => router.push(currentOrganization ? `/${currentOrganization.slug}` : '/')}
            className="text-primary hover:underline"
          >
            Go back
          </button>
        </div>
      </div>
    )
  }

  return <SettingsView />
}