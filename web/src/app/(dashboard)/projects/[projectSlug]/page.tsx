'use client'

import { useRouter } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { DashboardView } from '@/features/projects'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { Skeleton } from '@/components/ui/skeleton'

export default function ProjectPage() {
  const router = useRouter()
  const {
    currentProject,
    isLoading,
    error
  } = useProjectOnly()

  // No need for redirect logic anymore - direct ID lookup handles this

  // Show loading skeleton during authentication and context loading
  if (isLoading) {
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
  if (error || !currentProject) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Project Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            {error?.userMessage || "The requested project could not be found."}
          </p>
          <button 
            onClick={() => router.push('/')}
            className="text-primary hover:underline"
          >
            Go back to project selector
          </button>
        </div>
      </div>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <PageHeader title={currentProject.name} />
        <DashboardView />
      </Main>
    </>
  )
}