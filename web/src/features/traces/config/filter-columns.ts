/**
 * Filter column definitions for traces
 *
 * Uses shared filter types and utilities from @/components/shared/filter-builder
 * Defines trace-specific filterable columns with their types, operators, and UI configuration
 */

// Re-export shared types and utilities for backwards compatibility
export type {
  ColumnType,
  ColumnDefinition,
  FilterOperator,
} from '@/components/shared/filter-builder'

export {
  stringOperators,
  numberOperators,
  categoryOperators,
  booleanOperators,
  existsOperators,
  searchOperators,
  operatorLabels,
  getOperatorsForType,
  operatorRequiresValue,
  operatorAcceptsMultiple,
} from '@/components/shared/filter-builder'

import type { ColumnDefinition, FilterOperator } from '@/components/shared/filter-builder'
import { searchOperators, stringOperators, numberOperators, categoryOperators, existsOperators } from '@/components/shared/filter-builder'

// String operators WITHOUT IN/NOT IN for columns where backend doesn't support multi-value filtering
// model_name, provider_name, service_name only support single-value filters on the backend
const stringOperatorsWithoutMultiValue: FilterOperator[] = stringOperators.filter(
  op => op !== 'IN' && op !== 'NOT IN'
)

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
  // Note: These columns use stringOperatorsWithoutMultiValue because the backend
  // only supports single-value filtering for model_name, provider_name, service_name
  {
    id: 'model_name',
    label: 'Model',
    type: 'string',
    filterable: true,
    operators: stringOperatorsWithoutMultiValue,
    description: 'AI model name (e.g., gpt-4, claude-3)',
  },
  {
    id: 'provider_name',
    label: 'Provider',
    type: 'string',
    filterable: true,
    operators: stringOperatorsWithoutMultiValue,
    description: 'AI provider (e.g., openai, anthropic)',
  },
  {
    id: 'service_name',
    label: 'Service',
    type: 'string',
    filterable: true,
    operators: stringOperatorsWithoutMultiValue,
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
 * Get column definition by ID
 */
export function getColumnById(
  id: string,
  table: 'traces' | 'spans' = 'traces'
): ColumnDefinition | undefined {
  const columns = table === 'traces' ? traceFilterColumns : spanFilterColumns
  return columns.find((col) => col.id === id)
}
