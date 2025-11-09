'use client'

import { TracesProvider } from './context/traces-context'
import { TracesTable } from './components/traces-table'
import { useProjectTraces } from './hooks/use-project-traces'

interface TracesProps {
  projectSlug?: string
}

function TracesContent() {
  const { data, totalCount, isLoading } = useProjectTraces()

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
        {isLoading ? (
          <div className='flex items-center justify-center py-8 text-muted-foreground'>
            Loading traces...
          </div>
        ) : (
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
