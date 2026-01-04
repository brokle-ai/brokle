/**
 * Shared Filter Builder Utilities
 *
 * Helper functions for filter operations
 */

import type { FilterOperator, ColumnType } from './types'

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
 * Check if operator requires a value input
 */
export function operatorRequiresValue(operator: FilterOperator): boolean {
  return !['IS EMPTY', 'IS NOT EMPTY', 'EXISTS', 'NOT EXISTS'].includes(operator)
}

/**
 * Check if operator accepts multiple values
 */
export function operatorAcceptsMultiple(operator: FilterOperator): boolean {
  return ['IN', 'NOT IN'].includes(operator)
}
