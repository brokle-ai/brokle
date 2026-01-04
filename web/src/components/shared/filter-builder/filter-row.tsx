'use client'

import { useMemo, useCallback } from 'react'
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
import { MultiSelect } from '@/components/shared/forms/multi-select'
import {
  operatorLabels,
  getOperatorsForType,
  operatorRequiresValue,
  operatorAcceptsMultiple,
} from './utils'
import type { FilterRowProps, FilterCondition, FilterOperator } from './types'

export function FilterRow({
  columns,
  filter,
  onUpdate,
  onRemove,
  filterOptions = {},
  disabled = false,
  isFirst = false,
  columnOptionMapping = {},
}: FilterRowProps) {
  const column = useMemo(() => {
    return columns.find((c) => c.id === filter.column)
  }, [columns, filter.column])

  const operators = useMemo(() => {
    if (!column) return []
    return column.operators || getOperatorsForType(column.type)
  }, [column])

  const categoryOptions = useMemo(() => {
    if (!column) return []
    if (column.options) return column.options

    // Check if there's a mapping for this column to get dynamic options
    const optionKey = columnOptionMapping[column.id]
    if (optionKey && filterOptions[optionKey]) {
      return (filterOptions[optionKey] || []).map((v) => ({
        value: v,
        label: v,
      }))
    }

    // Default mapping by column id
    const defaultKey = column.id.replace(/_name$/, 's')
    if (filterOptions[defaultKey]) {
      return (filterOptions[defaultKey] || []).map((v) => ({
        value: v,
        label: v,
      }))
    }

    return []
  }, [column, filterOptions, columnOptionMapping])

  // Check if this is a multi-value operator (IN/NOT IN)
  const isMultiValueOperator = operatorAcceptsMultiple(filter.operator)
  const hasCategoryOptions = column?.type === 'category' || categoryOptions.length > 0

  // Helper to get current value as array for multi-select
  const getMultiSelectValue = useCallback((): string[] => {
    if (Array.isArray(filter.value)) return filter.value
    if (filter.value === null || filter.value === '') return []
    return [String(filter.value)]
  }, [filter.value])

  // Helper to format array value for text input display
  const formatValueForInput = useCallback((value: FilterCondition['value']): string => {
    if (Array.isArray(value)) return value.join(', ')
    return String(value ?? '')
  }, [])

  // Helper to parse text input to array for IN/NOT IN operators
  const parseInputToValue = useCallback((input: string, isMultiOp: boolean): FilterCondition['value'] => {
    if (!isMultiOp) return input || null
    // Parse comma-separated values for IN/NOT IN - ALWAYS return array
    const values = input.split(',').map(v => v.trim()).filter(Boolean)
    if (values.length === 0) return null
    return values  // Always return array for multi-value operators
  }, [])

  // Check if value input should be shown
  const showValueInput = operatorRequiresValue(filter.operator)

  // Handle column change
  const handleColumnChange = (columnId: string) => {
    const newColumn = columns.find((c) => c.id === columnId)
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
    const updates: Partial<FilterCondition> = { operator }
    const isEnteringMultiValue = operatorAcceptsMultiple(operator)
    const wasMultiValue = operatorAcceptsMultiple(filter.operator)

    if (!operatorRequiresValue(operator)) {
      updates.value = null
    } else if (filter.value === null || filter.value === '') {
      // Keep null for empty values - will be filtered out on apply
      updates.value = null
    } else if (isEnteringMultiValue && !wasMultiValue) {
      // Switching TO multi-value: normalize scalar to array
      updates.value = Array.isArray(filter.value)
        ? filter.value
        : [String(filter.value)]
    } else if (!isEnteringMultiValue && wasMultiValue) {
      // Switching FROM multi-value: take first element or empty
      updates.value = Array.isArray(filter.value)
        ? (filter.value[0] ?? '')
        : filter.value
    }

    onUpdate(updates)
  }

  const filterableColumns = columns.filter((c) => c.filterable)

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
          {filterableColumns.map((col) => (
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
          {/* Multi-select for category columns with IN/NOT IN */}
          {hasCategoryOptions && isMultiValueOperator ? (
            <MultiSelect
              options={categoryOptions}
              value={getMultiSelectValue()}
              onValueChange={(values) => {
                onUpdate({ value: values.length > 0 ? values : null })
              }}
              placeholder="Select values..."
              disabled={disabled}
              className="h-8"
            />
          ) : /* Single select for category columns without IN/NOT IN */
          hasCategoryOptions ? (
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
            /* Text input with comma-separated parsing for IN/NOT IN */
            <Input
              type="text"
              value={formatValueForInput(filter.value)}
              onChange={(e) => {
                const parsed = parseInputToValue(e.target.value, isMultiValueOperator)
                onUpdate({ value: parsed })
              }}
              placeholder={
                isMultiValueOperator
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
