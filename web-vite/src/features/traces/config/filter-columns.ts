/**
 * Trace-table filter column definitions. Backed by the shared
 * `@/components/shared/filter-builder` types so the column config is
 * the only place trace-specific filterable fields are declared. The
 * shared FilterBuilder component drives the UI; this file just maps
 * column names → operator + type + option metadata.
 */

import type {
  ColumnDefinition,
  FilterOperator,
} from '@/components/shared/filter-builder'
import {
  categoryOperators,
  numberOperators,
  searchOperators,
  stringOperators,
} from '@/components/shared/filter-builder'

// model_name / provider_name / service_name only support single-value
// equality on the backend — drop IN / NOT IN to avoid a footgun where
// the UI lets a user build a filter the backend silently rejects.
const stringOperatorsWithoutMultiValue: FilterOperator[] =
  stringOperators.filter((op) => op !== 'IN' && op !== 'NOT IN')

const statusOptions = [
  { value: '0', label: 'Unset' },
  { value: '1', label: 'OK' },
  { value: '2', label: 'Error' },
]

export const traceFilterColumns: ColumnDefinition[] = [
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
  {
    id: 'start_time',
    label: 'Start Time',
    type: 'datetime',
    filterable: true,
    operators: ['>', '<', '>=', '<='],
    description: 'Span start timestamp',
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
]
