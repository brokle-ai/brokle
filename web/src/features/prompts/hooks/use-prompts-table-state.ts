'use client'

import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import type { PromptType } from '../types'

const promptTypes = ['text', 'chat'] as const

export interface UsePromptsTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  search: string | null
  types: PromptType[]
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null

  // Setters (update URL)
  setSearch: (search: string) => void
  setTypes: (types: PromptType[]) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean
}

/**
 * Centralized hook that manages ALL prompts table state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - page: Page number (1-indexed)
 * - pageSize: Rows per page
 * - search: Text search query (filter by name)
 * - types: JSON-encoded array of prompt types (text, chat)
 * - sortBy: Column to sort by
 * - sortOrder: Sort direction (asc, desc)
 */
export function usePromptsTableState(): UsePromptsTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(10),
    search: parseAsString,
    types: parseAsString, // JSON-encoded array of PromptType
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  // Decode types from URL
  const types = useMemo((): PromptType[] => {
    if (!query.types) return []
    try {
      const parsed = JSON.parse(query.types)
      return Array.isArray(parsed)
        ? parsed.filter((t): t is PromptType => promptTypes.includes(t as PromptType))
        : []
    } catch {
      return []
    }
  }, [query.types])

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

  const setTypes = useCallback(
    (newTypes: PromptType[]) => {
      setQuery({
        types: newTypes.length > 0 ? JSON.stringify(newTypes) : null,
        page: 1, // Reset to page 1 when types change
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
      types: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  return {
    // State (read from URL) - validated to ensure minimum of 1
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    types,
    sortBy: query.sortBy,
    sortOrder: query.sortOrder as 'asc' | 'desc' | null,

    // Setters (update URL)
    setSearch,
    setTypes,
    setPagination,
    setSorting,
    resetAll,

    // Computed
    hasActiveFilters: !!query.search || types.length > 0,
  }
}
