'use client'

import { useParams } from 'next/navigation'
import { useWorkspace } from '@/context/workspace-context'
import { SettingsView } from '@/features/settings'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/features/organizations'

export default function OrganizationSettingsPage() {
  const params = useParams() as OrganizationParams
  const { 
    currentOrganization,
    isLoading,
    error
  } = useWorkspace()

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

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="mb-6">
          <h1 className="text-2xl font-bold tracking-tight">Settings</h1>
          <p className="text-muted-foreground">
            Manage your account settings and preferences
          </p>
        </div>
        <SettingsView />
      </Main>
    </>
  )
}