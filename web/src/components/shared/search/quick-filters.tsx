'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface QuickFilter {
  key: string
  label: string
  count?: number
  description?: string
  icon?: React.ComponentType<{ className?: string }>
}

interface QuickFiltersProps extends React.HTMLAttributes<HTMLDivElement> {
  filters: QuickFilter[]
  activeFilter?: string
  onFilterChange?: (filterKey: string | null) => void
  variant?: 'tabs' | 'buttons' | 'badges'
  showCounts?: boolean
  allowDeselect?: boolean
}

export function QuickFilters({
  filters,
  activeFilter,
  onFilterChange,
  variant = 'buttons',
  showCounts = true,
  allowDeselect = true,
  className,
  ...props
}: QuickFiltersProps) {
  const handleFilterClick = (filterKey: string) => {
    if (activeFilter === filterKey && allowDeselect) {
      onFilterChange?.(null)
    } else {
      onFilterChange?.(filterKey)
    }
  }

  if (variant === 'tabs') {
    return (
      <div
        className={cn(
          'inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground',
          className
        )}
        {...props}
      >
        {filters.map((filter) => {
          const isActive = activeFilter === filter.key
          const Icon = filter.icon

          return (
            <button
              key={filter.key}
              onClick={() => handleFilterClick(filter.key)}
              className={cn(
                'inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
                isActive && 'bg-background text-foreground shadow-sm'
              )}
            >
              {Icon && <Icon className='mr-2 h-4 w-4' />}
              {filter.label}
              {showCounts && filter.count !== undefined && (
                <Badge
                  variant='secondary'
                  className={cn(
                    'ml-2 h-5 px-1.5 text-xs',
                    isActive ? 'bg-muted' : 'bg-background'
                  )}
                >
                  {filter.count}
                </Badge>
              )}
            </button>
          )
        })}
      </div>
    )
  }

  if (variant === 'badges') {
    return (
      <div
        className={cn('flex flex-wrap items-center gap-2', className)}
        {...props}
      >
        {filters.map((filter) => {
          const isActive = activeFilter === filter.key
          const Icon = filter.icon

          return (
            <Button
              key={filter.key}
              variant={isActive ? 'default' : 'outline'}
              size='sm'
              onClick={() => handleFilterClick(filter.key)}
              className='h-7 gap-1 text-xs'
            >
              {Icon && <Icon className='h-3 w-3' />}
              {filter.label}
              {showCounts && filter.count !== undefined && (
                <Badge
                  variant='secondary'
                  className='ml-1 h-4 px-1 text-xs'
                >
                  {filter.count}
                </Badge>
              )}
            </Button>
          )
        })}
      </div>
    )
  }

  // Default buttons variant
  return (
    <div
      className={cn('flex flex-wrap items-center gap-2', className)}
      {...props}
    >
      {filters.map((filter) => {
        const isActive = activeFilter === filter.key
        const Icon = filter.icon

        return (
          <Button
            key={filter.key}
            variant={isActive ? 'default' : 'outline'}
            size='sm'
            onClick={() => handleFilterClick(filter.key)}
            className='gap-2'
            title={filter.description}
          >
            {Icon && <Icon className='h-4 w-4' />}
            {filter.label}
            {showCounts && filter.count !== undefined && (
              <Badge
                variant={isActive ? 'secondary' : 'outline'}
                className='ml-1 text-xs'
              >
                {filter.count}
              </Badge>
            )}
          </Button>
        )
      })}
    </div>
  )
}

// Hook for managing quick filter state
export function useQuickFilter(initialFilter?: string) {
  const [activeFilter, setActiveFilter] = React.useState<string | null>(initialFilter || null)

  const handleFilterChange = React.useCallback((filterKey: string | null) => {
    setActiveFilter(filterKey)
  }, [])

  const clearFilter = React.useCallback(() => {
    setActiveFilter(null)
  }, [])

  return {
    activeFilter,
    setActiveFilter: handleFilterChange,
    clearFilter,
  }
}