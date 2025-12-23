'use client'

import { useCallback, useEffect, useMemo, useRef } from 'react'
import { usePathname, useRouter, type ReadonlyURLSearchParams } from 'next/navigation'
import type { ColumnFiltersState, PaginationState, SortingState } from '@tanstack/react-table'
import { buildTableUrl, debounce } from '@/lib/utils/table-utils'

type UsePromptsTableNavigationProps = {
  searchParams: ReadonlyURLSearchParams
  onSearchChange?: () => void
}

/**
 * Custom hook for prompts table navigation in Next.js App Router
 * Handles URL-based state for pagination, filtering, and sorting
 */
export function usePromptsTableNavigation({ searchParams, onSearchChange }: UsePromptsTableNavigationProps) {
  const router = useRouter()
  const pathname = usePathname()

  // Memoize searchParams conversion
  const searchParamsObj = useMemo(
    () => Object.fromEntries(searchParams.entries()),
    [searchParams]
  )

  /**
   * Handle pagination changes
   */
  const handlePageChange = useCallback(
    (pagination: PaginationState) => {
      const updates: Record<string, string | null> = {
        page: String(pagination.pageIndex + 1),
      }

      if (pagination.pageSize !== 10) {
        updates.pageSize = String(pagination.pageSize)
      } else {
        updates.pageSize = null
      }

      const url = pathname + buildTableUrl(searchParamsObj, updates)
      router.push(url)
    },
    [router, pathname, searchParamsObj]
  )

  /**
   * Handle search/global filter changes with debounce
   */
  const latestRef = useRef({ searchParamsObj, pathname, router, onSearchChange })

  useEffect(() => {
    latestRef.current = { searchParamsObj, pathname, router, onSearchChange }
  }, [searchParamsObj, pathname, router, onSearchChange])

  const handleSearch = useMemo(
    () =>
      debounce((filter: string) => {
        const { pathname, searchParamsObj, router, onSearchChange } = latestRef.current
        const updates: Record<string, string | null> = {
          filter: filter || null,
          page: '1',
        }

        const url = pathname + buildTableUrl(searchParamsObj, updates)
        router.push(url)
        onSearchChange?.()
      }, 500),
    []
  )

  // Cleanup debounce on unmount
  useEffect(() => {
    return () => {
      if (handleSearch.cancel) {
        handleSearch.cancel()
      }
    }
  }, [handleSearch])

  /**
   * Handle column filter changes (type filter for prompts)
   */
  const handleFilter = useCallback(
    (filters: ColumnFiltersState) => {
      const updates: Record<string, string | null> = {
        page: '1',
      }

      // Process type filter
      filters.forEach((filter) => {
        if (filter.id === 'type') {
          if (Array.isArray(filter.value) && filter.value.length > 0) {
            updates.type = JSON.stringify(filter.value)
          } else {
            updates.type = null
          }
        }
      })

      // Clear type filter if not in new filters
      if (!filters.some((f) => f.id === 'type')) {
        updates.type = null
      }

      const url = pathname + buildTableUrl(searchParamsObj, updates)
      router.push(url)
      onSearchChange?.()
    },
    [router, pathname, searchParamsObj, onSearchChange]
  )

  /**
   * Handle sorting changes
   */
  const handleSort = useCallback(
    (sorting: SortingState) => {
      const updates: Record<string, string | null> = {}

      if (sorting.length === 0) {
        updates.sortBy = null
        updates.sortOrder = null
      } else {
        const [sort] = sorting
        updates.sortBy = sort.id
        updates.sortOrder = sort.desc ? 'desc' : 'asc'
      }

      const url = pathname + buildTableUrl(searchParamsObj, updates)
      router.push(url)
    },
    [router, pathname, searchParamsObj]
  )

  /**
   * Reset all filters
   */
  const handleReset = useCallback(() => {
    router.push(pathname)
  }, [router, pathname])

  return {
    handlePageChange,
    handleSearch,
    handleFilter,
    handleSort,
    handleReset,
  }
}
