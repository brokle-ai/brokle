'use client'

import { useParams, useRouter } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import {
  TraceDetailLayout,
  TracesProvider,
  getTraceById,
  getSpansForTrace,
} from '@/features/traces'
import { useProjectOnly } from '@/features/projects'
import { Button } from '@/components/ui/button'
import { ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import { Skeleton } from '@/components/ui/skeleton'
import type { ProjectParams } from '@/features/organizations'

interface TracePageParams extends ProjectParams {
  traceId: string
}

// Loading State Component
function LoadingState() {
  return (
    <div className='flex flex-col h-full'>
      <div className='border-b p-4 space-y-3'>
        <Skeleton className='h-6 w-48' />
        <Skeleton className='h-4 w-32' />
        <div className='flex gap-2'>
          <Skeleton className='h-5 w-16' />
          <Skeleton className='h-5 w-20' />
          <Skeleton className='h-5 w-24' />
        </div>
      </div>
      <div className='flex-1 flex'>
        <div className='w-1/3 border-r p-4 space-y-3'>
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-32 w-full' />
        </div>
        <div className='flex-1 p-4 space-y-4'>
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-24 w-full' />
          <Skeleton className='h-24 w-full' />
        </div>
      </div>
    </div>
  )
}

// Error State Component
function ErrorState({ message, is404, projectSlug }: { message: string; is404: boolean; projectSlug: string }) {
  return (
    <div className='flex flex-col items-center justify-center h-full py-12 text-destructive'>
      <AlertCircle className='h-12 w-12 mb-4' />
      <h3 className='text-lg font-semibold mb-2'>
        {is404 ? 'Trace Not Found' : 'Failed to load trace'}
      </h3>
      <p className='text-sm text-muted-foreground mb-4'>
        {is404
          ? "The trace you're looking for doesn't exist or has been deleted."
          : message}
      </p>
      <Link href={`/projects/${projectSlug}/traces`}>
        <Button variant='outline'>
          <ArrowLeft className='mr-2 h-4 w-4' />
          Back to Traces
        </Button>
      </Link>
    </div>
  )
}

// Inner content that uses TracesProvider context
function TraceDetailPageContent() {
  const params = useParams() as TracePageParams
  const { currentProject, hasProject } = useProjectOnly()
  const projectId = currentProject?.id

  // Fetch trace data with React Query
  const {
    data: trace,
    isLoading: traceLoading,
    error: traceError,
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

  // Fetch spans for this trace
  const {
    data: spans = [],
    isLoading: spansLoading,
  } = useQuery({
    queryKey: ['traceSpans', projectId, params.traceId],
    queryFn: () => getSpansForTrace(projectId!, params.traceId),
    enabled: !!projectId && !!params.traceId && !!trace,
    staleTime: 30_000,
  })

  // Loading State
  if (traceLoading) {
    return <LoadingState />
  }

  // Error State
  if (traceError) {
    const errorMessage = traceError instanceof Error ? traceError.message : String(traceError)
    const is404 = errorMessage.toLowerCase().includes('not found') || errorMessage.includes('404')
    return <ErrorState message={errorMessage} is404={is404} projectSlug={params.projectSlug} />
  }

  // No Project/Trace State
  if (!hasProject || !trace) {
    return (
      <ErrorState
        message={!hasProject ? 'No project selected' : 'Trace data could not be loaded'}
        is404={!hasProject}
        projectSlug={params.projectSlug}
      />
    )
  }

  // Success State - Use unified layout with page context
  return (
    <TraceDetailLayout
      trace={trace}
      spans={spans}
      spansLoading={spansLoading}
      projectId={projectId!}
      context='page'
    />
  )
}

export default function TraceDetailPage() {
  const params = useParams() as TracePageParams

  return (
    <TracesProvider projectSlug={params.projectSlug}>
      <DashboardHeader />
      <Main fixed className='p-0 sm:p-0'>
        <TraceDetailPageContent />
      </Main>
    </TracesProvider>
  )
}
