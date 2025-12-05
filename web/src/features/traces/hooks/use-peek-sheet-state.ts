'use client'

import { useCallback } from 'react'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import type { ViewMode } from '../components/peek-sheet/span-navigation-panel'

/**
 * Hook for managing peek sheet state via URL parameters
 *
 * URL parameters:
 * - peek: trace ID to display in peek sheet
 * - span: selected span ID within the trace
 * - view: span visualization mode ('tree' | 'timeline')
 */
export function usePeekSheetState() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // Current state from URL
  const selectedTraceId = searchParams.get('peek')
  const selectedSpanId = searchParams.get('span')
  const viewMode = (searchParams.get('view') as ViewMode) || 'tree'

  /**
   * Set selected span ID (or clear it)
   */
  const setSelectedSpan = useCallback(
    (spanId: string | null) => {
      const params = new URLSearchParams(searchParams.toString())
      if (spanId) {
        params.set('span', spanId)
      } else {
        params.delete('span')
      }
      router.replace(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  /**
   * Set view mode (tree or timeline)
   */
  const setViewMode = useCallback(
    (mode: ViewMode) => {
      const params = new URLSearchParams(searchParams.toString())
      if (mode !== 'tree') {
        params.set('view', mode)
      } else {
        params.delete('view') // 'tree' is default, no need to store
      }
      router.replace(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  /**
   * Clear all peek sheet state
   */
  const clearPeekState = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete('peek')
    params.delete('span')
    params.delete('view')
    router.push(`${pathname}?${params.toString()}`)
  }, [router, pathname, searchParams])

  return {
    // Current state
    selectedTraceId,
    selectedSpanId,
    viewMode,

    // State setters
    setSelectedSpan,
    setViewMode,
    clearPeekState,
  }
}
