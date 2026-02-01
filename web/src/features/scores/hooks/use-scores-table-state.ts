'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString, parseAsStringLiteral } from 'nuqs'
import type { ScoreType, ScoreSource } from '../types'

export type ScoreSortField = 'name' | 'value' | 'type' | 'source' | 'timestamp'

const DATA_TYPE_VALUES = ['NUMERIC', 'BOOLEAN', 'CATEGORICAL'] as const
const SOURCE_VALUES = ['code', 'llm', 'human'] as const

export interface UseScoresTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  dataType: ScoreType | null
  source: ScoreSource | null
  sortBy: ScoreSortField
  sortOrder: 'asc' | 'desc'

  // Setters (update URL)
  setSearch: (search: string) => void
  setDataType: (dataType: ScoreType | null) => void
  setSource: (source: ScoreSource | null) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: ScoreSortField | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean

  // API params format
  toApiParams: () => {
    name?: string
    type?: ScoreType
    source?: ScoreSource
    page: number
    limit: number
    sort_by: ScoreSortField
    sort_dir: 'asc' | 'desc'
  }
}

const validSortFields: ScoreSortField[] = ['name', 'value', 'type', 'source', 'timestamp']

/**
 * Centralized hook that manages ALL scores table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query (filter by name)
 * - dataType: Filter by score data type (NUMERIC, BOOLEAN, CATEGORICAL)
 * - source: Filter by score source (code, llm, human)
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function useScoresTableState(): UseScoresTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(50),
    search: parseAsString,
    dataType: parseAsStringLiteral(DATA_TYPE_VALUES),
    source: parseAsStringLiteral(SOURCE_VALUES),
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Validate sortBy
  const sortBy = useMemo((): ScoreSortField => {
    if (query.sortBy && validSortFields.includes(query.sortBy as ScoreSortField)) {
      return query.sortBy as ScoreSortField
    }
    return 'timestamp'
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

  const setDataType = useCallback(
    (dataType: ScoreType | null) => {
      setQuery({
        dataType: dataType || null,
        page: 1, // Reset to page 1 when filter changes
      })
    },
    [setQuery]
  )

  const setSource = useCallback(
    (source: ScoreSource | null) => {
      setQuery({
        source: source || null,
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
    (newSortBy: ScoreSortField | null, newSortOrder: 'asc' | 'desc' | null) => {
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
      dataType: null,
      source: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  // Convert to API params format
  const toApiParams = useCallback(() => ({
    name: query.search || undefined,
    type: query.dataType || undefined,
    source: query.source || undefined,
    page: Math.max(1, query.page),
    limit: Math.max(1, query.pageSize),
    sort_by: sortBy,
    sort_dir: sortOrder,
  }), [query.search, query.dataType, query.source, query.page, query.pageSize, sortBy, sortOrder])

  return {
    // State (read from URL) - validated to ensure minimum of 1
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    dataType: query.dataType,
    source: query.source,
    sortBy,
    sortOrder,

    // Setters (update URL)
    setSearch,
    setDataType,
    setSource,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: !!(query.search || query.dataType || query.source),

    // API params format
    toApiParams,
  }
}
