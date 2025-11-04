'use client'

import { useProjectOnly } from '@/hooks/use-project-only'
import { CostsView } from '@/views/costs-view'
import { Skeleton } from '@/components/ui/skeleton'

export default function ProjectCostsPage() {
  const { 
    currentProject,
    isLoading,
    error
  } = useProjectOnly()

  // TODO: Implement proper permission-based access control with backend integration
  // This page should verify user has 'costs:read' permission for the project

  if (isLoading) {
    return (
      <div className="space-y-6 p-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
        <div className="grid gap-6 md:grid-cols-2">
          <Skeleton className="h-80" />
          <Skeleton className="h-80" />
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

  return <CostsView />
}