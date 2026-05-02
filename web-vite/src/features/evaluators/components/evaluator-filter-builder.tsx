import { useCallback, useMemo, useState } from 'react'
import { nanoid } from 'nanoid'
import { Filter, Info, Plus, X } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { FilterClause, FilterOperator } from '../api/types'

// Column definition consumed by the filter builder. Limited to the
// shapes the evaluator backend understands today.
export interface EvaluatorFilterColumn {
  id: string
  label: string
  type: 'string' | 'number' | 'category'
  description?: string
  operators: FilterOperator[]
  options?: { value: string; label: string }[]
}

// Default column set for evaluator targeting. Lifted from web/'s
// reference implementation so v1 ports retain feature parity.
export const EVALUATOR_FILTER_COLUMNS: EvaluatorFilterColumn[] = [
  {
    id: 'span_name',
    label: 'Span Name',
    type: 'string',
    description: 'Name of the span (e.g. "chat-completion")',
    operators: ['equals', 'not_equals', 'contains'],
  },
  {
    id: 'span_kind',
    label: 'Span Kind',
    type: 'category',
    description: 'Type of span in the trace',
    operators: ['equals', 'not_equals'],
    options: [
      { value: 'llm', label: 'LLM' },
      { value: 'chain', label: 'Chain' },
      { value: 'agent', label: 'Agent' },
      { value: 'tool', label: 'Tool' },
      { value: 'retriever', label: 'Retriever' },
      { value: 'embedding', label: 'Embedding' },
    ],
  },
  {
    id: 'model_name',
    label: 'Model Name',
    type: 'string',
    description: 'AI model used (e.g. "gpt-4o")',
    operators: ['equals', 'not_equals', 'contains'],
  },
  {
    id: 'provider',
    label: 'Provider',
    type: 'category',
    description: 'AI provider (e.g. OpenAI, Anthropic)',
    operators: ['equals', 'not_equals'],
    options: [
      { value: 'openai', label: 'OpenAI' },
      { value: 'anthropic', label: 'Anthropic' },
      { value: 'google', label: 'Google' },
      { value: 'azure', label: 'Azure' },
      { value: 'aws', label: 'AWS Bedrock' },
    ],
  },
  {
    id: 'latency_ms',
    label: 'Latency (ms)',
    type: 'number',
    description: 'Span duration in milliseconds',
    operators: ['equals', 'gt', 'lt', 'gte', 'lte'],
  },
  {
    id: 'token_count',
    label: 'Total Tokens',
    type: 'number',
    description: 'Total token count (input + output)',
    operators: ['equals', 'gt', 'lt', 'gte', 'lte'],
  },
  {
    id: 'input_tokens',
    label: 'Input Tokens',
    type: 'number',
    description: 'Number of input/prompt tokens',
    operators: ['equals', 'gt', 'lt', 'gte', 'lte'],
  },
  {
    id: 'output_tokens',
    label: 'Output Tokens',
    type: 'number',
    description: 'Number of output/completion tokens',
    operators: ['equals', 'gt', 'lt', 'gte', 'lte'],
  },
  {
    id: 'status',
    label: 'Status',
    type: 'category',
    description: 'Span execution status',
    operators: ['equals', 'not_equals'],
    options: [
      { value: 'ok', label: 'OK' },
      { value: 'error', label: 'Error' },
    ],
  },
  {
    id: 'attributes',
    label: 'Custom Attribute',
    type: 'string',
    description:
      'Match against span attributes (use dot notation: key.subkey)',
    operators: [
      'equals',
      'not_equals',
      'contains',
      'is_empty',
      'is_not_empty',
    ],
  },
]

const OPERATOR_LABELS: Record<FilterOperator, string> = {
  equals: '=',
  not_equals: '≠',
  contains: 'contains',
  gt: '>',
  lt: '<',
  gte: '≥',
  lte: '≤',
  is_empty: 'is empty',
  is_not_empty: 'is not empty',
}

function operatorRequiresValue(operator: FilterOperator): boolean {
  return operator !== 'is_empty' && operator !== 'is_not_empty'
}

