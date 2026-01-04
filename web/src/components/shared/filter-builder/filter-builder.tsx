'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { nanoid } from 'nanoid'
import { Plus, Filter, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { FilterRow } from './filter-row'
import { operatorRequiresValue } from './utils'
import type { FilterBuilderProps, FilterCondition, FilterOperator } from './types'

export function FilterBuilder({
  columns,
  filters,
  onApply,
  filterOptions = {},
  disabled = false,
  maxFilters = 20,
  title = 'Filter Builder',
  emptyMessage = 'No filters applied',
}: FilterBuilderProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [localFilters, setLocalFilters] = useState<FilterCondition[]>(filters)

  useEffect(() => {
    setLocalFilters(filters)
  }, [filters])

  const filterCount = filters.length
  const localFilterCount = localFilters.length
  const canAddMore = localFilterCount < maxFilters

  const badgeVariant = useMemo(() => {
    if (filterCount === 0) return 'outline'
    return 'default'
  }, [filterCount])

  const addFilter = useCallback(() => {
    setLocalFilters((prev) => [
      ...prev,
      {
        id: nanoid(8),
        column: '',
        operator: '=' as FilterOperator,
        value: null,
      },
    ])
  }, [])

  const updateFilter = useCallback((id: string, updates: Partial<FilterCondition>) => {
    setLocalFilters((prev) =>
      prev.map((f) => (f.id === id ? { ...f, ...updates } : f))
    )
  }, [])

  const removeFilter = useCallback((id: string) => {
    setLocalFilters((prev) => prev.filter((f) => f.id !== id))
  }, [])

  const clearLocalFilters = useCallback(() => {
    setLocalFilters([])
  }, [])

  const handleApply = useCallback(() => {
    // Only keep valid filters
    const validFilters = localFilters.filter((f) => {
      // Must have column selected
      if (!f.column) return false
      // Operators that don't require value (IS EMPTY, EXISTS, etc.) are always valid
      if (!operatorRequiresValue(f.operator)) return true
      // Operators that require value must have non-null, non-empty value
      if (f.value === null || f.value === '') return false
      // Empty arrays are invalid (IN [] matches nothing, NOT IN [] is a no-op)
      if (Array.isArray(f.value) && f.value.length === 0) return false
      return true
    })
    onApply(validFilters)
    setIsOpen(false)
  }, [localFilters, onApply])

  const handleClear = useCallback(() => {
    setLocalFilters([])
    onApply([])
    setIsOpen(false)
  }, [onApply])

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-8 border-dashed"
          disabled={disabled}
        >
          <Filter className="mr-2 h-4 w-4" />
          Filters
          {filterCount > 0 && (
            <Badge variant={badgeVariant} className="ml-2 px-1.5">
              {filterCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[700px] p-4" align="start">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4" />
              <h4 className="font-medium">{title}</h4>
              {localFilterCount > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {localFilterCount} {localFilterCount === 1 ? 'filter' : 'filters'}
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-2">
              {localFilterCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={clearLocalFilters}
                  disabled={disabled}
                >
                  <RotateCcw className="mr-1 h-3 w-3" />
                  Clear all
                </Button>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {localFilters.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
                <Filter className="h-8 w-8 mb-2 opacity-50" />
                <p className="text-sm">{emptyMessage}</p>
                <p className="text-xs">Add a filter to narrow down results</p>
              </div>
            ) : (
              localFilters.map((filter, index) => (
                <FilterRow
                  key={filter.id}
                  columns={columns}
                  filter={filter}
                  onUpdate={(updates) => updateFilter(filter.id, updates)}
                  onRemove={() => removeFilter(filter.id)}
                  filterOptions={filterOptions}
                  disabled={disabled}
                  isFirst={index === 0}
                />
              ))
            )}
          </div>

          {canAddMore && (
            <>
              <Separator />
              <Button
                variant="ghost"
                size="sm"
                className="w-full h-8 border border-dashed"
                onClick={addFilter}
                disabled={disabled}
              >
                <Plus className="mr-2 h-4 w-4" />
                Add filter
              </Button>
            </>
          )}

          <Separator />
          <div className="flex justify-between">
            {filterCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                className="h-8"
                onClick={handleClear}
                disabled={disabled}
              >
                Clear all filters
              </Button>
            )}
            <div className="flex-1" />
            <Button
              size="sm"
              className="h-8"
              onClick={handleApply}
              disabled={disabled}
            >
              Apply filters
            </Button>
          </div>

          {!canAddMore && (
            <p className="text-xs text-muted-foreground text-center">
              Maximum of {maxFilters} filters reached
            </p>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
