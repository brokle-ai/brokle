'use client'

import { useState, useEffect, useCallback } from 'react'
import type { VisibilityState } from '@tanstack/react-table'

const STORAGE_KEY = 'brokle-columns-traces'

/**
 * Hook for managing column visibility state with localStorage persistence.
 *
 * @param defaultVisibility - Default visibility state for columns
 * @returns Visibility state and setter functions
 */
export function useColumnVisibility(defaultVisibility: VisibilityState = {}) {
  const [visibility, setVisibilityState] = useState<VisibilityState>(() => {
    if (typeof window === 'undefined') return defaultVisibility
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      return stored ? JSON.parse(stored) : defaultVisibility
    } catch {
      return defaultVisibility
    }
  })

  // Persist to localStorage whenever visibility changes
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(visibility))
    } catch {
      // Ignore storage errors (e.g., quota exceeded, private browsing)
    }
  }, [visibility])

  const setVisibility = useCallback(
    (
      updater: VisibilityState | ((prev: VisibilityState) => VisibilityState)
    ) => {
      setVisibilityState((prev) =>
        typeof updater === 'function' ? updater(prev) : updater
      )
    },
    []
  )

  const resetVisibility = useCallback(() => {
    setVisibilityState(defaultVisibility)
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch {
      // Ignore storage errors
    }
  }, [defaultVisibility])

  return { visibility, setVisibility, resetVisibility }
}