interface LocalFilter {
  id: string
  field: string
  operator: FilterOperator
  value: unknown
}

interface EvaluatorFilterBuilderProps {
  value: FilterClause[]
  onChange: (filters: FilterClause[]) => void
  disabled?: boolean
  maxFilters?: number
  columns?: EvaluatorFilterColumn[]
}

export function EvaluatorFilterBuilder({
  value,
  onChange,
  disabled = false,
  maxFilters = 10,
  columns = EVALUATOR_FILTER_COLUMNS,
}: EvaluatorFilterBuilderProps) {
  const [localFilters, setLocalFilters] = useState<LocalFilter[]>(() =>
    value.map((f) => ({
      id: nanoid(8),
      field: f.field,
      operator: f.operator,
      value: f.value,
    })),
  )

  const canAddMore = localFilters.length < maxFilters

  const emitChange = useCallback(
    (filters: LocalFilter[]) => {
      const valid: FilterClause[] = filters
        .filter((f) => {
          if (!f.field) return false
          if (!operatorRequiresValue(f.operator)) return true
          return f.value !== null && f.value !== undefined && f.value !== ''
        })
        .map((f) => ({
          field: f.field,
          operator: f.operator,
          value: f.value,
        }))
      onChange(valid)
    },
    [onChange],
  )

  const addFilter = useCallback(() => {
    setLocalFilters((prev) => [
      ...prev,
      {
        id: nanoid(8),
        field: '',
        operator: 'equals',
        value: '',
      },
    ])
  }, [])

  const updateFilter = useCallback(
    (id: string, updates: Partial<LocalFilter>) => {
      setLocalFilters((prev) => {
        const next = prev.map((f) => (f.id === id ? { ...f, ...updates } : f))
        emitChange(next)
        return next
      })
    },
    [emitChange],
  )

  const removeFilter = useCallback(
    (id: string) => {
      setLocalFilters((prev) => {
        const next = prev.filter((f) => f.id !== id)
        emitChange(next)
        return next
      })
    },
    [emitChange],
  )

  const handleColumnChange = useCallback(
    (id: string, columnId: string) => {
      const column = columns.find((c) => c.id === columnId)
      const defaultOperator: FilterOperator =
        column?.operators[0] ?? 'equals'
      updateFilter(id, {
        field: columnId,
        operator: defaultOperator,
        value: '',
      })
    },
    [columns, updateFilter],
  )

  const handleOperatorChange = useCallback(
    (id: string, operator: FilterOperator) => {
      const updates: Partial<LocalFilter> = { operator }
      if (!operatorRequiresValue(operator)) {
        updates.value = null
      }
      updateFilter(id, updates)
    },
    [updateFilter],
  )

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <span className="text-sm font-medium">Targeting Filters</span>
        {localFilters.length > 0 ? (
          <Badge variant="secondary" className="text-xs">
            {localFilters.length}{' '}
            {localFilters.length === 1 ? 'filter' : 'filters'}
          </Badge>
        ) : null}
      </div>

      {localFilters.length === 0 ? (
        <div className="rounded-lg border border-dashed p-4 text-center">
          <p className="mb-2 text-sm text-muted-foreground">
            No filters configured. All matching spans will be evaluated.
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addFilter}
            disabled={disabled}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Filter
          </Button>
        </div>
      ) : (
        <div className="space-y-2">
          {localFilters.map((filter, index) => (
            <FilterRow
              key={filter.id}
              filter={filter}
              columns={columns}
              isFirst={index === 0}
              disabled={disabled}
              onColumnChange={(columnId) =>
                handleColumnChange(filter.id, columnId)
              }
              onOperatorChange={(op) => handleOperatorChange(filter.id, op)}
              onValueChange={(v) => updateFilter(filter.id, { value: v })}
              onRemove={() => removeFilter(filter.id)}
            />
          ))}

          {canAddMore ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-full border border-dashed"
              onClick={addFilter}
              disabled={disabled}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Filter
            </Button>
          ) : null}
        </div>
      )}

      <p className="text-xs text-muted-foreground">
        Filters use AND logic. All conditions must match for a span to be
        evaluated.
      </p>
    </div>
  )
}

