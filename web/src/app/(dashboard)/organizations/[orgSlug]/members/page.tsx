'use client'

import { useParams } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { MemberManagement } from '@/components/organization/member-management'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ContextNavbar } from '@/components/layout/context-navbar'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
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
          <ContextNavbar />
          <div className='ml-auto flex items-center space-x-4'>
            <Search />
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
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
          <ContextNavbar />
          <div className='ml-auto flex items-center space-x-4'>
            <Search />
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
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
        <ContextNavbar />
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main>
        <div className="mb-6">
          <h1 className="text-2xl font-bold tracking-tight">Member Management</h1>
          <p className="text-muted-foreground">
            Manage team members and their permissions for {currentOrganization.name}
          </p>
        </div>
        
        <MemberManagement />
      </Main>
    </>
  )
}