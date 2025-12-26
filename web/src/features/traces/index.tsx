'use client'

import { useSearchParams } from 'next/navigation'
import { Activity } from 'lucide-react'
import { TracesProvider } from './context/traces-context'
import { TracesTable } from './components/traces-table'
import { useProjectTraces } from './hooks/use-project-traces'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { PageHeader } from '@/components/layout/page-header'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface TracesProps {
  projectSlug?: string
}

function TracesContent() {
  const searchParams = useSearchParams()
  const { data, totalCount, isLoading, isFetching, error, hasProject, refetch } = useProjectTraces()
  const { filter, status } = useTableSearchParams(searchParams)

  // Check if there are active filters
  const hasActiveFilters = !!(filter || status.length > 0)

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && data.length === 0

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters

  return (
    <>
      <PageHeader title="Traces" />
      <div className='-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1'>
        {/* Initial loading (first load, no cache) */}
        {isInitialLoad && (
          <div className='flex flex-1 items-center justify-center py-16'>
            <LoadingSpinner message="Loading traces..." />
          </div>
        )}

        {/* Error state */}
        {error && !isInitialLoad && (
          <div className='flex flex-col items-center justify-center py-12 space-y-4'>
            <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
              <h3 className='font-semibold text-destructive mb-2'>Failed to load traces</h3>
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
        {!hasProject && !isInitialLoad && !error && (
          <div className='flex flex-col items-center justify-center py-12 text-center'>
            <p className='text-muted-foreground'>No project selected</p>
          </div>
        )}

        {/* Empty project (never had data) */}
        {!error && hasProject && !isInitialLoad && isEmptyProject && (
          <DataTableEmptyState
            icon={<Activity className="h-full w-full" />}
            title="No traces yet"
            description="Start sending traces from your application to see them here."
          />
        )}

        {/* Table (has data OR has active filters) */}
        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <TracesTable data={data} totalCount={totalCount} isFetching={isFetching} />
        )}
      </div>
    </>
  )
}

export function Traces({ projectSlug }: TracesProps) {
  return (
    <TracesProvider projectSlug={projectSlug}>
      <TracesContent />
    </TracesProvider>
  )
}

export { TraceDetailView } from './components/trace-detail-view'
export { TraceDetailLayout } from './components/trace-detail-layout'

// Context
export { TracesProvider, useTraces } from './context/traces-context'

// Hooks
export { useProjectTraces, useTraceFilterOptions } from './hooks/use-project-traces'
export { useTraceDetailState } from './hooks/use-trace-detail-state'
export { usePeekSheetState } from './hooks/use-peek-sheet-state'
export { usePeekData } from './hooks/use-peek-data'

// Types
export type { TraceFilterOptions, FilterRange, GetTracesParams } from './api/traces-api'
export type { UseProjectTracesReturn, UseTraceFilterOptionsReturn } from './hooks/use-project-traces'
export type { ViewMode } from './hooks/use-trace-detail-state'

// API
export { getProjectTraces, getTraceFilterOptions, getTraceById, getSpansForTrace } from './api/traces-api'
