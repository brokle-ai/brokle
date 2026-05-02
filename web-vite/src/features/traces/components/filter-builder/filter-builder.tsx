import { useMemo } from 'react'
import {
  FilterBuilder as SharedFilterBuilder,
  type FilterCondition,
} from '@/components/shared/filter-builder'
import { traceFilterColumns } from '../../config/filter-columns'

interface TraceFilterBuilderProps {
  filters: FilterCondition[]
  onApply: (filters: FilterCondition[]) => void
  filterOptions?: {
    models?: string[]
    providers?: string[]
    services?: string[]
    environments?: string[]
  }
  disabled?: boolean
  maxFilters?: number
}

/**
 * Trace-specific wrapper around the shared FilterBuilder. Owns the
 * column catalogue (`traceFilterColumns`) and the dynamic option
 * inputs that drive autocomplete on string columns. The shared
 * component handles row layout, operator selection, and value-input
 * dispatch.
 */
export function FilterBuilder({
  filters,
  onApply,
  filterOptions = {},
  disabled = false,
  maxFilters = 20,
}: TraceFilterBuilderProps) {
  const sharedFilterOptions = useMemo(
    () => ({
      models: filterOptions.models,
      providers: filterOptions.providers,
      services: filterOptions.services,
      environments: filterOptions.environments,
    }),
    [filterOptions],
  )

  return (
    <SharedFilterBuilder
      columns={traceFilterColumns}
      filters={filters}
      onApply={onApply}
      filterOptions={sharedFilterOptions}
      disabled={disabled}
      maxFilters={maxFilters}
      title="Filter Builder"
      emptyMessage="No filters applied"
    />
  )
}
