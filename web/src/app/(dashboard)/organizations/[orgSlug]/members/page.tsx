'use client'

import { useParams } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { MemberManagement } from '@/components/organization/member-management'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/types/organization'

export default function MembersSettingsPage() {
  const params = useParams() as OrganizationParams
  const { 
    currentOrganization,
    isLoading,
    error
  } = useOrganization()

  // TODO: Implement proper permission-based access control with backend integration
  // This page should verify user has 'members:manage' permission for the organization

  if (isLoading) {
    return (
      <>
        <Header>
          <Skeleton className="h-8 w-64" />
        </Header>
        <Main className="space-y-6">
          <Skeleton className="h-6 w-96" />
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-16" />
            ))}
          </div>
        </Main>
      </>
    )
  }

  if (error || !currentOrganization) {
    return (
      <>
        <Header>
          <h1 className="text-2xl font-bold text-foreground">Organization Not Found</h1>
        </Header>
        <Main>
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Organization Not Found</h2>
            <p className="text-muted-foreground mb-4">
              The requested organization could not be found or loaded.
            </p>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <Header>
        <div className="space-y-2">
          <Breadcrumbs />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              Member Management
            </h1>
            <p className="text-muted-foreground">
              Manage team members and their permissions for {currentOrganization.name}
            </p>
          </div>
        </div>
      </Header>

      <Main>
        <MemberManagement />
      </Main>
    </>
  )
}