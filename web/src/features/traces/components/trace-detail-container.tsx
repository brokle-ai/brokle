'use client'

import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircle } from 'lucide-react'

import { useProjectOnly } from '@/features/projects'
import { useTraceDetailState } from '../hooks/use-trace-detail-state'
import { getTraceById, getSpansForTrace } from '../api/traces-api'
import { TraceDetailLayout } from './trace-detail-layout'

// ============================================================================
// Loading State
// ============================================================================

function TraceDetailSkeleton() {
  return (
    <div className='flex flex-col h-full'>
      {/* Header skeleton */}
      <div className='border-b p-4 space-y-3'>
        <Skeleton className='h-6 w-48' />
        <Skeleton className='h-4 w-32' />
        <div className='flex gap-2'>
          <Skeleton className='h-5 w-16' />
          <Skeleton className='h-5 w-20' />
          <Skeleton className='h-5 w-24' />
        </div>
      </div>

      {/* Content skeleton */}
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

// ============================================================================
// Error State
// ============================================================================

function TraceDetailError({ message }: { message: string }) {
  return (
    <div className='flex flex-col items-center justify-center h-full py-12 text-destructive'>
      <AlertCircle className='h-12 w-12 mb-4' />
      <h3 className='text-lg font-semibold mb-2'>Failed to load trace</h3>
      <p className='text-sm text-muted-foreground'>{message}</p>
    </div>
  )
}

// ============================================================================
// Main Container Component
// ============================================================================

/**
 * TraceDetailContainer - Sheet-based peek view for trace details
 *
 * Renders trace detail in a sheet sidebar (70vw) when a trace is selected.
 * For full-page view, use the /traces/[traceId] route instead.
 *
 * Uses useTraceDetailState for URL state management (peek, span, view params)
 */
export function TraceDetailContainer() {
  const { traceId, closeTrace } = useTraceDetailState()
  const { currentProject, hasProject } = useProjectOnly()
  const projectId = currentProject?.id

  // Fetch trace data
  const {
    data: trace,
    isLoading: traceLoading,
    error: traceError,
  } = useQuery({
    queryKey: ['trace', projectId, traceId],
    queryFn: async () => {
      if (!projectId || !traceId) {
        throw new Error('Missing project or trace ID')
      }
      return getTraceById(projectId, traceId)
    },
    enabled: !!projectId && !!traceId && hasProject,
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })

  // Fetch spans for this trace
  const {
    data: spans = [],
    isLoading: spansLoading,
  } = useQuery({
    queryKey: ['traceSpans', projectId, traceId],
    queryFn: () => getSpansForTrace(projectId!, traceId!),
    enabled: !!projectId && !!traceId && !!trace,
    staleTime: 30_000,
  })

  // Keyboard shortcuts
  React.useEffect(() => {
    if (!traceId) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Escape to close (only when not in an input/textarea)
      if (e.key === 'Escape') {
        const target = e.target as HTMLElement
        if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
          e.preventDefault()
          closeTrace()
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [traceId, closeTrace])

  // Don't render anything if no trace is selected
  if (!traceId) return null

  // Determine content to render
  const isLoading = traceLoading
  const error = traceError instanceof Error ? traceError : null

  const content = isLoading ? (
    <TraceDetailSkeleton />
  ) : error || !trace ? (
    <TraceDetailError message={error?.message || 'Trace not found'} />
  ) : (
    <TraceDetailLayout
      trace={trace}
      spans={spans}
      spansLoading={spansLoading}
      projectId={projectId!}
      context='peek'
    />
  )

  // Render in Sheet sidebar (peek mode)
  return (
    <Sheet open={!!traceId} onOpenChange={(open) => !open && closeTrace()} modal={false}>
      <SheetContent
        side='right'
        className='flex max-h-full min-h-0 min-w-[70vw] flex-col gap-0 overflow-hidden rounded-l-xl p-0'
        onPointerDownOutside={(e) => {
          // Prevent sheet closure when clicking outside
          e.preventDefault()
        }}
        tabIndex={-1}
        hideCloseButton
      >
        {/* Visually hidden title for accessibility */}
        <SheetTitle className='sr-only'>Trace Details</SheetTitle>
        {content}
      </SheetContent>
    </Sheet>
  )
}
