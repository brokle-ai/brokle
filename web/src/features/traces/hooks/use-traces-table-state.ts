'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import type { FilterCondition } from '../api/traces-api'

export interface UseTracesTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  searchType: string | null
  filters: FilterCondition[]
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null

  // Setters (update URL)
  setFilters: (filters: FilterCondition[]) => void
  setSearch: (search: string, searchType?: string) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean
}

/**
 * Centralized hook that manages ALL table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query
 * - searchType: Type of search (id, content, all)
 * - filters: JSON-encoded array of FilterCondition
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function useTracesTableState(): UseTracesTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(20),
    search: parseAsString,
    searchType: parseAsString,
    filters: parseAsString, // JSON-encoded array of FilterCondition
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Decode filters from URL
  const filters = useMemo((): FilterCondition[] => {
    if (!query.filters) return []
    try {
      return JSON.parse(query.filters)
    } catch {
      return []
    }
  }, [query.filters])

  // Setters that update URL
  const setFilters = useCallback(
    (newFilters: FilterCondition[]) => {
      setQuery({
        filters: newFilters.length > 0 ? JSON.stringify(newFilters) : null,
        page: 1, // Reset to page 1 when filters change
      })
    },
    [setQuery]
  )

  const setSearch = useCallback(
    (search: string, searchType?: string) => {
      setQuery({
        search: search || null,
        searchType: searchType || null,
        page: 1, // Reset to page 1 when search changes
      })
    },
    [setQuery]
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        page,
        ...(pageSize !== undefined && { pageSize }),
      })
    },
    [setQuery]
  )

  const setSorting = useCallback(
    (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => {
      setQuery({
        sortBy: sortBy || null,
        sortOrder: sortOrder || null,
      })
    },
    [setQuery]
  )

  const resetAll = useCallback(() => {
    setQuery({
      page: 1,
      pageSize: null,
      search: null,
      searchType: null,
      filters: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  return {
    // State (read from URL)
    page: query.page,
    pageSize: query.pageSize,
    search: query.search,
    searchType: query.searchType,
    filters,
    sortBy: query.sortBy,
    sortOrder: query.sortOrder as 'asc' | 'desc' | null,

    // Setters (update URL)
    setFilters,
    setSearch,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: filters.length > 0 || !!query.search,
  }
}
