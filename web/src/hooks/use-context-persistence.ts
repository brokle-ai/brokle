'use client'

import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '@/features/authentication'

interface PersistedContext {
  organizationSlug: string
  projectSlug?: string
  timestamp: number
  userEmail: string
}

const STORAGE_KEY = 'brokle_last_context'
const MAX_AGE = 30 * 24 * 60 * 60 * 1000 // 30 days in milliseconds

export function useContextPersistence() {
  const { user } = useAuth()
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    setIsLoaded(true)
  }, [])

  const saveLastContext = useCallback((
    organizationSlug: string,
    projectSlug?: string
  ) => {
    if (!isLoaded || !user?.email) return

    const contextData: PersistedContext = {
      organizationSlug,
      projectSlug,
      timestamp: Date.now(),
      userEmail: user.email,
    }

    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(contextData))
    } catch (error) {
      console.warn('[ContextPersistence] Failed to save context to localStorage:', error)
    }
  }, [isLoaded, user?.email])

  const getLastContext = useCallback((): {
    organizationSlug: string
    projectSlug?: string
  } | null => {
    if (!isLoaded || !user?.email) return null

    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (!stored) return null

      const contextData: PersistedContext = JSON.parse(stored)
      
      // Check if context is for the current user
      if (contextData.userEmail !== user.email) {
        return null
      }

      // Check if context is not too old
      if (Date.now() - contextData.timestamp > MAX_AGE) {
        localStorage.removeItem(STORAGE_KEY)
        return null
      }

      return {
        organizationSlug: contextData.organizationSlug,
        projectSlug: contextData.projectSlug,
      }
    } catch (error) {
      console.warn('[ContextPersistence] Failed to parse context from localStorage:', error)
      // Clean up corrupted data
      localStorage.removeItem(STORAGE_KEY)
      return null
    }
  }, [isLoaded, user?.email])

  const clearLastContext = useCallback(() => {
    if (!isLoaded) return

    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
      console.warn('[ContextPersistence] Failed to clear context from localStorage:', error)
    }
  }, [isLoaded])

  useEffect(() => {
    if (isLoaded) {
      // Clean up any old context keys
      const oldKeys = ['brokle_org', 'brokle_project', 'last_org', 'last_project']
      oldKeys.forEach(key => {
        try {
          localStorage.removeItem(key)
        } catch {
          // Ignore cleanup errors
        }
      })
    }
  }, [isLoaded])

  return {
    saveLastContext,
    getLastContext,
    clearLastContext,
    isLoaded,
  }
}

