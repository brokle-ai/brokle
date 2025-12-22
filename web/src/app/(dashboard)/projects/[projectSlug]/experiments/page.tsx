'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { useProjectOnly } from '@/features/projects'
import {
  ExperimentList,
  CreateExperimentDialog,
} from '@/features/experiments'
import { Skeleton } from '@/components/ui/skeleton'

export default function ExperimentsPage() {
  const params = useParams<{ projectSlug: string }>()
  const { currentProject, hasProject, isLoading } = useProjectOnly()

  if (isLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <Skeleton className="h-8 w-32" />
              <Skeleton className="h-10 w-40" />
            </div>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-32" />
              ))}
            </div>
          </div>
        </Main>
      </>
    )
  }

  if (!hasProject || !currentProject) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex items-center justify-center py-12">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">Experiments</h1>
              <p className="text-muted-foreground">
                Run and compare model outputs with batch evaluations
              </p>
            </div>
            <CreateExperimentDialog projectId={currentProject.id} />
          </div>
          <ExperimentList
            projectId={currentProject.id}
            projectSlug={params.projectSlug}
          />
        </div>
      </Main>
    </>
  )
}
