'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { TraceDetailView } from '@/features/traces'
import { getTraceById } from '@/features/traces/api/traces-api'
import { useProjectOnly } from '@/features/projects'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import type { ProjectParams } from '@/features/organizations'

interface TracePageParams extends ProjectParams {
  traceId: string
}

export default function TraceDetailPage() {
  const params = useParams() as TracePageParams
  const { currentProject, hasProject } = useProjectOnly()

  // Extract project ID
  const projectId = currentProject?.id

  // Fetch trace data with React Query
  const {
    data: trace,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['trace', projectId, params.traceId],
    queryFn: async () => {
      if (!projectId) {
        throw new Error('No project selected')
      }
      return getTraceById(projectId, params.traceId)
    },
    enabled: !!projectId && hasProject,
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  // Loading State
  if (isLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className='flex flex-col items-center justify-center py-12'>
            <div className='flex flex-col items-center space-y-2'>
              <div className='h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent' />
              <p className='text-sm text-muted-foreground'>Loading trace...</p>
            </div>
          </div>
        </Main>
      </>
    )
  }

  // Error State
  if (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    const is404 = errorMessage.toLowerCase().includes('not found') || errorMessage.includes('404')

    return (
      <>
        <DashboardHeader />
        <Main>
          <div className='flex flex-col items-center justify-center py-12'>
            <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
              <h2 className='text-xl font-bold text-destructive mb-2'>
                {is404 ? 'Trace Not Found' : 'Failed to Load Trace'}
              </h2>
              <p className='text-sm text-muted-foreground mb-4'>
                {is404
                  ? "The trace you're looking for doesn't exist or has been deleted."
                  : errorMessage}
              </p>
              <div className='flex gap-2 justify-center'>
                <Link href={`/projects/${params.projectSlug}/traces`}>
                  <Button variant='outline'>
                    <ArrowLeft className='mr-2 h-4 w-4' />
                    Back to Traces
                  </Button>
                </Link>
                {!is404 && (
                  <Button onClick={() => refetch()}>
                    Try Again
                  </Button>
                )}
              </div>
            </div>
          </div>
        </Main>
      </>
    )
  }

  // No Project State
  if (!hasProject || !trace) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className='flex flex-col items-center justify-center py-12'>
            <h2 className='text-2xl font-bold mb-2'>Trace Not Found</h2>
            <p className='text-muted-foreground mb-4'>
              {!hasProject ? 'No project selected' : 'Trace data could not be loaded'}
            </p>
            <Link href={`/projects/${params.projectSlug}/traces`}>
              <Button variant='outline'>
                <ArrowLeft className='mr-2 h-4 w-4' />
                Back to Traces
              </Button>
            </Link>
          </div>
        </Main>
      </>
    )
  }

  // Success State
  return (
    <>
      <DashboardHeader />
      <Main>
        <div className='mb-4'>
          <Link href={`/projects/${params.projectSlug}/traces`}>
            <Button variant='ghost' size='sm'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to Traces
            </Button>
          </Link>
        </div>
        <TraceDetailView trace={trace} />
      </Main>
    </>
  )
}
