'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useOrganization } from '@/context/organization-context'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationParams } from '@/types/organization'

interface ProjectSettingsLayoutParams extends OrganizationParams {
  projectSlug: string
}

export default function ProjectSettingsLayout({
  children
}: {
  children: React.ReactNode
}) {
  const params = useParams() as ProjectSettingsLayoutParams
  const router = useRouter()
  const { 
    currentOrganization,
    currentProject,
    isLoading,
    hasAccess,
    getUserRole,
    switchProject
  } = useOrganization()

  useEffect(() => {
    if (isLoading) return

    if (!hasAccess(params.orgSlug)) {
      router.push('/')
      return
    }

    // Check if user has access to project settings
    const userRole = getUserRole(params.orgSlug)
    if (userRole !== 'owner' && userRole !== 'admin' && userRole !== 'developer') {
      router.push(`/${params.orgSlug}/projects/${params.projectSlug}`)
      return
    }

    // Ensure correct project is loaded
    if (!currentProject || currentProject.slug !== params.projectSlug) {
      switchProject(params.projectSlug).catch(() => {
        router.push(`/${params.orgSlug}`)
      })
    }
  }, [params.orgSlug, params.projectSlug, isLoading, hasAccess, getUserRole, currentProject, switchProject, router])

  if (isLoading) {
    return (
      <>
        <Header>
          <Skeleton className="h-8 w-64" />
        </Header>
        <Main className="space-y-6">
          <Skeleton className="h-6 w-96" />
          <Skeleton className="h-10 w-full" />
          <div className="space-y-4">
            <Skeleton className="h-64" />
          </div>
        </Main>
      </>
    )
  }

  if (!currentOrganization || !currentProject) {
    return (
      <>
        <Header>
          <h1 className="text-2xl font-bold text-foreground">Project Not Found</h1>
        </Header>
        <Main>
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Project Not Found</h2>
            <p className="text-muted-foreground mb-4">
              The project you're looking for doesn't exist or you don't have access to it.
            </p>
            <button 
              onClick={() => router.push(`/${params.orgSlug}`)}
              className="text-primary hover:underline"
            >
              Go back to organization
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
              {currentProject.name} Settings
            </h1>
            <p className="text-muted-foreground">
              Configure settings and preferences for this project
            </p>
          </div>
        </div>
      </Header>

      <Main className="space-y-6">
        <Tabs defaultValue="general" className="space-y-6">
          <TabsList>
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="api-keys">API Keys</TabsTrigger>
            <TabsTrigger value="models">Models</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
            <TabsTrigger value="integrations">Integrations</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
            <TabsTrigger value="danger">Danger Zone</TabsTrigger>
          </TabsList>
          
          {children}
        </Tabs>
      </Main>
    </>
  )
}