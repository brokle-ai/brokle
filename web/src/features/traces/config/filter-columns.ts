/**
 * Filter column definitions for the advanced filter builder
 *
 * Defines filterable columns with their types, operators, and UI configuration
 */

import type { FilterOperator } from '../api/traces-api'

export type ColumnType =
  | 'string'
  | 'number'
  | 'duration'
  | 'cost'
  | 'datetime'
  | 'category'
  | 'boolean'
  | 'json'

export interface ColumnDefinition {
  id: string
  label: string
  type: ColumnType
  filterable: boolean
  operators: FilterOperator[]
  description?: string
  // For category type: predefined options
  options?: { value: string; label: string }[]
  // For number/duration/cost: unit information
  unit?: string
  // For dynamic columns from attributes
  dynamic?: boolean
}

// Operators by type
export const stringOperators: FilterOperator[] = [
  '=',
  '!=',
  'CONTAINS',
  'NOT CONTAINS',
  'STARTS WITH',
  'ENDS WITH',
  'REGEX',
  'IS EMPTY',
  'IS NOT EMPTY',
  'IN',
  'NOT IN',
]

export const numberOperators: FilterOperator[] = [
  '=',
  '!=',
  '>',
  '<',
  '>=',
  '<=',
  'IS EMPTY',
  'IS NOT EMPTY',
]

export const categoryOperators: FilterOperator[] = ['=', '!=', 'IN', 'NOT IN']

export const booleanOperators: FilterOperator[] = ['=', '!=']

export const existsOperators: FilterOperator[] = ['EXISTS', 'NOT EXISTS']

export const searchOperators: FilterOperator[] = ['~', 'CONTAINS', 'REGEX']

// Operator labels for display
export const operatorLabels: Record<FilterOperator, string> = {
  '=': 'equals',
  '!=': 'not equals',
  '>': 'greater than',
  '<': 'less than',
  '>=': 'greater or equal',
  '<=': 'less or equal',
  CONTAINS: 'contains',
  'NOT CONTAINS': 'not contains',
  IN: 'in',
  'NOT IN': 'not in',
  EXISTS: 'exists',
  'NOT EXISTS': 'not exists',
  'STARTS WITH': 'starts with',
  'ENDS WITH': 'ends with',
  REGEX: 'matches regex',
  'IS EMPTY': 'is empty',
  'IS NOT EMPTY': 'is not empty',
  '~': 'search',
}

// Status options for category filter
const statusOptions = [
  { value: '0', label: 'Unset' },
  { value: '1', label: 'OK' },
  { value: '2', label: 'Error' },
]

/**
 * Trace table column definitions
 */
