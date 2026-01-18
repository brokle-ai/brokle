'use client'

import { MessageSquare } from 'lucide-react'
import { SessionsTable } from './components/sessions-table'
import { useProjectSessions } from './hooks/use-project-sessions'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface SessionsProps {
  projectSlug?: string
}

export function Sessions({ projectSlug }: SessionsProps) {
  const { data, totalCount, isLoading, isFetching, error, hasProject, refetch, tableState } =
    useProjectSessions()

  // True initial load: ONLY when first loading with absolutely no data
  const isInitialLoad = isLoading && data.length === 0 && !tableState.hasActiveFilters

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmpty = !isLoading && totalCount === 0 && !tableState.hasActiveFilters

  return (
    <div className='-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1'>
      {/* Error state */}
      {error && !isInitialLoad && (
        <div className='flex flex-col items-center justify-center py-12 space-y-4'>
          <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
            <h3 className='font-semibold text-destructive mb-2'>Failed to load sessions</h3>
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

      {/* Initial loading state */}
      {isInitialLoad && (
        <div className='flex flex-1 items-center justify-center py-16'>
          <LoadingSpinner message='Loading sessions...' />
        </div>
      )}

      {/* Empty state */}
      {!error && hasProject && !isInitialLoad && isEmpty && (
        <DataTableEmptyState
          icon={<MessageSquare className='h-full w-full' />}
          title='No sessions yet'
          description='Sessions group multi-turn conversations. Add session_id to your traces to see them here.'
        />
      )}

      {/* Data table */}
      {!error && hasProject && !isInitialLoad && !isEmpty && (
        <SessionsTable data={data} totalCount={totalCount} isFetching={isFetching} />
      )}
    </div>
  )
}

// Hooks
export { useProjectSessions, useSessionsTableState } from './hooks/use-project-sessions'

// Types
export type { Session, UseProjectSessionsReturn, UseSessionsTableStateReturn } from './hooks/use-project-sessions'
