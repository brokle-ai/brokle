import { type Table } from '@tanstack/react-table'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { DataTableViewOptions } from '@/components/shared/tables/data-table-view-options'
import { DataTableFacetedFilter } from '@/components/shared/tables/data-table-faceted-filter'
import { TracesFilterBar, type TracesFilterValue } from './traces-filter-bar'
import { statuses } from '../data/constants'
import type { TraceListItem } from '../api/types'

interface TracesToolbarProps {
  table: Table<TraceListItem>
  filterValue: TracesFilterValue
  onFilterChange: (next: TracesFilterValue) => void
  /**
   * Facets sourced from the current page of results so the faceted
   * filter menu only offers values actually present. Models + providers
   * are the only useful facets the list endpoint surfaces today.
   */
  modelFacets?: string[]
  providerFacets?: string[]
}

/**
 * Toolbar pairing the free-text/range filter bar (URL-backed) with
 * client-side faceted filters on `status_code` / `model_name` /
 * `provider_name` and the column-visibility menu on the right.
 */
export function TracesToolbar({
  table,
  filterValue,
  onFilterChange,
  modelFacets = [],
  providerFacets = [],
}: TracesToolbarProps) {
  const statusColumn = table.getColumn('status_code')
  const modelColumn = table.getColumn('model_name')
  const providerColumn = table.getColumn('provider_name')

  const isFiltered =
    table.getState().columnFilters.length > 0 ||
    !!filterValue.q ||
    !!filterValue.status ||
    !!filterValue.model ||
    filterValue.range !== 'all'

  return (
    <div className="flex items-center justify-between gap-2">
      <div className="flex flex-1 flex-wrap items-center gap-2">
        <TracesFilterBar value={filterValue} onChange={onFilterChange} />

        {statusColumn && (
          <DataTableFacetedFilter
            column={statusColumn}
            title="Status"
            options={statuses.map((s) => ({
              label: s.label,
              value: s.value,
              icon: s.icon,
            }))}
          />
        )}

        {modelColumn && modelFacets.length > 0 && (
          <DataTableFacetedFilter
            column={modelColumn}
            title="Model"
            options={modelFacets.map((m) => ({ label: m, value: m }))}
          />
        )}

        {providerColumn && providerFacets.length > 0 && (
          <DataTableFacetedFilter
            column={providerColumn}
            title="Provider"
            options={providerFacets.map((p) => ({ label: p, value: p }))}
          />
        )}

        {isFiltered && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 lg:px-3"
            onClick={() => {
              table.resetColumnFilters()
              onFilterChange({ range: 'all' })
            }}
          >
            Reset
            <Cross2Icon className="ms-2 h-4 w-4" />
          </Button>
        )}
      </div>

      <DataTableViewOptions table={table} />
    </div>
  )
}
