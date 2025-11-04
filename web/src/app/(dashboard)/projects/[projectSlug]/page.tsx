'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useProject } from '@/context/project-context'
import { DashboardView } from '@/views/dashboard-view'
import { Skeleton } from '@/components/ui/skeleton'
import type { ProjectParams } from '@/types/organization'

export default function ProjectPage() {
  const params = useParams() as ProjectParams
  const router = useRouter()
  const {
    currentProject,
    isLoading,
    error
  } = useProject()

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
            {error || "The requested project could not be found."}
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

  return <DashboardView />
}