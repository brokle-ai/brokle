'use client'

import { useState, useEffect, useCallback } from 'react'
import type { RowHeight } from '../components/cells/types'

const STORAGE_KEY = 'datasets-row-height'
const DEFAULT_HEIGHT: RowHeight = 'medium'

export function useRowHeight() {
  const [rowHeight, setRowHeightState] = useState<RowHeight>(DEFAULT_HEIGHT)
  const [isLoaded, setIsLoaded] = useState(false)

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored && isValidRowHeight(stored)) {
        setRowHeightState(stored as RowHeight)
      }
    } catch {
      // localStorage not available
    }
    setIsLoaded(true)
  }, [])

  // Save to localStorage when changed
  const setRowHeight = useCallback((height: RowHeight) => {
    setRowHeightState(height)
    try {
      localStorage.setItem(STORAGE_KEY, height)
    } catch {
      // localStorage not available
    }
  }, [])

  return {
    rowHeight,
    setRowHeight,
    isLoaded,
  }
}

function isValidRowHeight(value: string): value is RowHeight {
  return ['small', 'medium', 'large'].includes(value)
}
