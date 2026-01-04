'use client'

import { useCallback } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { QueryFilter, FilterOperator, DimensionDefinition } from '../../types'

interface FilterBuilderProps {
  /** Available dimensions for filtering */
  dimensions?: DimensionDefinition[]
  /** Current filters */
  filters: QueryFilter[]
  /** Callback when filters change */
  onFiltersChange: (filters: QueryFilter[]) => void
  /** Whether the builder is disabled */
  disabled?: boolean
  /** Additional class name */
  className?: string
}

// Operator configurations with human-readable labels
const OPERATORS: { value: FilterOperator; label: string; description: string }[] = [
  { value: 'eq', label: '=', description: 'equals' },
  { value: 'neq', label: '!=', description: 'not equals' },
  { value: 'gt', label: '>', description: 'greater than' },
  { value: 'lt', label: '<', description: 'less than' },
  { value: 'gte', label: '>=', description: 'greater than or equal' },
  { value: 'lte', label: '<=', description: 'less than or equal' },
  { value: 'contains', label: 'contains', description: 'contains text' },
  { value: 'in', label: 'in', description: 'is one of' },
]

// Get applicable operators based on column type
function getOperatorsForType(columnType: string): FilterOperator[] {
  switch (columnType) {
    case 'number':
    case 'datetime':
      return ['eq', 'neq', 'gt', 'lt', 'gte', 'lte']
    case 'string':
    default:
      return ['eq', 'neq', 'contains', 'in']
  }
}

/**
 * FilterBuilder component for creating query filters.
 *
 * Supports different operators based on dimension type:
 * - String: equals, not equals, contains, in
 * - Number: equals, not equals, gt, lt, gte, lte
 * - DateTime: equals, not equals, gt, lt, gte, lte
 */
export function FilterBuilder({
  dimensions = [],
  filters,
  onFiltersChange,
  disabled,
  className,
}: FilterBuilderProps) {
  // Add a new empty filter
  const handleAddFilter = useCallback(() => {
    const newFilter: QueryFilter = {
      field: '',
      operator: 'eq',
      value: '',
    }
    onFiltersChange([...filters, newFilter])
  }, [filters, onFiltersChange])

  // Update a filter at a specific index
  const handleUpdateFilter = useCallback(
    (index: number, updates: Partial<QueryFilter>) => {
      const updatedFilters = filters.map((filter, i) =>
        i === index ? { ...filter, ...updates } : filter
      )
      onFiltersChange(updatedFilters)
    },
    [filters, onFiltersChange]
  )

  // Remove a filter at a specific index
  const handleRemoveFilter = useCallback(
    (index: number) => {
      onFiltersChange(filters.filter((_, i) => i !== index))
    },
    [filters, onFiltersChange]
  )

  // Get dimension by ID
  const getDimension = useCallback(
    (id: string): DimensionDefinition | undefined => {
      return dimensions.find((d) => d.id === id)
    },
    [dimensions]
  )

  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center justify-between">
        <Label className="text-sm font-medium">Filters</Label>
        <Button
          variant="outline"
          size="sm"
          onClick={handleAddFilter}
          disabled={disabled || dimensions.length === 0}
          className="h-7 text-xs"
        >
          <Plus className="mr-1 h-3 w-3" />
          Add Filter
        </Button>
      </div>

      {filters.length === 0 ? (
        <p className="text-xs text-muted-foreground py-2">
          No filters applied. Click "Add Filter" to filter your data.
        </p>
      ) : (
        <div className="space-y-2">
          {filters.map((filter, index) => (
            <FilterRow
              key={index}
              filter={filter}
              dimensions={dimensions}
              getDimension={getDimension}
              onUpdate={(updates) => handleUpdateFilter(index, updates)}
              onRemove={() => handleRemoveFilter(index)}
              disabled={disabled}
            />
          ))}
        </div>
      )}

      {/* Active filters summary */}
      {filters.length > 0 && (
        <div className="flex flex-wrap gap-1.5 pt-1">
          {filters
            .filter((f) => f.field && f.value !== '')
            .map((filter, index) => (
              <Badge key={index} variant="secondary" className="text-xs">
                {filter.field} {OPERATORS.find((o) => o.value === filter.operator)?.label}{' '}
                {formatFilterValue(filter.value)}
              </Badge>
            ))}
        </div>
      )}
    </div>
  )
}

