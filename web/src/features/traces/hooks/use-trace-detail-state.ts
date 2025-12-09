'use client'

import { useCallback, useMemo } from 'react'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useTraces } from '../context/traces-context'

export type ViewMode = 'tree' | 'timeline'

/**
 * Unified hook for managing trace detail state via URL parameters
 *
 * URL parameters (peek mode only):
 * - peek: trace ID to display in sheet sidebar
 * - span: selected span ID within the trace
 * - view: span visualization mode ('tree' | 'timeline')
 *
 * For full-page mode, navigate to /traces/[traceId] route instead
 */
export function useTraceDetailState() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const { currentPageTraceIds, projectSlug } = useTraces()

  // ==========================================================================
  // URL State
  // ==========================================================================

  const traceId = searchParams.get('peek')
  const selectedSpanId = searchParams.get('span')
  const viewMode = (searchParams.get('view') as ViewMode) || 'tree'

  // ==========================================================================
  // Navigation State (prev/next through traces)
  // ==========================================================================

  const currentIndex = useMemo(() => {
    if (!traceId || !currentPageTraceIds.length) return -1
    return currentPageTraceIds.indexOf(traceId)
  }, [traceId, currentPageTraceIds])

  const canGoPrev = currentIndex > 0
  const canGoNext = currentIndex >= 0 && currentIndex < currentPageTraceIds.length - 1
  const totalInPage = currentPageTraceIds.length
  const position = currentIndex >= 0 ? currentIndex + 1 : 0

  // ==========================================================================
  // Actions
  // ==========================================================================

  /**
   * Open trace detail in peek mode (sheet sidebar)
   * Uses router.push() to create history entry
   */
  const openTrace = useCallback(
    (id: string) => {
      const params = new URLSearchParams(searchParams.toString())
      params.set('peek', id)
      // Reset span selection when opening new trace
      params.delete('span')
      router.push(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  /**
   * Close trace detail
   * Uses router.push() to create history entry
   */
  const closeTrace = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete('peek')
    params.delete('span')
    params.delete('view')
    router.push(`${pathname}?${params.toString()}`)
  }, [router, pathname, searchParams])

  /**
   * Expand to full page view by navigating to /traces/[traceId]
   * @param newTab - If true, opens in new browser tab
   */
  const expandToFullPage = useCallback(
    (newTab: boolean = false) => {
      if (!traceId || !projectSlug) return

      // Build the full page URL with span and view params if set
      const fullPageUrl = `/projects/${projectSlug}/traces/${traceId}`
      const params = new URLSearchParams()
      if (selectedSpanId) {
        params.set('span', selectedSpanId)
      }
      if (viewMode !== 'tree') {
        params.set('view', viewMode)
      }
      const urlWithParams = params.toString() ? `${fullPageUrl}?${params.toString()}` : fullPageUrl

      if (newTab) {
        window.open(urlWithParams, '_blank')
      } else {
        router.push(urlWithParams)
      }
    },
    [router, traceId, projectSlug, selectedSpanId, viewMode]
  )

  /**
   * Select a span (or clear selection)
   * Uses router.replace() to avoid history clutter
   */
  const selectSpan = useCallback(
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
   * Uses router.replace() to avoid history clutter
   */
  const setViewMode = useCallback(
    (mode: ViewMode) => {
      const params = new URLSearchParams(searchParams.toString())
      if (mode !== 'tree') {
        params.set('view', mode)
      } else {
        params.delete('view') // 'tree' is default
      }
      router.replace(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  /**
   * Navigate to previous trace in list
   */
  const goToPrev = useCallback(() => {
    if (!canGoPrev) return
    const prevId = currentPageTraceIds[currentIndex - 1]
    openTrace(prevId)
  }, [canGoPrev, currentPageTraceIds, currentIndex, openTrace])

  /**
   * Navigate to next trace in list
   */
  const goToNext = useCallback(() => {
    if (!canGoNext) return
    const nextId = currentPageTraceIds[currentIndex + 1]
    openTrace(nextId)
  }, [canGoNext, currentPageTraceIds, currentIndex, openTrace])

  return {
    // URL State
    traceId,
    selectedSpanId,
    viewMode,

    // Navigation State
    canGoPrev,
    canGoNext,
    position,
    totalInPage,

    // Actions
    openTrace,
    closeTrace,
    expandToFullPage,
    selectSpan,
    setViewMode,
    goToPrev,
    goToNext,
  }
}
