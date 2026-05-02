/**
 * Auto-Save Hook for Dashboards
 */

import { useEffect, useRef, useState, useCallback } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateDashboard } from '../api/dashboards-api'
import { dashboardQueryKeys } from './use-dashboards-queries'
import type { Dashboard, UpdateDashboardRequest } from '../types'

export type AutoSaveStatus = 'idle' | 'pending' | 'saving' | 'saved' | 'error'

interface UseAutoSaveOptions {
  enabled?: boolean
  debounceMs?: number
  savedDurationMs?: number
  onSaveSuccess?: () => void
  onSaveError?: (error: Error) => void
}

interface UseAutoSaveReturn {
  status: AutoSaveStatus
  hasUnsavedChanges: boolean
  saveNow: () => void
  scheduleSave: (data: Partial<UpdateDashboardRequest>) => void
  cancelPendingSave: () => void
  error: string | null
}

export function useAutoSave(
  projectId: string,
  dashboardId: string,
  dashboard: Dashboard | null | undefined,
  options: UseAutoSaveOptions = {},
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
  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const savedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

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

      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })

      savedTimerRef.current = setTimeout(() => {
        setStatus('idle')
      }, savedDurationMs)

      onSaveSuccess?.()
    },
    onError: (err) => {
      setStatus('error')
      const message =
        err instanceof Error ? err.message : 'Failed to save dashboard'
      setError(message)
      onSaveError?.(err instanceof Error ? err : new Error(message))
    },
  })

  const executeSave = useCallback(() => {
    if (!pendingDataRef.current || !enabled) return
    const dataToSave = pendingDataRef.current
    saveMutation.mutate(dataToSave)
  }, [enabled, saveMutation])

  const scheduleSave = useCallback(
    (data: Partial<UpdateDashboardRequest>) => {
      if (!enabled) return

      pendingDataRef.current = pendingDataRef.current
        ? { ...pendingDataRef.current, ...data }
        : data

      setHasUnsavedChanges(true)
      setStatus('pending')
      setError(null)

      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
      debounceTimerRef.current = setTimeout(executeSave, debounceMs)
    },
    [enabled, debounceMs, executeSave],
  )

  const saveNow = useCallback(() => {
    if (!enabled || !pendingDataRef.current) return
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
      debounceTimerRef.current = null
    }
    executeSave()
  }, [enabled, executeSave])

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

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) clearTimeout(debounceTimerRef.current)
      if (savedTimerRef.current) clearTimeout(savedTimerRef.current)
    }
  }, [])

  useEffect(() => {
    if (!enabled || !hasUnsavedChanges) return

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = ''
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
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