interface FilterRowProps {
  filter: QueryFilter
  dimensions: DimensionDefinition[]
  getDimension: (id: string) => DimensionDefinition | undefined
  onUpdate: (updates: Partial<QueryFilter>) => void
  onRemove: () => void
  disabled?: boolean
}

function FilterRow({
  filter,
  dimensions,
  getDimension,
  onUpdate,
  onRemove,
  disabled,
}: FilterRowProps) {
  const selectedDimension = filter.field ? getDimension(filter.field) : undefined
  const columnType = selectedDimension?.column_type || 'string'
  const applicableOperators = getOperatorsForType(columnType)

  // When field changes, reset operator if it's not applicable
  const handleFieldChange = (field: string) => {
    const dim = getDimension(field)
    const newColumnType = dim?.column_type || 'string'
    const newApplicableOperators = getOperatorsForType(newColumnType)

    // Reset operator if current one is not applicable to new type
    const newOperator = newApplicableOperators.includes(filter.operator)
      ? filter.operator
      : newApplicableOperators[0]

    onUpdate({
      field,
      operator: newOperator,
      value: '', // Reset value when field changes
    })
  }

  return (
    <div className="flex items-center gap-2 p-2 rounded-md border bg-muted/30">
      {/* Field Selector */}
      <Select
        value={filter.field}
        onValueChange={handleFieldChange}
        disabled={disabled}
      >
        <SelectTrigger className="h-8 w-36 text-xs">
          <SelectValue placeholder="Field" />
        </SelectTrigger>
        <SelectContent>
          {dimensions.map((dim) => (
            <SelectItem key={dim.id} value={dim.id} className="text-xs">
              {dim.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Operator Selector */}
      <Select
        value={filter.operator}
        onValueChange={(value) => onUpdate({ operator: value as FilterOperator })}
        disabled={disabled || !filter.field}
      >
        <SelectTrigger className="h-8 w-24 text-xs">
          <SelectValue placeholder="Op" />
        </SelectTrigger>
        <SelectContent>
          {OPERATORS.filter((op) => applicableOperators.includes(op.value)).map((op) => (
            <SelectItem key={op.value} value={op.value} className="text-xs">
              <span className="font-mono">{op.label}</span>
              <span className="ml-1 text-muted-foreground">({op.description})</span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Value Input */}
      {filter.operator === 'in' ? (
        <Input
          type="text"
          value={formatInValue(filter.value)}
          onChange={(e) => onUpdate({ value: parseInValue(e.target.value) })}
          placeholder="value1, value2, ..."
          disabled={disabled || !filter.field}
          className="h-8 flex-1 text-xs"
        />
      ) : (
        <Input
          type={columnType === 'number' ? 'number' : 'text'}
          value={filter.value as string}
          onChange={(e) =>
            onUpdate({
              value:
                columnType === 'number' ? parseFloat(e.target.value) || 0 : e.target.value,
            })
          }
          placeholder="Value"
          disabled={disabled || !filter.field}
          className="h-8 flex-1 text-xs"
        />
      )}

      {/* Remove Button */}
      <Button
        variant="ghost"
        size="icon"
        onClick={onRemove}
        disabled={disabled}
        className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </div>
  )
}

// Helper functions for formatting values

function formatFilterValue(value: unknown): string {
  if (Array.isArray(value)) {
    return value.join(', ')
  }
  return String(value)
}

function formatInValue(value: unknown): string {
  if (Array.isArray(value)) {
    return value.join(', ')
  }
  return String(value || '')
}

function parseInValue(input: string): string[] {
  return input
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
}

export type { FilterBuilderProps }