export const traceFilterColumns: ColumnDefinition[] = [
  // Identifiers
  {
    id: 'trace_id',
    label: 'Trace ID',
    type: 'string',
    filterable: true,
    operators: ['=', '!=', 'STARTS WITH', 'IN'],
    description: 'Unique trace identifier (32 hex characters)',
  },
  {
    id: 'span_id',
    label: 'Span ID',
    type: 'string',
    filterable: true,
    operators: ['=', '!=', 'STARTS WITH', 'IN'],
    description: 'Unique span identifier (16 hex characters)',
  },
  {
    id: 'span_name',
    label: 'Name',
    type: 'string',
    filterable: true,
    operators: stringOperators,
    description: 'Trace or span name',
  },

  // Status
  {
    id: 'status_code',
    label: 'Status',
    type: 'category',
    filterable: true,
    operators: categoryOperators,
    options: statusOptions,
    description: 'Span status code (OK, Error, Unset)',
  },
  {
    id: 'status_message',
    label: 'Status Message',
    type: 'string',
    filterable: true,
    operators: stringOperators,
    description: 'Status message (typically for errors)',
  },

  // Provider/Model
  {
    id: 'model_name',
    label: 'Model',
    type: 'string',
    filterable: true,
    operators: stringOperators,
    description: 'AI model name (e.g., gpt-4, claude-3)',
  },
  {
    id: 'provider_name',
    label: 'Provider',
    type: 'string',
    filterable: true,
    operators: stringOperators,
    description: 'AI provider (e.g., openai, anthropic)',
  },
  {
    id: 'service_name',
    label: 'Service',
    type: 'string',
    filterable: true,
    operators: stringOperators,
    description: 'Service name from resource attributes',
  },

  // Timing
  {
    id: 'start_time',
    label: 'Start Time',
    type: 'datetime',
    filterable: true,
    operators: ['>', '<', '>=', '<='],
    description: 'Span start timestamp',
  },
  {
    id: 'end_time',
    label: 'End Time',
    type: 'datetime',
    filterable: true,
    operators: ['>', '<', '>=', '<='],
    description: 'Span end timestamp',
  },
  {
    id: 'duration_nano',
    label: 'Duration',
    type: 'duration',
    filterable: true,
    operators: numberOperators,
    unit: 'ns',
    description: 'Span duration in nanoseconds',
  },

  // Tokens
  {
    id: 'input_tokens',
    label: 'Input Tokens',
    type: 'number',
    filterable: true,
    operators: numberOperators,
    description: 'Number of input/prompt tokens',
  },
  {
    id: 'output_tokens',
    label: 'Output Tokens',
    type: 'number',
    filterable: true,
    operators: numberOperators,
    description: 'Number of output/completion tokens',
  },
  {
    id: 'total_tokens',
    label: 'Total Tokens',
    type: 'number',
    filterable: true,
    operators: numberOperators,
    description: 'Total tokens (input + output)',
  },

  // Cost
  {
    id: 'total_cost',
    label: 'Cost',
    type: 'cost',
    filterable: true,
    operators: numberOperators,
    unit: 'USD',
    description: 'Total cost in USD',
  },
  {
    id: 'input_cost',
    label: 'Input Cost',
    type: 'cost',
    filterable: true,
    operators: numberOperators,
    unit: 'USD',
    description: 'Input token cost',
  },
  {
    id: 'output_cost',
    label: 'Output Cost',
    type: 'cost',
    filterable: true,
    operators: numberOperators,
    unit: 'USD',
    description: 'Output token cost',
  },

  // Content search
  {
    id: 'input',
    label: 'Input',
    type: 'string',
    filterable: true,
    operators: searchOperators,
    description: 'Input/prompt content',
  },
  {
    id: 'output',
    label: 'Output',
    type: 'string',
    filterable: true,
    operators: searchOperators,
    description: 'Output/completion content',
  },

  // Context identifiers
  {
    id: 'session_id',
    label: 'Session ID',
    type: 'string',
    filterable: true,
    operators: ['=', '!=', 'IS EMPTY', 'IS NOT EMPTY'],
    description: 'Session identifier for grouping related traces',
  },
  {
    id: 'user_id',
    label: 'User ID',
    type: 'string',
    filterable: true,
    operators: ['=', '!=', 'IS EMPTY', 'IS NOT EMPTY'],
    description: 'End user identifier',
  },

  // Attributes (dynamic)
  {
    id: 'span_attributes',
    label: 'Span Attributes',
    type: 'json',
    filterable: true,
    operators: existsOperators,
    dynamic: true,
    description: 'Custom span attributes (key-value pairs)',
  },
  {
    id: 'resource_attributes',
    label: 'Resource Attributes',
    type: 'json',
    filterable: true,
    operators: existsOperators,
    dynamic: true,
    description: 'Resource attributes (key-value pairs)',
  },
]

/**
 * Span table column definitions (subset of trace columns)
 */
export const spanFilterColumns: ColumnDefinition[] = traceFilterColumns.filter(
  (col) => !['trace_id'].includes(col.id) // Spans always have a trace_id context
)

/**
 * Get operators for a column type
 */
export function getOperatorsForType(type: ColumnType): FilterOperator[] {
  switch (type) {
    case 'string':
      return stringOperators
    case 'number':
    case 'duration':
    case 'cost':
      return numberOperators
    case 'datetime':
      return ['>', '<', '>=', '<=']
    case 'category':
      return categoryOperators
    case 'boolean':
      return booleanOperators
    case 'json':
      return existsOperators
    default:
      return stringOperators
  }
}

/**
 * Get column definition by ID
 */
export function getColumnById(
  id: string,
  table: 'traces' | 'spans' = 'traces'
): ColumnDefinition | undefined {
  const columns = table === 'traces' ? traceFilterColumns : spanFilterColumns
  return columns.find((col) => col.id === id)
}

/**
 * Check if operator requires a value input
 */
export function operatorRequiresValue(operator: FilterOperator): boolean {
  return !['IS EMPTY', 'IS NOT EMPTY', 'EXISTS', 'NOT EXISTS'].includes(
    operator
  )
}

/**
 * Check if operator accepts multiple values
 */
export function operatorAcceptsMultiple(operator: FilterOperator): boolean {
  return ['IN', 'NOT IN'].includes(operator)
}
