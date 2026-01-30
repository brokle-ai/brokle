'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString, parseAsStringLiteral } from 'nuqs'
import type { QueueStatus } from '../types'

export type QueueSortField = 'name' | 'status' | 'created_at' | 'updated_at'

const STATUS_VALUES = ['active', 'paused', 'archived'] as const

export interface UseQueuesTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  status: QueueStatus | null
  sortBy: QueueSortField
  sortOrder: 'asc' | 'desc'

  // Setters (update URL)
  setSearch: (search: string) => void
  setStatus: (status: QueueStatus | null) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: QueueSortField | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean

  // API params format
  toApiParams: () => {
    search?: string
    status?: QueueStatus
    page: number
    limit: number
    sort_by: QueueSortField
    sort_dir: 'asc' | 'desc'
  }
}

const validSortFields: QueueSortField[] = ['name', 'status', 'created_at', 'updated_at']

/**
 * Centralized hook that manages ALL queues table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query (filter by name)
 * - status: Filter by queue status (active, paused, archived)
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function useQueuesTableState(): UseQueuesTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(25),
    search: parseAsString,
    status: parseAsStringLiteral(STATUS_VALUES),
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Validate sortBy
  const sortBy = useMemo((): QueueSortField => {
    if (query.sortBy && validSortFields.includes(query.sortBy as QueueSortField)) {
      return query.sortBy as QueueSortField
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

  const setStatus = useCallback(
    (status: QueueStatus | null) => {
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
    (newSortBy: QueueSortField | null, newSortOrder: 'asc' | 'desc' | null) => {
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
      status: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  // Convert to API params format
  const toApiParams = useCallback(() => ({
    search: query.search || undefined,
    status: query.status || undefined,
    page: Math.max(1, query.page),
    limit: Math.max(1, query.pageSize),
    sort_by: sortBy,
    sort_dir: sortOrder,
  }), [query.search, query.status, query.page, query.pageSize, sortBy, sortOrder])

  return {
    // State (read from URL) - validated to ensure minimum of 1
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    status: query.status,
    sortBy,
    sortOrder,

    // Setters (update URL)
    setSearch,
    setStatus,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: !!(query.search || query.status),

    // API params format
    toApiParams,
  }
}
