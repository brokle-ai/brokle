'use client'

import { useParams } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { SettingsView } from '@/views/settings-view'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/types/organization'

export default function OrganizationSettingsPage() {
  const params = useParams() as OrganizationParams
  const { 
    currentOrganization,
    isLoading,
    error
  } = useOrganization()

  // TODO: Implement proper permission-based access control with backend integration
  // This page should verify user has 'settings:write' permission for the organization

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
            Organization Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            The requested organization could not be found or loaded.
          </p>
        </div>
      </div>
    )
  }

  return <SettingsView />
}