'use client'

import { useSearchParams } from 'next/navigation'
import { Database } from 'lucide-react'
import { DatasetsProvider } from '../context/datasets-context'
import { DatasetList } from './dataset-list'
import { DatasetsDialogs } from './datasets-dialogs'
import { CreateDatasetDialog } from './create-dataset-dialog'
import { useProjectDatasets } from '../hooks/use-project-datasets'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useCardListNavigation } from '@/hooks/use-card-list-navigation'
import { PageHeader } from '@/components/layout/page-header'
import { CardListToolbar, CardListPagination } from '@/components/card-list'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface DatasetsProps {
  projectSlug: string
}

function DatasetsContent() {
  const searchParams = useSearchParams()
  const { data, totalCount, page, pageSize, isLoading, isFetching, error, hasProject, refetch, currentProject } =
    useProjectDatasets()
  const { filter } = useTableSearchParams(searchParams)
  const { handleSearch, handleReset, handlePageChange, handlePageSizeChange } = useCardListNavigation({ searchParams })

  // Check if there are active filters
  const hasActiveFilters = !!filter

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && data.length === 0

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters

  return (
    <>
      <PageHeader title="Datasets">
        {currentProject && <CreateDatasetDialog projectId={currentProject.id} />}
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {isInitialLoad && (
          <div className="flex flex-1 items-center justify-center py-16">
            <LoadingSpinner message="Loading datasets..." />
          </div>
        )}

        {error && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">Failed to load datasets</h3>
              <p className="text-sm text-muted-foreground mb-4">{error}</p>
              <button
                onClick={() => refetch()}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        )}

        {!hasProject && !isInitialLoad && !error && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {!error && hasProject && !isInitialLoad && isEmptyProject && (
          <DataTableEmptyState
            icon={<Database className="h-full w-full" />}
            title="No datasets yet"
            description="Create a dataset to start batch evaluations"
          />
        )}

        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <>
            <CardListToolbar
              searchPlaceholder="Filter datasets..."
              searchValue={filter}
              onSearchChange={handleSearch}
              isPending={isFetching}
              onReset={handleReset}
              isFiltered={hasActiveFilters}
            />
            <DatasetList data={data} />
            <CardListPagination
              page={page}
              pageSize={pageSize}
              totalCount={totalCount}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
              isPending={isFetching}
            />
          </>
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
