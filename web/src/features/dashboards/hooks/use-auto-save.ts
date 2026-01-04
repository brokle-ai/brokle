/**
 * Auto-Save Hook for Dashboards
 *
 * Provides debounced auto-save functionality with save status indicator.
 * Automatically saves dashboard changes after 1 second of inactivity.
 */

import { useEffect, useRef, useState, useCallback } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateDashboard } from '../api/dashboards-api'
import type { Dashboard, UpdateDashboardRequest } from '../types'

export type AutoSaveStatus = 'idle' | 'pending' | 'saving' | 'saved' | 'error'

interface UseAutoSaveOptions {
  /** Enable auto-save (default: true) */
  enabled?: boolean
  /** Debounce delay in milliseconds (default: 1000) */
  debounceMs?: number
  /** Duration to show "saved" status before returning to idle (default: 2000) */
  savedDurationMs?: number
  /** Callback when save succeeds */
  onSaveSuccess?: () => void
  /** Callback when save fails */
  onSaveError?: (error: Error) => void
}

interface UseAutoSaveReturn {
  /** Current save status */
  status: AutoSaveStatus
  /** Whether there are unsaved changes */
  hasUnsavedChanges: boolean
  /** Trigger an immediate save */
  saveNow: () => void
  /** Schedule a debounced save */
  scheduleSave: (data: Partial<UpdateDashboardRequest>) => void
  /** Clear any pending save */
  cancelPendingSave: () => void
  /** Error message if save failed */
  error: string | null
}

export function useAutoSave(
  projectId: string,
  dashboardId: string,
  dashboard: Dashboard | null | undefined,
  options: UseAutoSaveOptions = {}
): UseAutoSaveReturn {
  const {
    enabled = true,
    debounceMs = 1000,
    savedDurationMs = 2000,
    onSaveSuccess,
    onSaveError,
  } = options

  const queryClient = useQueryClient()
  const [status, setStatus] = useState<AutoSaveStatus>('idle')
  const [error, setError] = useState<string | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)

  const pendingDataRef = useRef<Partial<UpdateDashboardRequest> | null>(null)
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null)
  const savedTimerRef = useRef<NodeJS.Timeout | null>(null)

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: async (data: Partial<UpdateDashboardRequest>) => {
      if (!dashboard) throw new Error('Dashboard not loaded')
      return updateDashboard(projectId, dashboardId, data)
    },
    onMutate: () => {
      setStatus('saving')
      setError(null)
    },
    onSuccess: () => {
      setStatus('saved')
      setHasUnsavedChanges(false)
      pendingDataRef.current = null

      // Invalidate dashboard query to refresh data
      queryClient.invalidateQueries({
        queryKey: ['dashboard', projectId, dashboardId],
      })

      // Return to idle after savedDurationMs
      savedTimerRef.current = setTimeout(() => {
        setStatus('idle')
      }, savedDurationMs)

      onSaveSuccess?.()
    },
    onError: (err) => {
      setStatus('error')
      const message = err instanceof Error ? err.message : 'Failed to save dashboard'
      setError(message)
      onSaveError?.(err instanceof Error ? err : new Error(message))
    },
  })

  // Execute the save
  const executeSave = useCallback(() => {
    if (!pendingDataRef.current || !enabled) return

    const dataToSave = pendingDataRef.current
    saveMutation.mutate(dataToSave)
  }, [enabled, saveMutation])

  // Schedule a debounced save
  const scheduleSave = useCallback(
    (data: Partial<UpdateDashboardRequest>) => {
      if (!enabled) return

      // Merge with any existing pending data
      pendingDataRef.current = pendingDataRef.current
        ? { ...pendingDataRef.current, ...data }
        : data

      setHasUnsavedChanges(true)
      setStatus('pending')
      setError(null)

      // Clear any existing timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      // Schedule new save
      debounceTimerRef.current = setTimeout(() => {
        executeSave()
      }, debounceMs)
    },
    [enabled, debounceMs, executeSave]
  )

  // Trigger immediate save
  const saveNow = useCallback(() => {
    if (!enabled || !pendingDataRef.current) return

    // Clear debounce timer
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
      debounceTimerRef.current = null
    }

    executeSave()
  }, [enabled, executeSave])

  // Cancel pending save
  const cancelPendingSave = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
      debounceTimerRef.current = null
    }
    pendingDataRef.current = null
    setHasUnsavedChanges(false)
    setStatus('idle')
    setError(null)
  }, [])

  // Cleanup timers on unmount
  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
      if (savedTimerRef.current) {
        clearTimeout(savedTimerRef.current)
      }
    }
  }, [])

  // Save before unload if there are pending changes
  useEffect(() => {
    if (!enabled || !hasUnsavedChanges) return

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      // Modern browsers require returnValue to be set
      e.returnValue = ''
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [enabled, hasUnsavedChanges])

  return {
    status,
    hasUnsavedChanges,
    saveNow,
    scheduleSave,
    cancelPendingSave,
    error,
  }
}

/**
 * Get a human-readable label for the auto-save status
 */
export function getAutoSaveStatusLabel(status: AutoSaveStatus): string {
  switch (status) {
    case 'idle':
      return ''
    case 'pending':
      return 'Unsaved changes'
    case 'saving':
      return 'Saving...'
    case 'saved':
      return 'Saved'
    case 'error':
      return 'Save failed'
    default:
      return ''
  }
}
