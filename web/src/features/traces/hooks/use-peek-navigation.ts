'use client'

import { useCallback } from 'react'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'

/**
 * Hook for managing peek sheet navigation via URL parameters
 * Uses ?peek=traceId for overlay view, full page navigation for expansion
 */
export function usePeekNavigation() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  /**
   * Open peek sheet with trace ID
   * Uses router.push() to create history entry (back button works)
   */
  const openPeek = useCallback(
    (traceId: string) => {
      const params = new URLSearchParams(searchParams.toString())
      params.set('peek', traceId)
      // Reset tab when opening new trace
      params.delete('tab')
      router.push(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  /**
   * Close peek sheet
   * Uses router.push() to create history entry
   */
  const closePeek = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete('peek')
    params.delete('tab')
    router.push(`${pathname}?${params.toString()}`)
  }, [router, pathname, searchParams])

  /**
   * Expand peek to full page
   * Navigates to /traces/[traceId] preserving tab parameter
   */
  const expandPeek = useCallback(
    (newTab: boolean = false) => {
      const peekId = searchParams.get('peek')
      if (!peekId) return

      const tab = searchParams.get('tab')
      const fullPageUrl = `${pathname}/${peekId}${tab ? `?tab=${tab}` : ''}`

      if (newTab) {
        window.open(fullPageUrl, '_blank')
      } else {
        router.push(fullPageUrl)
      }
    },
    [router, pathname, searchParams]
  )

  /**
   * Set active tab in peek view
   * Uses router.replace() to avoid creating history entries for tab switches
   */
  const setTab = useCallback(
    (tab: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (tab) {
        params.set('tab', tab)
      } else {
        params.delete('tab')
      }
      router.replace(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams]
  )

  return {
    openPeek,
    closePeek,
    expandPeek,
    setTab,
  }
}
