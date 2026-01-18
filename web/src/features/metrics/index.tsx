'use client'

import { BarChart3 } from 'lucide-react'
import { MetricsView } from './components/metrics-view'
import { useTraceMetrics } from './hooks/use-trace-metrics'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface MetricsProps {
  projectSlug?: string
}

export function Metrics({ projectSlug }: MetricsProps) {
  const { metrics, isLoading, error, hasProject, refetch } = useTraceMetrics()

  return (
    <div className='-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1'>
      {/* Error state */}
      {error && (
        <div className='flex flex-col items-center justify-center py-12 space-y-4'>
          <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
            <h3 className='font-semibold text-destructive mb-2'>Failed to load metrics</h3>
            <p className='text-sm text-muted-foreground mb-4'>{error}</p>
            <button
              onClick={() => refetch()}
              className='inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors'
            >
              Try Again
            </button>
          </div>
        </div>
      )}

      {/* No project selected */}
      {!hasProject && !error && (
        <div className='flex flex-col items-center justify-center py-12 text-center'>
          <p className='text-muted-foreground'>No project selected</p>
        </div>
      )}

      {/* Initial loading state */}
      {isLoading && (
        <div className='flex flex-1 items-center justify-center py-16'>
          <LoadingSpinner message='Loading metrics...' />
        </div>
      )}

      {/* Empty state */}
      {!error && hasProject && !isLoading && metrics && metrics.totalTraces === 0 && (
        <DataTableEmptyState
          icon={<BarChart3 className='h-full w-full' />}
          title='No metrics data'
          description='Start sending traces to see usage analytics and cost metrics.'
        />
      )}

      {/* Metrics view */}
      {!error && hasProject && !isLoading && metrics && metrics.totalTraces > 0 && (
        <MetricsView />
      )}
    </div>
  )
}

// Hooks
export { useTraceMetrics, useMetricsState, TIME_RANGES } from './hooks/use-trace-metrics'

// Types
export type { TraceMetrics, TimeRange, UseTraceMetricsReturn, UseMetricsStateReturn } from './hooks/use-trace-metrics'
