'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { MemberManagement } from '@/components/organization/member-management'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/types/organization'

export default function MembersSettingsPage() {
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
          <h1 className="text-2xl font-bold text-foreground">Access Denied</h1>
        </Header>
        <Main>
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Access Denied</h2>
            <p className="text-muted-foreground mb-4">
              You don't have permission to manage members for this organization.
            </p>
            <button 
              onClick={() => router.push(currentOrganization ? `/${currentOrganization.slug}` : '/')}
              className="text-primary hover:underline"
            >
              Go back
            </button>
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