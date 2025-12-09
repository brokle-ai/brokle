'use client'

import { useState, useEffect, useCallback } from 'react'

/**
 * useLocalStorage - Hook for persisting state to localStorage
 *
 * Features:
 * - SSR-safe (checks typeof window)
 * - Type-safe with generics
 * - Cross-tab synchronization via storage events
 *
 * @param key - localStorage key
 * @param defaultValue - Default value if key doesn't exist
 * @returns [value, setValue] tuple
 */
export function useLocalStorage<T>(
  key: string,
  defaultValue: T
): [T, (value: T) => void] {
  // Initialize state with default value (SSR-safe)
  const [storedValue, setStoredValue] = useState<T>(defaultValue)
  const [isInitialized, setIsInitialized] = useState(false)

  // Read from localStorage on mount (client-side only)
  useEffect(() => {
    if (typeof window === 'undefined') return

    try {
      const item = window.localStorage.getItem(key)
      if (item !== null) {
        setStoredValue(JSON.parse(item) as T)
      }
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error)
    }
    setIsInitialized(true)
  }, [key])

  // Setter that also persists to localStorage
  const setValue = useCallback(
    (value: T) => {
      try {
        setStoredValue(value)
        if (typeof window !== 'undefined') {
          window.localStorage.setItem(key, JSON.stringify(value))
          // Dispatch custom event for cross-tab sync
          window.dispatchEvent(
            new CustomEvent('local-storage-change', { detail: { key, value } })
          )
        }
      } catch (error) {
        console.warn(`Error setting localStorage key "${key}":`, error)
      }
    },
    [key]
  )

  // Listen for changes from other tabs/windows
  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === key && event.newValue !== null) {
        try {
          setStoredValue(JSON.parse(event.newValue) as T)
        } catch (error) {
          console.warn(`Error parsing localStorage change for key "${key}":`, error)
        }
      }
    }

    // Listen for custom events from same tab
    const handleCustomEvent = (event: CustomEvent<{ key: string; value: T }>) => {
      if (event.detail.key === key) {
        setStoredValue(event.detail.value)
      }
    }

    window.addEventListener('storage', handleStorageChange)
    window.addEventListener('local-storage-change', handleCustomEvent as EventListener)

    return () => {
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('local-storage-change', handleCustomEvent as EventListener)
    }
  }, [key])

  return [storedValue, setValue]
}
