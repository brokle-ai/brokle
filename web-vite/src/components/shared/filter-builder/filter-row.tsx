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

    const optionKey = columnOptionMapping[column.id]
    if (optionKey && filterOptions[optionKey]) {
      return (filterOptions[optionKey] || []).map((v) => ({
        value: v,
        label: v,
      }))
    }

    const defaultKey = column.id.replace(/_name$/, 's')
    if (filterOptions[defaultKey]) {
      return (filterOptions[defaultKey] || []).map((v) => ({
        value: v,
        label: v,
      }))
    }

    return []
  }, [column, filterOptions, columnOptionMapping])

  const isMultiValueOperator = operatorAcceptsMultiple(filter.operator)
  const hasCategoryOptions = column?.type === 'category' || categoryOptions.length > 0

  const getMultiSelectValue = useCallback((): string[] => {
    if (Array.isArray(filter.value)) return filter.value
    if (filter.value === null || filter.value === '') return []
    return [String(filter.value)]
  }, [filter.value])

  const formatValueForInput = useCallback((value: FilterCondition['value']): string => {
    if (Array.isArray(value)) return value.join(', ')
    return String(value ?? '')
  }, [])

  const parseInputToValue = useCallback((input: string, isMultiOp: boolean): FilterCondition['value'] => {
    if (!isMultiOp) return input || null
    const values = input.split(',').map(v => v.trim()).filter(Boolean)
    if (values.length === 0) return null
    return values
  }, [])

  const showValueInput = operatorRequiresValue(filter.operator)

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

  const handleOperatorChange = (operator: FilterOperator) => {
    const updates: Partial<FilterCondition> = { operator }
    const isEnteringMultiValue = operatorAcceptsMultiple(operator)
    const wasMultiValue = operatorAcceptsMultiple(filter.operator)

    if (!operatorRequiresValue(operator)) {
      updates.value = null
    } else if (filter.value === null || filter.value === '') {
      updates.value = null
    } else if (isEnteringMultiValue && !wasMultiValue) {
      updates.value = Array.isArray(filter.value)
        ? filter.value
        : [String(filter.value)]
    } else if (!isEnteringMultiValue && wasMultiValue) {
      updates.value = Array.isArray(filter.value)
        ? (filter.value[0] ?? '')
        : filter.value
    }

    onUpdate(updates)
  }

  const filterableColumns = columns.filter((c) => c.filterable)

  return (
    <div className="flex items-center gap-2 group">
      <div className="w-12 text-xs text-muted-foreground text-center">
        {isFirst ? 'Where' : 'And'}
      </div>

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

      {showValueInput && (
        <div className="flex-1 min-w-[150px]">
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
          ) : hasCategoryOptions ? (
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

      {!showValueInput && <div className="flex-1 min-w-[150px]" />}

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
