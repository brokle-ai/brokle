'use client'

import { useProjectOnly } from '@/features/projects'
import { ModelsView } from '@/features/analytics'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Skeleton } from '@/components/ui/skeleton'

export default function ProjectModelsPage() {
  const { 
    currentProject,
    isLoading,
    error
  } = useProjectOnly()

  // TODO: Implement proper permission-based access control with backend integration
  // This page should verify user has 'models:read' permission for the project

  if (isLoading) {
    return (
      <div className="space-y-6 p-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} className="h-16" />
          ))}
        </div>
      </div>
    )
  }

  if (error || !currentProject) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Project Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            The requested project could not be found.
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
          <h1 className="text-2xl font-bold tracking-tight">Models</h1>
          <p className="text-muted-foreground">
            Compare AI model performance and usage for {currentProject.name}
          </p>
        </div>
        <ModelsView />
      </Main>
    </>
  )
}