'use client'

import { useState, useEffect, useCallback } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Search, X } from 'lucide-react'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DataTableViewOptions } from '@/components/data-table'
import type { UseDatasetsTableStateReturn } from '../../hooks/use-datasets-table-state'

type DatasetsToolbarProps<TData> = {
  table: Table<TData>
  tableState: UseDatasetsTableStateReturn
  onReset?: () => void
}

export function DatasetsToolbar<TData>({
  table,
  tableState,
  onReset,
}: DatasetsToolbarProps<TData>) {
  // Read directly from tableState (NOT table.getState().globalFilter)
  // This bypasses the table's re-render cycle that causes focus loss
  const searchValue = tableState.search || ''

  // Local state for debounced input
  const [localSearch, setLocalSearch] = useState(searchValue)

  // Sync from URL when it changes externally (reset, browser navigation)
  useEffect(() => {
    setLocalSearch(searchValue)
  }, [searchValue])

  // Debounced sync to URL - call tableState.setSearch() directly
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

  const isFiltered = searchValue.length > 0

  const handleReset = () => {
    setLocalSearch('')
    onReset?.()
  }

  return (
    <div className='flex items-center justify-between'>
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2'>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder='Search datasets...'
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className='h-8 w-[150px] lg:w-[250px] pl-8 pr-8'
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
        {isFiltered && (
          <Button
            variant='ghost'
            onClick={handleReset}
            className='h-8 px-2 lg:px-3'
          >
            Reset
            <Cross2Icon className='ms-2 h-4 w-4' />
          </Button>
        )}
      </div>
      <DataTableViewOptions table={table} />
    </div>
  )
}
