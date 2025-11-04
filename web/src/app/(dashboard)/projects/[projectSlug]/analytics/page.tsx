'use client'

import { useProject } from '@/context/project-context'
import { AnalyticsView } from '@/views/analytics-view'
import { Skeleton } from '@/components/ui/skeleton'

export default function ProjectAnalyticsPage() {
  const { 
    currentProject,
    isLoading,
    error
  } = useProject()

  if (isLoading) {
    return (
      <div className="space-y-6 p-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
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

  return <AnalyticsView />
}