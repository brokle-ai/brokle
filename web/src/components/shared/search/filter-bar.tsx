'use client'

import * as React from 'react'
import { X, Filter, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export interface FilterOption {
  key: string
  label: string
  type: 'select' | 'multiselect' | 'date' | 'daterange' | 'text'
  options?: Array<{ value: string; label: string }>
  placeholder?: string
}

export interface ActiveFilter {
  key: string
  label: string
  value: string | string[]
  displayValue: string
}

interface FilterBarProps extends React.HTMLAttributes<HTMLDivElement> {
  filters: FilterOption[]
  activeFilters: ActiveFilter[]
  onFilterAdd?: (filterKey: string) => void
  onFilterRemove?: (filterKey: string) => void
  onFiltersClear?: () => void
  showAddButton?: boolean
  maxDisplayedFilters?: number
}

export function FilterBar({
  filters,
  activeFilters,
  onFilterAdd,
  onFilterRemove,
  onFiltersClear,
  showAddButton = true,
  maxDisplayedFilters = 5,
  className,
  ...props
}: FilterBarProps) {
  const availableFilters = filters.filter(
    (filter) => !activeFilters.find((active) => active.key === filter.key)
  )

  const displayedFilters = activeFilters.slice(0, maxDisplayedFilters)
  const remainingCount = activeFilters.length - maxDisplayedFilters

  if (activeFilters.length === 0 && !showAddButton) {
    return null
  }

  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-2',
        className
      )}
      {...props}
    >
      {/* Active Filters */}
      {displayedFilters.map((filter) => (
        <Badge
          key={filter.key}
          variant='secondary'
          className='gap-1 pr-1 text-xs'
        >
          <span className='font-medium'>{filter.label}:</span>
          <span>{filter.displayValue}</span>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => onFilterRemove?.(filter.key)}
            className='h-4 w-4 p-0 hover:bg-transparent'
          >
            <X className='h-3 w-3' />
          </Button>
        </Badge>
      ))}

      {/* Remaining filters count */}
      {remainingCount > 0 && (
        <Badge variant='secondary' className='text-xs'>
          +{remainingCount} more
        </Badge>
      )}

      {/* Add Filter Button */}
      {showAddButton && availableFilters.length > 0 && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='outline'
              size='sm'
              className='h-6 gap-1 text-xs'
            >
              <Plus className='h-3 w-3' />
              Add Filter
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='start'>
            {availableFilters.map((filter) => (
              <DropdownMenuItem
                key={filter.key}
                onClick={() => onFilterAdd?.(filter.key)}
              >
                <Filter className='mr-2 h-4 w-4' />
                {filter.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )}

      {/* Clear All Filters */}
      {activeFilters.length > 0 && (
        <Button
          variant='ghost'
          size='sm'
          onClick={onFiltersClear}
          className='h-6 gap-1 text-xs text-muted-foreground hover:text-foreground'
        >
          <X className='h-3 w-3' />
          Clear All
        </Button>
      )}
    </div>
  )
}

// Hook for managing filter state
export function useFilters(initialFilters: ActiveFilter[] = []) {
  const [activeFilters, setActiveFilters] = React.useState<ActiveFilter[]>(initialFilters)

  const addFilter = React.useCallback((filter: ActiveFilter) => {
    setActiveFilters((prev) => {
      const existing = prev.find((f) => f.key === filter.key)
      if (existing) {
        return prev.map((f) => (f.key === filter.key ? filter : f))
      }
      return [...prev, filter]
    })
  }, [])

  const removeFilter = React.useCallback((filterKey: string) => {
    setActiveFilters((prev) => prev.filter((f) => f.key !== filterKey))
  }, [])

  const clearFilters = React.useCallback(() => {
    setActiveFilters([])
  }, [])

  const updateFilter = React.useCallback((filterKey: string, value: string | string[], displayValue: string) => {
    setActiveFilters((prev) =>
      prev.map((f) =>
        f.key === filterKey
          ? { ...f, value, displayValue }
          : f
      )
    )
  }, [])

  return {
    activeFilters,
    addFilter,
    removeFilter,
    clearFilters,
    updateFilter,
    setActiveFilters,
  }
}