'use client'

import { useState, useEffect, useCallback } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Search, X, GitCompare } from 'lucide-react'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DataTableViewOptions } from '@/components/data-table'
import type { UseExperimentsTableStateReturn } from '../../hooks/use-experiments-table-state'
import type { ExperimentStatus } from '../../types'

const STATUS_OPTIONS: { value: ExperimentStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
]

type ExperimentsToolbarProps<TData> = {
  table: Table<TData>
  tableState: UseExperimentsTableStateReturn
  onReset?: () => void
  onCompare?: (selectedIds: string[]) => void
  selectedCount: number
}

export function ExperimentsToolbar<TData>({
  table,
  tableState,
  onReset,
  onCompare,
  selectedCount,
}: ExperimentsToolbarProps<TData>) {
  // Read directly from tableState
  const searchValue = tableState.search || ''
  const statusValue = tableState.status

  // Local state for debounced search input
  const [localSearch, setLocalSearch] = useState(searchValue)

  // Sync from URL when it changes externally (reset, browser navigation)
  useEffect(() => {
    setLocalSearch(searchValue)
  }, [searchValue])

  // Debounced sync to URL
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== searchValue) {
        tableState.setSearch(localSearch)
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [localSearch, searchValue, tableState])

  const handleClearSearch = useCallback(() => {
    setLocalSearch('')
    tableState.setSearch('')
  }, [tableState])

  const handleStatusChange = useCallback((value: string) => {
    if (value === 'all') {
      tableState.setStatus(null)
    } else {
      tableState.setStatus(value as ExperimentStatus)
    }
  }, [tableState])

  const handleReset = useCallback(() => {
    setLocalSearch('')
    onReset?.()
  }, [onReset])

  const handleCompare = useCallback(() => {
    const selectedRows = table.getFilteredSelectedRowModel().rows
    const selectedIds = selectedRows.map((row) => (row.original as { id: string }).id)
    onCompare?.(selectedIds)
  }, [table, onCompare])

  const isFiltered = tableState.hasActiveFilters

  return (
    <div className="flex items-center justify-between gap-2">
      <div className="flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2">
        {/* Search Input */}
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search experiments..."
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="h-8 w-[150px] lg:w-[250px] pl-8 pr-8"
          />
          {localSearch && (
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-0.5 top-1/2 h-7 w-7 -translate-y-1/2"
              onClick={handleClearSearch}
            >
              <X className="h-3.5 w-3.5" />
              <span className="sr-only">Clear search</span>
            </Button>
          )}
        </div>

        {/* Status Filter */}
        <Select
          value={statusValue || 'all'}
          onValueChange={handleStatusChange}
        >
          <SelectTrigger className="h-8 w-[130px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            {STATUS_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Reset Button */}
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={handleReset}
            className="h-8 px-2 lg:px-3"
          >
            Reset
            <Cross2Icon className="ms-2 h-4 w-4" />
          </Button>
        )}
      </div>

      <div className="flex items-center gap-2">
        {/* Compare Button - appears when 2+ experiments selected */}
        {selectedCount >= 2 && onCompare && (
          <Button
            variant="default"
            size="sm"
            onClick={handleCompare}
            className="h-8"
          >
            <GitCompare className="mr-2 h-4 w-4" />
            Compare ({selectedCount})
          </Button>
        )}

        <DataTableViewOptions table={table} />
      </div>
    </div>
  )
}
