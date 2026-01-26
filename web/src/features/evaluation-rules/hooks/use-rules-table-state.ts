'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString, parseAsStringLiteral } from 'nuqs'
import type { ScorerType, RuleStatus } from '../types'

export type RuleSortField = 'name' | 'status' | 'sampling_rate' | 'created_at' | 'updated_at'

const SCORER_TYPE_VALUES = ['llm', 'builtin', 'regex'] as const
const STATUS_VALUES = ['active', 'inactive', 'paused'] as const

export interface UseRulesTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  scorerType: ScorerType | null
  status: RuleStatus | null
  sortBy: RuleSortField
  sortOrder: 'asc' | 'desc'

  // Setters (update URL)
  setSearch: (search: string) => void
  setScorerType: (scorerType: ScorerType | null) => void
  setStatus: (status: RuleStatus | null) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: RuleSortField | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean

  // API params format
  toApiParams: () => {
    search?: string
    scorer_type?: ScorerType
    status?: RuleStatus
    page: number
    limit: number
    sort_by: RuleSortField
    sort_dir: 'asc' | 'desc'
  }
}

const validSortFields: RuleSortField[] = ['name', 'status', 'sampling_rate', 'created_at', 'updated_at']

/**
 * Centralized hook that manages ALL rules table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query (filter by name)
 * - scorerType: Filter by scorer type (llm, builtin, regex)
 * - status: Filter by rule status (active, inactive, paused)
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function useRulesTableState(): UseRulesTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(25),
    search: parseAsString,
    scorerType: parseAsStringLiteral(SCORER_TYPE_VALUES),
    status: parseAsStringLiteral(STATUS_VALUES),
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Validate sortBy
  const sortBy = useMemo((): RuleSortField => {
    if (query.sortBy && validSortFields.includes(query.sortBy as RuleSortField)) {
      return query.sortBy as RuleSortField
    }
    return 'created_at'
  }, [query.sortBy])

  // Validate sortOrder
  const sortOrder = useMemo((): 'asc' | 'desc' => {
    if (query.sortOrder === 'asc' || query.sortOrder === 'desc') {
      return query.sortOrder
    }
    return 'desc'
  }, [query.sortOrder])

  // Setters that update URL
  const setSearch = useCallback(
    (search: string) => {
      setQuery({
        search: search || null,
        page: 1, // Reset to page 1 when search changes
      })
    },
    [setQuery]
  )

  const setScorerType = useCallback(
    (scorerType: ScorerType | null) => {
      setQuery({
        scorerType: scorerType || null,
        page: 1, // Reset to page 1 when filter changes
      })
    },
    [setQuery]
  )

  const setStatus = useCallback(
    (status: RuleStatus | null) => {
      setQuery({
        status: status || null,
        page: 1, // Reset to page 1 when filter changes
      })
    },
    [setQuery]
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        page: Math.max(1, page),
        ...(pageSize !== undefined && { pageSize: Math.max(1, pageSize) }),
      })
    },
    [setQuery]
  )

  const setSorting = useCallback(
    (newSortBy: RuleSortField | null, newSortOrder: 'asc' | 'desc' | null) => {
      setQuery({
        sortBy: newSortBy || null,
        sortOrder: newSortOrder || null,
      })
    },
    [setQuery]
  )

  const resetAll = useCallback(() => {
    setQuery({
      page: 1,
      pageSize: null,
      search: null,
      scorerType: null,
      status: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  // Convert to API params format
  const toApiParams = useCallback(() => ({
    search: query.search || undefined,
    scorer_type: query.scorerType || undefined,
    status: query.status || undefined,
    page: Math.max(1, query.page),
    limit: Math.max(1, query.pageSize),
    sort_by: sortBy,
    sort_dir: sortOrder,
  }), [query.search, query.scorerType, query.status, query.page, query.pageSize, sortBy, sortOrder])

  return {
    // State (read from URL) - validated to ensure minimum of 1
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    scorerType: query.scorerType,
    status: query.status,
    sortBy,
    sortOrder,

    // Setters (update URL)
    setSearch,
    setScorerType,
    setStatus,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: !!(query.search || query.scorerType || query.status),

    // API params format
    toApiParams,
  }
}
