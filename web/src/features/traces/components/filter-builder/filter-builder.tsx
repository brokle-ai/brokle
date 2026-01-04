'use client'

import { useMemo } from 'react'
import {
  FilterBuilder as SharedFilterBuilder,
  type FilterCondition as SharedFilterCondition,
} from '@/components/shared/filter-builder'
import { traceFilterColumns } from '../../config/filter-columns'
import type { FilterCondition } from '../../api/traces-api'

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
 * Trace-specific filter builder that wraps the shared FilterBuilder
 * with trace column definitions and option mappings
 */
export function FilterBuilder({
  filters,
  onApply,
  filterOptions = {},
  disabled = false,
  maxFilters = 20,
}: TraceFilterBuilderProps) {
  // Convert trace filter options to shared format
  const sharedFilterOptions = useMemo(
    () => ({
      models: filterOptions.models,
      providers: filterOptions.providers,
      services: filterOptions.services,
      environments: filterOptions.environments,
    }),
    [filterOptions]
  )

  // Handle apply with type conversion
  const handleApply = (sharedFilters: SharedFilterCondition[]) => {
    // SharedFilterCondition is compatible with FilterCondition
    onApply(sharedFilters as FilterCondition[])
  }

  return (
    <SharedFilterBuilder
      columns={traceFilterColumns}
      filters={filters as SharedFilterCondition[]}
      onApply={handleApply}
      filterOptions={sharedFilterOptions}
      disabled={disabled}
      maxFilters={maxFilters}
      title="Filter Builder"
      emptyMessage="No filters applied"
    />
  )
}
