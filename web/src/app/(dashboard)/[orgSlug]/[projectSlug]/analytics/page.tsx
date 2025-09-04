'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useOrganization } from '@/context/organization-context'
import { AnalyticsView } from '@/views/analytics-view'
import { Skeleton } from '@/components/ui/skeleton'
import type { ProjectParams } from '@/types/organization'

export default function ProjectAnalyticsPage() {
  const params = useParams() as ProjectParams
  const router = useRouter()
  const { 
    currentOrganization,
    currentProject,
    isLoading,
    error,
    hasAccess
  } = useOrganization()

  useEffect(() => {
    if (isLoading) return

    if (!hasAccess(params.orgSlug, params.projectSlug)) {
      router.push('/')
      return
    }
  }, [params.orgSlug, params.projectSlug, isLoading, hasAccess, router])

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

  if (error || !currentOrganization || !currentProject) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            Project Not Found
          </h1>
          <p className="text-muted-foreground mb-4">
            The requested project could not be found.
          </p>
          <button 
            onClick={() => router.push('/')}
            className="text-primary hover:underline"
          >
            Go back to organization selector
          </button>
        </div>
      </div>
    )
  }

  return <AnalyticsView />
}