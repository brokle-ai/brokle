'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import type { DatasetListParams } from '../types'

export type SortField = 'name' | 'created_at' | 'updated_at' | 'item_count'

export interface UseDatasetsTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  sortBy: SortField
  sortOrder: 'asc' | 'desc'

  // Setters (update URL)
  setSearch: (search: string) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: SortField | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean

  // API params format
  toApiParams: () => DatasetListParams
}

const validSortFields: SortField[] = ['name', 'created_at', 'updated_at', 'item_count']

/**
 * Centralized hook that manages ALL datasets table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query (filter by name)
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function useDatasetsTableState(): UseDatasetsTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(50),
    search: parseAsString,
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Validate sortBy
  const sortBy = useMemo((): SortField => {
    if (query.sortBy && validSortFields.includes(query.sortBy as SortField)) {
      return query.sortBy as SortField
    }
    return 'updated_at'
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
    (newSortBy: SortField | null, newSortOrder: 'asc' | 'desc' | null) => {
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
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  // Convert to API params format
  const toApiParams = useCallback((): DatasetListParams => ({
    search: query.search || undefined,
    page: Math.max(1, query.page),
    limit: Math.max(1, query.pageSize),
    sortBy,
    sortDir: sortOrder,
  }), [query.search, query.page, query.pageSize, sortBy, sortOrder])

  return {
    // State (read from URL) - validated to ensure minimum of 1
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    sortBy,
    sortOrder,

    // Setters (update URL)
    setSearch,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: !!query.search,

    // API params format
    toApiParams,
  }
}
