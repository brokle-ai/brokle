/**
 * Shared Filter Builder Types
 *
 * Generic type definitions for the filter builder that can be used
 * across different features (traces, dashboards, etc.)
 */

// Column types for filter inputs
export type ColumnType =
  | 'string'
  | 'number'
  | 'duration'
  | 'cost'
  | 'datetime'
  | 'category'
  | 'boolean'
  | 'json'

// Common filter operators
export type FilterOperator =
  | '='
  | '!='
  | '>'
  | '<'
  | '>='
  | '<='
  | 'CONTAINS'
  | 'NOT CONTAINS'
  | 'IN'
  | 'NOT IN'
  | 'EXISTS'
  | 'NOT EXISTS'
  | 'STARTS WITH'
  | 'ENDS WITH'
  | 'REGEX'
  | 'IS EMPTY'
  | 'IS NOT EMPTY'
  | '~'

// Column definition for filterable columns
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

// Filter condition representing a single filter
export interface FilterCondition {
  id: string
  column: string
  operator: FilterOperator
  value: string | number | string[] | null
}

// Dynamic options that can be passed to the filter builder
export interface FilterOptions {
  [key: string]: string[] | undefined
}

// Props for the FilterBuilder component
export interface FilterBuilderProps {
  columns: ColumnDefinition[]
  filters: FilterCondition[]
  onApply: (filters: FilterCondition[]) => void
  filterOptions?: FilterOptions
  disabled?: boolean
  maxFilters?: number
  title?: string
  emptyMessage?: string
}

// Props for the FilterRow component
export interface FilterRowProps {
  columns: ColumnDefinition[]
  filter: FilterCondition
  onUpdate: (updates: Partial<FilterCondition>) => void
  onRemove: () => void
  filterOptions?: FilterOptions
  disabled?: boolean
  isFirst?: boolean
  // Optional: mapping of column IDs to their dynamic option keys
  columnOptionMapping?: Record<string, string>
}
