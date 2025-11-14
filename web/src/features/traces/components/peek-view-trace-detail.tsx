'use client'

import { usePeekData } from '../hooks/use-peek-data'
import { TraceDetailView } from './trace-detail-view'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircle } from 'lucide-react'

/**
 * Wrapper component for peek view
 * Fetches trace data and renders detail view or loading/error states
 */
export function PeekViewTraceDetail() {
  const { trace, isLoading, error, peekId } = usePeekData()

  if (!peekId) {
    return (
      <div className='flex items-center justify-center py-12 text-muted-foreground'>
        Select a trace to view details
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className='space-y-4 p-6'>
        <Skeleton className='h-8 w-3/4' />
        <Skeleton className='h-4 w-1/2' />
        <div className='grid grid-cols-2 md:grid-cols-4 gap-4 mt-6'>
          <Skeleton className='h-24' />
          <Skeleton className='h-24' />
          <Skeleton className='h-24' />
          <Skeleton className='h-24' />
        </div>
        <Skeleton className='h-64 mt-6' />
      </div>
    )
  }

  if (error || !trace) {
    return (
      <div className='flex flex-col items-center justify-center py-12 text-destructive'>
        <AlertCircle className='h-12 w-12 mb-4' />
        <h3 className='text-lg font-semibold mb-2'>Failed to load trace</h3>
        <p className='text-sm text-muted-foreground'>{error?.message || 'Trace not found'}</p>
      </div>
    )
  }

  return <TraceDetailView trace={trace} />
}
