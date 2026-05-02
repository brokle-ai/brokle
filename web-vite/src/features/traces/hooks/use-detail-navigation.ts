import { useCallback, useMemo } from 'react'
import { useSearch } from '@tanstack/react-router'
import { useTraces } from '../context/traces-context'
import { usePeekNavigation } from './use-peek-navigation'

// Prev/next navigation through the current page of traces. Limited to
// the in-memory page because the list endpoint is page-paginated with
// no cursor-based "next after X" affordance — matches web/.

interface PeekSearchShape {
  peek?: string
  [key: string]: unknown
}

export function useDetailNavigation() {
  const search = useSearch({ strict: false }) as PeekSearchShape
  const peekId = typeof search.peek === 'string' ? search.peek : null
  const { currentPageTraceIds } = useTraces()
  const { openPeek } = usePeekNavigation()

  const currentIndex = useMemo(() => {
    if (!peekId || currentPageTraceIds.length === 0) return -1
    return currentPageTraceIds.indexOf(peekId)
  }, [peekId, currentPageTraceIds])

  const canGoPrev = currentIndex > 0
  const canGoNext =
    currentIndex >= 0 && currentIndex < currentPageTraceIds.length - 1
  const totalInPage = currentPageTraceIds.length
  const position = currentIndex >= 0 ? currentIndex + 1 : 0

  const handlePrev = useCallback(() => {
    if (!canGoPrev) return
    openPeek(currentPageTraceIds[currentIndex - 1])
  }, [canGoPrev, currentPageTraceIds, currentIndex, openPeek])

  const handleNext = useCallback(() => {
    if (!canGoNext) return
    openPeek(currentPageTraceIds[currentIndex + 1])
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
