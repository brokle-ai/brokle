'use client'

import { useCallback, useMemo, useRef } from 'react'
import { useRouter, usePathname, type ReadonlyURLSearchParams } from 'next/navigation'
import { debounce } from '@/lib/utils/table-utils'

interface UseCardListNavigationOptions {
  searchParams: ReadonlyURLSearchParams
  debounceMs?: number
}

interface UseCardListNavigationReturn {
  handlePageChange: (page: number) => void
  handlePageSizeChange: (pageSize: number) => void
  handleSearch: (filter: string) => void
  handleReset: () => void
}

/**
 * Build URL with updated search params
 */
function buildCardListUrl(
  currentParams: Record<string, string>,
  updates: Record<string, string | null>
): string {
  const newParams = new URLSearchParams()

  // Copy existing params
  for (const [key, value] of Object.entries(currentParams)) {
    if (value) {
      newParams.set(key, value)
    }
  }

  // Apply updates
  for (const [key, value] of Object.entries(updates)) {
    if (value === null || value === '') {
      newParams.delete(key)
    } else {
      newParams.set(key, value)
    }
  }

  const queryString = newParams.toString()
  return queryString ? `?${queryString}` : ''
}

/**
 * Hook for managing card list URL-based navigation
 * Simplified version of useTableNavigation for card grids
 */
export function useCardListNavigation({
  searchParams,
  debounceMs = 500,
}: UseCardListNavigationOptions): UseCardListNavigationReturn {
  const router = useRouter()
  const pathname = usePathname()

  // Convert URLSearchParams to plain object for easier manipulation
  const searchParamsObj = useMemo(() => {
    const obj: Record<string, string> = {}
    searchParams.forEach((value, key) => {
      obj[key] = value
    })
    return obj
  }, [searchParams])

  // Page change handler
  const handlePageChange = useCallback(
    (page: number) => {
      const updates: Record<string, string | null> = {
        page: page > 1 ? String(page) : null,
      }
      const url = pathname + buildCardListUrl(searchParamsObj, updates)
      router.push(url)
    },
    [router, pathname, searchParamsObj]
  )

  // Page size change handler
  const handlePageSizeChange = useCallback(
    (pageSize: number) => {
      const updates: Record<string, string | null> = {
        pageSize: pageSize !== 50 ? String(pageSize) : null,
        page: null, // Reset to first page
      }
      const url = pathname + buildCardListUrl(searchParamsObj, updates)
      router.push(url)
    },
    [router, pathname, searchParamsObj]
  )

  // Debounced search handler
  const debouncedSearchRef = useRef(
    debounce((filter: string) => {
      const updates: Record<string, string | null> = {
        filter: filter || null,
        page: null, // Reset to first page on search
      }
      const url = pathname + buildCardListUrl(searchParamsObj, updates)
      router.push(url)
    }, debounceMs)
  )

  const handleSearch = useCallback((filter: string) => {
    debouncedSearchRef.current(filter)
  }, [])

  // Reset handler
  const handleReset = useCallback(() => {
    // Cancel any pending debounced search
    debouncedSearchRef.current.cancel()
    router.push(pathname)
  }, [router, pathname])

  return {
    handlePageChange,
    handlePageSizeChange,
    handleSearch,
    handleReset,
  }
}
