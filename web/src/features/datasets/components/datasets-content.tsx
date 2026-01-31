'use client'

import { Database } from 'lucide-react'
import { DatasetsProvider, useDatasets } from '../context/datasets-context'
import { DatasetsDialogs } from './datasets-dialogs'
import { CreateDatasetDialog } from './create-dataset-dialog'
import { DatasetsTable } from './dataset-list/datasets-table'
import { useDatasetsQuery } from '../hooks/use-datasets'
import { useDatasetsTableState } from '../hooks/use-datasets-table-state'
import { useProjectOnly } from '@/features/projects'
import { PageHeader } from '@/components/layout/page-header'
import { DataTableEmptyState } from '@/components/data-table'
import type { DatasetWithItemCount } from '../types'

interface DatasetsProps {
  projectSlug: string
}

function DatasetsContent() {
  const { currentProject } = useProjectOnly()
  const { setOpen, setCurrentRow } = useDatasets()
  const tableState = useDatasetsTableState()

  // Fetch paginated data using API params from table state
  const {
    data: response,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useDatasetsQuery(currentProject?.id, tableState.toApiParams())

  // Extract data from response
  const data = response?.datasets ?? []
  const totalCount = response?.totalCount ?? 0

  // Check if there are active filters
  const hasActiveFilters = tableState.hasActiveFilters

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters

  // Handle delete action from table
  const handleDelete = (dataset: DatasetWithItemCount) => {
    setCurrentRow(dataset)
    setOpen('delete')
  }

  return (
    <>
      <PageHeader title="Datasets">
        {currentProject && <CreateDatasetDialog projectId={currentProject.id} />}
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {error && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">Failed to load datasets</h3>
              <p className="text-sm text-muted-foreground mb-4">
                {error instanceof Error ? error.message : 'An error occurred'}
              </p>
              <button
                onClick={() => refetch()}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        )}

        {!currentProject && !isLoading && !error && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {!error && currentProject && isEmptyProject && (
          <DataTableEmptyState
            icon={<Database className="h-full w-full" />}
            title="No datasets yet"
            description="Create a dataset to start batch evaluations"
          />
        )}

        {!error && currentProject && !isEmptyProject && (
          <DatasetsTable
            data={data}
            totalCount={totalCount}
            isLoading={isLoading}
            isFetching={isFetching}
            projectSlug={currentProject.slug}
            onDelete={handleDelete}
          />
        )}
      </div>
      <DatasetsDialogs />
    </>
  )
}

export function Datasets({ projectSlug }: DatasetsProps) {
  return (
    <DatasetsProvider projectSlug={projectSlug}>
      <DatasetsContent />
    </DatasetsProvider>
  )
}
