'use client'

import { useCallback, useEffect, useMemo, useRef } from 'react'
import { usePathname, useRouter, type ReadonlyURLSearchParams } from 'next/navigation'
import type { ColumnFiltersState, PaginationState, SortingState } from '@tanstack/react-table'
import { buildTableUrl, debounce } from '@/lib/utils/table-utils'

type UseTableNavigationProps = {
  searchParams: ReadonlyURLSearchParams
  onSearchChange?: () => void
}

/**
 * Custom hook for table navigation in Next.js App Router
 * Accepts ReadonlyURLSearchParams directly, memoizes conversion internally
 * All handlers use stable memoized object to prevent recreation loops
 */
export function useTableNavigation({ searchParams, onSearchChange }: UseTableNavigationProps) {
  const router = useRouter()
  const pathname = usePathname()

  // Memoize searchParams conversion - only recreates when URL actually changes
  const searchParamsObj = useMemo(
    () => Object.fromEntries(searchParams.entries()),
    [searchParams]
  )

  /**
   * Handle pagination changes
   * Updates page and pageSize in URL
   */
  const handlePageChange = useCallback(
    (pagination: PaginationState) => {
      const updates: Record<string, string | null> = {
        page: String(pagination.pageIndex + 1), // Convert to 1-indexed
      }

      // Only include pageSize if it's not the default (10)
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
   * Handle search/global filter changes
   * Uses ref pattern for stable debounce with fresh dependencies
   * Resets to page 1 on search
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
          page: '1', // Reset to first page on search
        }

        const url = pathname + buildTableUrl(searchParamsObj, updates)
        router.push(url)
        onSearchChange?.()
      }, 500),
    []
  )

  // Cleanup: cancel pending debounce on unmount
  useEffect(() => {
    return () => {
      if (handleSearch.cancel) {
        handleSearch.cancel()
      }
    }
  }, [handleSearch])

  /**
   * Handle column filter changes (status, priority, etc.)
   * Resets to page 1 on filter change
   */
  const handleFilter = useCallback(
    (filters: ColumnFiltersState) => {
      const updates: Record<string, string | null> = {
        page: '1', // Reset to first page on filter change
      }

      // Process each filter
      filters.forEach((filter) => {
        if (filter.id === 'status' || filter.id === 'priority') {
          if (Array.isArray(filter.value) && filter.value.length > 0) {
            updates[filter.id] = JSON.stringify(filter.value)
          } else {
            updates[filter.id] = null
          }
        }
      })

      // Clear filters that are not in the new filters array
      if (!filters.some((f) => f.id === 'status')) {
        updates.status = null
      }
      if (!filters.some((f) => f.id === 'priority')) {
        updates.priority = null
      }

      const url = pathname + buildTableUrl(searchParamsObj, updates)
      router.push(url)
      onSearchChange?.()
    },
    [router, pathname, searchParamsObj, onSearchChange]
  )

  /**
   * Handle sorting changes
   * Supports: none → asc → desc → none cycle
   */
  const handleSort = useCallback(
    (sorting: SortingState) => {
      const updates: Record<string, string | null> = {}

      if (sorting.length === 0) {
        // No sorting - clear params
        updates.sortBy = null
        updates.sortOrder = null
      } else {
        // Apply sorting
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
   * Reset all filters and return to default state
   * Navigates to pathname without any query params
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
