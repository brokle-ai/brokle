import { LayoutDashboard, Search, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { DataTableEmptyState } from '@/components/data-table'
import { DashboardCard } from './dashboard-card'
import { DashboardsDialogs } from './dashboards-dialogs'
import type { Dashboard } from '../types'

interface DashboardsCardListProps {
  data: Dashboard[]
  totalCount: number
  searchValue: string
  onSearchChange: (value: string) => void
  isFetching?: boolean
}

function CardListToolbar({
  searchValue,
  onSearchChange,
  isFetching,
}: {
  searchValue: string
  onSearchChange: (value: string) => void
  isFetching?: boolean
}) {
  const isFiltered = searchValue.length > 0
  return (
    <div className="flex items-center gap-2 mb-4">
      <div className="relative flex-1 max-w-sm">
        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Filter dashboards..."
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-8 h-8"
        />
      </div>
      {isFiltered && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onSearchChange('')}
          className="h-8 px-2"
        >
          Reset
          <X className="ml-1 h-4 w-4" />
        </Button>
      )}
      {isFetching && (
        <span className="text-xs text-muted-foreground">Loading…</span>
      )}
    </div>
  )
}

export function DashboardsCardList({
  data,
  totalCount,
  searchValue,
  onSearchChange,
  isFetching,
}: DashboardsCardListProps) {
  const hasActiveFilters = searchValue.length > 0
  const isEmptyProject = totalCount === 0 && !hasActiveFilters

  if (isEmptyProject) {
    return (
      <>
        <DataTableEmptyState
          icon={<LayoutDashboard className="h-full w-full" />}
          title="No dashboards yet"
          description="Create a dashboard to visualize your observability data"
        />
        <DashboardsDialogs />
      </>
    )
  }

  return (
    <>
      <CardListToolbar
        searchValue={searchValue}
        onSearchChange={onSearchChange}
        isFetching={isFetching}
      />
      {data.length === 0 ? (
        <div className="rounded-lg border p-12 text-center">
          <p className="text-sm text-muted-foreground">
            No dashboards match your filter
          </p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data.map((dashboard) => (
            <DashboardCard key={dashboard.id} dashboard={dashboard} />
          ))}
        </div>
      )}
      <DashboardsDialogs />
    </>
  )
}
