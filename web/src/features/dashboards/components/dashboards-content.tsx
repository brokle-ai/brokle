'use client'

import { useSearchParams } from 'next/navigation'
import { LayoutDashboard } from 'lucide-react'
import { DashboardsProvider } from '../context/dashboards-context'
import { DashboardList } from './dashboard-list'
import { DashboardsDialogs } from './dashboards-dialogs'
import { CreateDashboardDialog } from './create-dashboard-dialog'
import { useProjectDashboards } from '../hooks/use-project-dashboards'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useCardListNavigation } from '@/hooks/use-card-list-navigation'
import { PageHeader } from '@/components/layout/page-header'
import { CardListToolbar } from '@/components/card-list'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface DashboardsProps {
  projectSlug: string
}

function DashboardsInner() {
  const searchParams = useSearchParams()
  const { data, totalCount, isLoading, isFetching, error, hasProject, refetch, currentProject } =
    useProjectDashboards()
  const { filter } = useTableSearchParams(searchParams)
  const { handleSearch, handleReset } = useCardListNavigation({ searchParams })

  // Check if there are active filters
  const hasActiveFilters = !!filter

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && data.length === 0

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters

  return (
    <>
      <PageHeader title="Dashboards">
        {currentProject && <CreateDashboardDialog projectId={currentProject.id} />}
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {isInitialLoad && (
          <div className="flex flex-1 items-center justify-center py-16">
            <LoadingSpinner message="Loading dashboards..." />
          </div>
        )}

        {error && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">Failed to load dashboards</h3>
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
            icon={<LayoutDashboard className="h-full w-full" />}
            title="No dashboards yet"
            description="Create a dashboard to visualize your observability data"
          />
        )}

        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <>
            <CardListToolbar
              searchPlaceholder="Filter dashboards..."
              searchValue={filter}
              onSearchChange={handleSearch}
              isPending={isFetching}
              onReset={handleReset}
              isFiltered={hasActiveFilters}
            />
            <DashboardList data={data} />
          </>
        )}
      </div>
      <DashboardsDialogs />
    </>
  )
}

export function Dashboards({ projectSlug }: DashboardsProps) {
  return (
    <DashboardsProvider projectSlug={projectSlug}>
      <DashboardsInner />
    </DashboardsProvider>
  )
}
