'use client'

import { TracesProvider } from './context/traces-context'
import { TracesTable } from './components/traces-table'
import { useProjectTraces } from './hooks/use-project-traces'

interface TracesProps {
  projectSlug?: string
}

function TracesContent() {
  const { data, totalCount, isLoading, error, hasProject, refetch } = useProjectTraces()

  return (
    <>
      <div className='mb-6 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>Traces</h2>
          <p className='text-muted-foreground'>
            View and analyze distributed traces for this project
          </p>
        </div>
      </div>
      <div className='-mx-4 flex-1 overflow-auto px-4 py-1'>
        {/* Error State */}
        {error && !isLoading && (
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

        {/* No Project State */}
        {!hasProject && !isLoading && !error && (
          <div className='flex flex-col items-center justify-center py-12 text-center'>
            <p className='text-muted-foreground'>No project selected</p>
          </div>
        )}

        {/* Loading State */}
        {isLoading && (
          <div className='flex items-center justify-center py-8'>
            <div className='flex flex-col items-center space-y-2'>
              <div className='h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent' />
              <p className='text-sm text-muted-foreground'>Loading traces...</p>
            </div>
          </div>
        )}

        {/* Empty State */}
        {!isLoading && !error && hasProject && data.length === 0 && (
          <div className='flex flex-col items-center justify-center py-12 text-center'>
            <div className='rounded-lg border border-dashed p-8 max-w-md'>
              <h3 className='font-semibold mb-2'>No traces found</h3>
              <p className='text-sm text-muted-foreground'>
                Start sending telemetry data to this project to see traces here.
              </p>
            </div>
          </div>
        )}

        {/* Data Table */}
        {!isLoading && !error && hasProject && data.length > 0 && (
          <TracesTable data={data} totalCount={totalCount} />
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
