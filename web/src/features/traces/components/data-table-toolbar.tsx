'use client'

import { useCallback } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { DataTableViewOptions } from '@/components/data-table'
import {
  FilterBuilder,
  SearchBar,
  FilterPresetsDrawer,
} from './filter-builder'
import type { UseTracesTableStateReturn } from '../hooks/use-traces-table-state'
import type { FilterPreset, FilterCondition } from '../api/traces-api'

type DataTableToolbarProps<TData> = {
  table: Table<TData>
  tableState: UseTracesTableStateReturn
  filterOptions?: {
    models?: string[]
    providers?: string[]
    services?: string[]
    environments?: string[]
  }
}

export function DataTableToolbar<TData>({
  table,
  tableState,
  filterOptions,
}: DataTableToolbarProps<TData>) {
  const handleApplyFilters = useCallback(
    (filters: FilterCondition[]) => {
      tableState.setFilters(filters)
    },
    [tableState]
  )

  const handleSearchChange = useCallback(
    (search: string, searchType?: string) => {
      tableState.setSearch(search, searchType)
    },
    [tableState]
  )

  const handleApplyPreset = useCallback(
    (preset: FilterPreset) => {
      tableState.setFilters(preset.filters || [])
      if (preset.search_query) {
        // Pass search_types[0] as the URL param accepts a single value
        tableState.setSearch(preset.search_query, preset.search_types?.[0])
      }
    },
    [tableState]
  )

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2">
        <SearchBar
          value={tableState.search || ''}
          onChange={handleSearchChange}
          searchType={tableState.searchType}
          placeholder="Search traces..."
          className="w-[250px] lg:w-[350px]"
        />

        <div className="flex gap-x-2">
          <FilterBuilder
            filters={tableState.filters}
            onApply={handleApplyFilters}
            filterOptions={filterOptions}
          />

          <FilterPresetsDrawer
            currentFilters={tableState.filters}
            currentSearchQuery={tableState.search}
            currentSearchType={tableState.searchType}
            onApplyPreset={handleApplyPreset}
            tableName="traces"
          />
        </div>

        {tableState.hasActiveFilters && (
          <Button
            variant="ghost"
            onClick={tableState.resetAll}
            className="h-8 px-2 lg:px-3"
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
