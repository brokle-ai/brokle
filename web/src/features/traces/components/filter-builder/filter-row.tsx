'use client'

import { useMemo } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { FilterCondition, FilterOperator } from '../../api/traces-api'
import {
  traceFilterColumns,
  operatorLabels,
  getOperatorsForType,
  operatorRequiresValue,
  operatorAcceptsMultiple,
  type ColumnDefinition,
} from '../../config/filter-columns'

interface FilterRowProps {
  filter: FilterCondition
  onUpdate: (updates: Partial<FilterCondition>) => void
  onRemove: () => void
  filterOptions?: {
    models?: string[]
    providers?: string[]
    services?: string[]
    environments?: string[]
  }
  disabled?: boolean
  isFirst?: boolean
}

export function FilterRow({
  filter,
  onUpdate,
  onRemove,
  filterOptions = {},
  disabled = false,
  isFirst = false,
}: FilterRowProps) {
  const column = useMemo(() => {
    return traceFilterColumns.find((c) => c.id === filter.column)
  }, [filter.column])

  const operators = useMemo(() => {
    if (!column) return []
    return column.operators || getOperatorsForType(column.type)
  }, [column])

  const categoryOptions = useMemo(() => {
    if (!column) return []
    if (column.options) return column.options
    switch (column.id) {
      case 'model_name':
        return (filterOptions.models || []).map((v) => ({
          value: v,
          label: v,
        }))
      case 'provider_name':
        return (filterOptions.providers || []).map((v) => ({
          value: v,
          label: v,
        }))
      case 'service_name':
        return (filterOptions.services || []).map((v) => ({
          value: v,
          label: v,
        }))
      default:
        return []
    }
  }, [column, filterOptions])

  // Check if value input should be shown
  const showValueInput = operatorRequiresValue(filter.operator)

  // Handle column change
  const handleColumnChange = (columnId: string) => {
    const newColumn = traceFilterColumns.find((c) => c.id === columnId)
    const newOperators = newColumn
      ? newColumn.operators || getOperatorsForType(newColumn.type)
      : []
    const defaultOperator = newOperators[0] || '='

    onUpdate({
      column: columnId,
      operator: defaultOperator,
      value: '',
    })
  }

  // Handle operator change
  const handleOperatorChange = (operator: FilterOperator) => {
    // Clear value if new operator doesn't require one
    const updates: Partial<FilterCondition> = { operator }
    if (!operatorRequiresValue(operator)) {
      updates.value = null
    } else if (filter.value === null) {
      updates.value = ''
    }
    onUpdate(updates)
  }

  return (
    <div className="flex items-center gap-2 group">
      {/* Logic connector (AND for all except first) */}
      <div className="w-12 text-xs text-muted-foreground text-center">
        {isFirst ? 'Where' : 'And'}
      </div>

      {/* Column selector */}
      <Select
        value={filter.column}
        onValueChange={handleColumnChange}
        disabled={disabled}
      >
        <SelectTrigger className="w-[180px] h-8">
          <SelectValue placeholder="Select column" />
        </SelectTrigger>
        <SelectContent>
          {traceFilterColumns
            .filter((c) => c.filterable)
            .map((col) => (
              <SelectItem key={col.id} value={col.id}>
                {col.label}
              </SelectItem>
            ))}
        </SelectContent>
      </Select>

      {/* Operator selector */}
      <Select
        value={filter.operator}
        onValueChange={(v) => handleOperatorChange(v as FilterOperator)}
        disabled={disabled || !filter.column}
      >
        <SelectTrigger className="w-[150px] h-8">
          <SelectValue placeholder="Operator" />
        </SelectTrigger>
        <SelectContent>
          {operators.map((op) => (
            <SelectItem key={op} value={op}>
              {operatorLabels[op] || op}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Value input */}
      {showValueInput && (
        <div className="flex-1 min-w-[150px]">
          {column?.type === 'category' || categoryOptions.length > 0 ? (
            <Select
              value={String(filter.value || '')}
              onValueChange={(v) => onUpdate({ value: v })}
              disabled={disabled}
            >
              <SelectTrigger className="h-8">
                <SelectValue placeholder="Select value" />
              </SelectTrigger>
              <SelectContent>
                {categoryOptions.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : column?.type === 'number' ||
            column?.type === 'duration' ||
            column?.type === 'cost' ? (
            <Input
              type="number"
              value={filter.value ?? ''}
              onChange={(e) => {
                const value = e.target.value
                onUpdate({ value: value === '' ? null : Number(value) })
              }}
              placeholder={`Enter ${column?.unit || 'value'}`}
              className="h-8"
              disabled={disabled}
            />
          ) : column?.type === 'datetime' ? (
            <Input
              type="datetime-local"
              value={filter.value ?? ''}
              onChange={(e) => onUpdate({ value: e.target.value })}
              className="h-8"
              disabled={disabled}
            />
          ) : (
            <Input
              type="text"
              value={filter.value ?? ''}
              onChange={(e) => onUpdate({ value: e.target.value })}
              placeholder={
                operatorAcceptsMultiple(filter.operator)
                  ? 'Value 1, Value 2, ...'
                  : 'Enter value'
              }
              className="h-8"
              disabled={disabled}
            />
          )}
        </div>
      )}

      {/* Spacer for operators that don't need value */}
      {!showValueInput && <div className="flex-1 min-w-[150px]" />}

      {/* Remove button */}
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={onRemove}
        disabled={disabled}
      >
        <X className="h-4 w-4" />
        <span className="sr-only">Remove filter</span>
      </Button>
    </div>
  )
}
