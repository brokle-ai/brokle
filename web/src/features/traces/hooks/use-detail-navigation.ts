'use client'

import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'next/navigation'
import { useTraces } from '../context/traces-context'
import { usePeekNavigation } from './use-peek-navigation'

/**
 * Hook for navigating prev/next through traces in peek view
 * Limited to current page only (server-side pagination limitation)
 */
export function useDetailNavigation() {
  const searchParams = useSearchParams()
  const peekId = searchParams.get('peek')
  const { currentPageTraceIds } = useTraces()
  const { openPeek } = usePeekNavigation()

  // Find current index in the list
  const currentIndex = useMemo(() => {
    if (!peekId || !currentPageTraceIds.length) return -1
    return currentPageTraceIds.indexOf(peekId)
  }, [peekId, currentPageTraceIds])

  const canGoPrev = currentIndex > 0
  const canGoNext = currentIndex >= 0 && currentIndex < currentPageTraceIds.length - 1

  const totalInPage = currentPageTraceIds.length
  const position = currentIndex >= 0 ? currentIndex + 1 : 0

  const handlePrev = useCallback(() => {
    if (!canGoPrev) return
    const prevId = currentPageTraceIds[currentIndex - 1]
    openPeek(prevId)
  }, [canGoPrev, currentPageTraceIds, currentIndex, openPeek])

  const handleNext = useCallback(() => {
    if (!canGoNext) return
    const nextId = currentPageTraceIds[currentIndex + 1]
    openPeek(nextId)
  }, [canGoNext, currentPageTraceIds, currentIndex, openPeek])

  return {
    canGoPrev,
    canGoNext,
    handlePrev,
    handleNext,
    position,
    totalInPage,
  }
}
