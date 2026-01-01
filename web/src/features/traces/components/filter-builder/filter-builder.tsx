'use client'

import { useMemo } from 'react'
import { Plus, Filter, RotateCcw, Save, Bookmark } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { FilterRow } from './filter-row'
import type { FilterCondition, FilterOperator } from '../../api/traces-api'

interface FilterBuilderProps {
  filters: FilterCondition[]
  onAddFilter: (column?: string, operator?: FilterOperator, value?: any) => void
  onUpdateFilter: (id: string, updates: Partial<FilterCondition>) => void
  onRemoveFilter: (id: string) => void
  onClearFilters: () => void
  onApply?: () => void
  onSavePreset?: () => void
  filterOptions?: {
    models?: string[]
    providers?: string[]
    services?: string[]
    environments?: string[]
  }
  disabled?: boolean
  maxFilters?: number
}

export function FilterBuilder({
  filters,
  onAddFilter,
  onUpdateFilter,
  onRemoveFilter,
  onClearFilters,
  onApply,
  onSavePreset,
  filterOptions = {},
  disabled = false,
  maxFilters = 20,
}: FilterBuilderProps) {
  const filterCount = filters.length
  const canAddMore = filterCount < maxFilters

  const badgeVariant = useMemo(() => {
    if (filterCount === 0) return 'outline'
    return 'default'
  }, [filterCount])

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 border-dashed">
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
              <h4 className="font-medium">Filter Builder</h4>
              {filterCount > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {filterCount} {filterCount === 1 ? 'filter' : 'filters'}
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-2">
              {filterCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={onClearFilters}
                  disabled={disabled}
                >
                  <RotateCcw className="mr-1 h-3 w-3" />
                  Clear all
                </Button>
              )}
              {onSavePreset && filterCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={onSavePreset}
                  disabled={disabled}
                >
                  <Bookmark className="mr-1 h-3 w-3" />
                  Save preset
                </Button>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {filters.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
                <Filter className="h-8 w-8 mb-2 opacity-50" />
                <p className="text-sm">No filters applied</p>
                <p className="text-xs">Add a filter to narrow down results</p>
              </div>
            ) : (
              filters.map((filter, index) => (
                <FilterRow
                  key={filter.id}
                  filter={filter}
                  onUpdate={(updates) => onUpdateFilter(filter.id, updates)}
                  onRemove={() => onRemoveFilter(filter.id)}
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
                onClick={() => onAddFilter()}
                disabled={disabled}
              >
                <Plus className="mr-2 h-4 w-4" />
                Add filter
              </Button>
            </>
          )}

          {onApply && filterCount > 0 && (
            <>
              <Separator />
              <div className="flex justify-end">
                <Button
                  size="sm"
                  className="h-8"
                  onClick={onApply}
                  disabled={disabled}
                >
                  Apply filters
                </Button>
              </div>
            </>
          )}

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