interface FilterRowProps {
  filter: LocalFilter
  columns: EvaluatorFilterColumn[]
  isFirst: boolean
  disabled: boolean
  onColumnChange: (columnId: string) => void
  onOperatorChange: (operator: FilterOperator) => void
  onValueChange: (value: unknown) => void
  onRemove: () => void
}

function FilterRow({
  filter,
  columns,
  isFirst,
  disabled,
  onColumnChange,
  onOperatorChange,
  onValueChange,
  onRemove,
}: FilterRowProps) {
  const column = useMemo(
    () => columns.find((c) => c.id === filter.field),
    [columns, filter.field],
  )

  const operators: FilterOperator[] = column?.operators ?? ['equals']
  const showValueInput = operatorRequiresValue(filter.operator)

  return (
    <div className="group flex items-center gap-2">
      <div className="w-12 shrink-0 text-center text-xs text-muted-foreground">
        {isFirst ? 'Where' : 'And'}
      </div>

      <Select
        value={filter.field}
        onValueChange={onColumnChange}
        disabled={disabled}
      >
        <SelectTrigger className="h-8 w-[160px]">
          <SelectValue placeholder="Select field" />
        </SelectTrigger>
        <SelectContent>
          {columns.map((col) => (
            <SelectItem key={col.id} value={col.id}>
              <div className="flex items-center gap-2">
                <span>{col.label}</span>
                {col.description ? (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent side="right">
                      <p className="text-xs">{col.description}</p>
                    </TooltipContent>
                  </Tooltip>
                ) : null}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={filter.operator}
        onValueChange={(v) => onOperatorChange(v as FilterOperator)}
        disabled={disabled || !filter.field}
      >
        <SelectTrigger className="h-8 w-[120px]">
          <SelectValue placeholder="Op" />
        </SelectTrigger>
        <SelectContent>
          {operators.map((op) => (
            <SelectItem key={op} value={op}>
              {OPERATOR_LABELS[op]}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {showValueInput ? (
        <div className="min-w-[120px] flex-1">
          <FilterValueInput
            column={column}
            value={filter.value}
            onChange={onValueChange}
            disabled={disabled || !filter.field}
          />
        </div>
      ) : (
        <div className="min-w-[120px] flex-1" />
      )}

      <Button
        type="button"
        variant="ghost"
        size="icon"
        className={cn(
          'h-8 w-8 shrink-0 opacity-0 transition-opacity group-hover:opacity-100',
        )}
        onClick={onRemove}
        disabled={disabled}
      >
        <X className="h-4 w-4" />
        <span className="sr-only">Remove filter</span>
      </Button>
    </div>
  )
}

interface FilterValueInputProps {
  column?: EvaluatorFilterColumn
  value: unknown
  onChange: (value: unknown) => void
  disabled: boolean
}

function FilterValueInput({
  column,
  value,
  onChange,
  disabled,
}: FilterValueInputProps) {
  if (column?.type === 'category' && column.options) {
    return (
      <Select
        value={String(value ?? '')}
        onValueChange={onChange}
        disabled={disabled}
      >
        <SelectTrigger className="h-8">
          <SelectValue placeholder="Select value" />
        </SelectTrigger>
        <SelectContent>
          {column.options.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )
  }

  if (column?.type === 'number') {
    const numValue =
      typeof value === 'number' ? value : value === '' ? '' : ''
    return (
      <Input
        type="number"
        value={numValue}
        onChange={(e) => {
          const val = e.target.value
          onChange(val === '' ? '' : Number(val))
        }}
        placeholder="Enter value"
        className="h-8"
        disabled={disabled}
      />
    )
  }

  return (
    <Input
      type="text"
      value={String(value ?? '')}
      onChange={(e) => onChange(e.target.value)}
      placeholder="Enter value"
      className="h-8"
      disabled={disabled}
    />
  )
}
