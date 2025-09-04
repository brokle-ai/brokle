'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useOrganization, useProject } from '@/context/organization-context'
import { DashboardView } from '@/views/dashboard-view'
import { Skeleton } from '@/components/ui/skeleton'
import type { ProjectParams } from '@/types/organization'

export default function ProjectPage() {
  const params = useParams() as ProjectParams
  const router = useRouter()
  const { 
    currentOrganization,
    projects,
    isLoading: orgLoading,
    error: orgError,
    hasAccess
  } = useOrganization()
  
  const {
    currentProject,
    hasProject
  } = useProject()

  useEffect(() => {
    // Don't do any access checks or redirects while loading
    if (orgLoading) return

    // Only check access after loading is complete AND we have organization data
    if (!currentOrganization) return

    // For project access, we need to wait for projects to be loaded too
    // Otherwise hasAccess will return false during loading and cause redirect flash
    if (projects.length === 0) return

    // Check if user has access to this organization and project
    if (!hasAccess(params.orgSlug, params.projectSlug)) {
      router.push('/')
      return
    }

    // If we're in the wrong context, let the organization context handle switching
    if (currentOrganization.slug !== params.orgSlug) {
      return // Context will handle the switch
    }

    if (currentProject && currentProject.slug !== params.projectSlug) {
      return // Context will handle the switch
    }
  }, [
    params.orgSlug,
    params.projectSlug,
    currentOrganization,
    currentProject,
    projects,
    orgLoading,
    hasAccess,
    router
  ])

  // Show loading skeleton during authentication and context loading
  if (orgLoading) {
    return (
      <div className="space-y-6 p-6">
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-5 w-96" />
        </div>
        
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
        
        <div className="grid gap-6 md:grid-cols-2">
          <Skeleton className="h-80" />
          <Skeleton className="h-80" />
        </div>
      </div>
    )
  }

  // Only show error states after loading is complete
  if (orgError || !currentOrganization) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Organization Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            {orgError || "The requested organization could not be found."}
          </p>
          <button 
            onClick={() => router.push('/')}
            className="text-primary hover:underline"
          >
            Go to organization selector
          </button>
        </div>
      </div>
    )
  }

  // Check for project after organization is loaded AND context loading is complete
  // Don't show "Project Not Found" while the context is still loading/updating
  if (!hasProject && !orgLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Project Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            The requested project could not be found in {currentOrganization.name}.
          </p>
          <button 
            onClick={() => router.push(`/${currentOrganization.slug}`)}
            className="text-primary hover:underline"
          >
            Go back to {currentOrganization.name}
          </button>
        </div>
      </div>
    )
  }

  return <DashboardView />
}