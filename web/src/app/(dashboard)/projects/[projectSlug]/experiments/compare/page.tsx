'use client'

import { Suspense } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { useProjectOnly } from '@/features/projects'
import { ExperimentCompareView } from '@/features/experiments'
import { Skeleton } from '@/components/ui/skeleton'
import { Loader2 } from 'lucide-react'

function ExperimentCompareContent() {
  const { currentProject, hasProject, isLoading } = useProjectOnly()

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <Skeleton className="h-8 w-8 rounded-full" />
          <div>
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-4 w-64 mt-2" />
          </div>
        </div>
        <Skeleton className="h-10 w-[400px]" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      </div>
    )
  }

  if (!hasProject || !currentProject) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  return <ExperimentCompareView projectId={currentProject.id} />
}

export default function ExperimentComparePage() {
  return (
    <>
      <DashboardHeader />
      <Main>
        <Suspense
          fallback={
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          }
        >
          <ExperimentCompareContent />
        </Suspense>
      </Main>
    </>
  )
}
