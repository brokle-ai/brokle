'use client'

import { useCallback, useMemo } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { DataTableViewOptions, DataTableFacetedFilter } from '@/components/data-table'
import { statuses } from '../data/constants'
import {
  FilterBuilder,
  SearchBar,
  FilterPresetsDrawer,
} from './filter-builder'
import { useFilterState } from '../hooks/use-filter-state'
import type { FilterPreset } from '../api/traces-api'

type DataTableToolbarProps<TData> = {
  table: Table<TData>
  isPending?: boolean
  onReset?: () => void
  filterOptions?: {
    models?: string[]
    providers?: string[]
    services?: string[]
    environments?: string[]
  }
  onFiltersChange?: (params: {
    filters: string
    search?: string
    searchTypes?: string[]
  }) => void
}

export function DataTableToolbar<TData>({
  table,
  isPending = false,
  onReset,
  filterOptions,
  onFiltersChange,
}: DataTableToolbarProps<TData>) {
  const {
    filters,
    searchQuery,
    searchTypes,
    addFilter,
    updateFilter,
    removeFilter,
    clearFilters,
    setSearchQuery,
    setSearchTypes,
    hasFilters,
    hasSearch,
    toPresetFormat,
    loadFromPreset,
  } = useFilterState({
    syncWithUrl: false, // We handle URL sync at page level
  })

  const isFiltered = useMemo(() => {
    const hasTableFilters =
      table.getState().columnFilters.length > 0 ||
      table.getState().globalFilter
    return hasTableFilters || hasFilters || hasSearch
  }, [
    table.getState().columnFilters.length,
    table.getState().globalFilter,
    hasFilters,
    hasSearch,
  ])

  const handleReset = useCallback(() => {
    clearFilters()
    onReset?.()
  }, [clearFilters, onReset])

  const handleApplyFilters = useCallback(() => {
    if (onFiltersChange) {
      const presetData = toPresetFormat()
      const filterStrings = presetData.filters.map((f) => {
        const value = Array.isArray(f.value)
          ? f.value.join(',')
          : String(f.value ?? '')
        return `${f.column} ${f.operator} ${value}`
      })
      onFiltersChange({
        filters: filterStrings.join(' AND '),
        search: presetData.search_query,
        searchTypes: presetData.search_types,
      })
    }
  }, [onFiltersChange, toPresetFormat])

  const handleApplyPreset = useCallback(
    (preset: FilterPreset) => {
      loadFromPreset({
        filters: preset.filters,
        search_query: preset.search_query,
        search_types: preset.search_types,
      })
      if (onFiltersChange) {
        const filterStrings = (preset.filters || []).map((f) => {
          const value = Array.isArray(f.value)
            ? f.value.join(',')
            : String(f.value ?? '')
          return `${f.column} ${f.operator} ${value}`
        })
        onFiltersChange({
          filters: filterStrings.join(' AND '),
          search: preset.search_query,
          searchTypes: preset.search_types,
        })
      }
    },
    [loadFromPreset, onFiltersChange]
  )

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2">
        <SearchBar
          value={searchQuery}
          onChange={setSearchQuery}
          searchTypes={searchTypes}
          onSearchTypesChange={setSearchTypes}
          placeholder="Search traces..."
          disabled={isPending}
          className="w-[250px] lg:w-[350px]"
        />

        <div className="flex gap-x-2">
          {/* Status faceted filter (kept for backwards compatibility) */}
          {table.getColumn('status_code') && (
            <DataTableFacetedFilter
              column={table.getColumn('status_code')}
              title="Status"
              options={statuses}
            />
          )}

          <FilterBuilder
            filters={filters}
            onAddFilter={addFilter}
            onUpdateFilter={updateFilter}
            onRemoveFilter={removeFilter}
            onClearFilters={clearFilters}
            onApply={handleApplyFilters}
            filterOptions={filterOptions}
            disabled={isPending}
          />

          <FilterPresetsDrawer
            currentFilters={filters}
            currentSearchQuery={searchQuery}
            currentSearchTypes={searchTypes}
            onApplyPreset={handleApplyPreset}
            tableName="traces"
          />
        </div>

        {isFiltered && (
          <Button
            variant="ghost"
            onClick={handleReset}
            className="h-8 px-2 lg:px-3"
            disabled={isPending}
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
