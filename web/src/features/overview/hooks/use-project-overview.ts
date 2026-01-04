'use client'

import { useState, useMemo, useCallback, useSyncExternalStore } from 'react'
import { useOverviewQuery } from './use-overview-queries'
import type { OverviewTimeRange, OverviewResponse } from '../types'

const ONBOARDING_DISMISSED_KEY = 'brokle:onboarding-dismissed'

interface UseProjectOverviewOptions {
  initialTimeRange?: OverviewTimeRange
}

export function useProjectOverview(
  projectId: string | undefined,
  options: UseProjectOverviewOptions = {}
) {
  const { initialTimeRange = '24h' } = options

  const [timeRange, setTimeRange] = useState<OverviewTimeRange>(initialTimeRange)

  // Derive dismissed state from localStorage using useSyncExternalStore
  const storageKey = projectId ? `${ONBOARDING_DISMISSED_KEY}:${projectId}` : null

  const isOnboardingDismissed = useSyncExternalStore(
    // Subscribe to storage changes
    useCallback(
      (callback) => {
        if (!storageKey || typeof window === 'undefined') return () => {}

        const handler = (e: StorageEvent) => {
          if (e.key === storageKey) callback()
        }
        window.addEventListener('storage', handler)
        return () => window.removeEventListener('storage', handler)
      },
      [storageKey]
    ),
    // Get current value (client)
    useCallback(() => {
      if (!storageKey || typeof window === 'undefined') return false
      return localStorage.getItem(storageKey) === 'true'
    }, [storageKey]),
    // Server snapshot (SSR)
    () => false
  )

  const {
    data,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useOverviewQuery(projectId, timeRange)

  const handleTimeRangeChange = useCallback((newTimeRange: OverviewTimeRange) => {
    setTimeRange(newTimeRange)
  }, [])

  const dismissOnboarding = useCallback(() => {
    if (projectId && typeof window !== 'undefined') {
      localStorage.setItem(`${ONBOARDING_DISMISSED_KEY}:${projectId}`, 'true')
      // Trigger re-render via storage event for same-tab updates
      window.dispatchEvent(
        new StorageEvent('storage', {
          key: `${ONBOARDING_DISMISSED_KEY}:${projectId}`,
          newValue: 'true',
        })
      )
    }
  }, [projectId])

  // Memoize the overview data
  const overview = useMemo<OverviewResponse | null>(() => {
    return data ?? null
  }, [data])

  // Calculate onboarding progress
  const onboardingProgress = useMemo(() => {
    if (!overview?.checklist_status) return { completed: 0, total: 4, percentage: 0 }

    const { checklist_status } = overview
    const completed = [
      checklist_status.has_project,
      checklist_status.has_traces,
      checklist_status.has_ai_provider,
      checklist_status.has_evaluations,
    ].filter(Boolean).length

    return {
      completed,
      total: 4,
      percentage: (completed / 4) * 100,
    }
  }, [overview])

  // Check if onboarding is complete
  const isOnboardingComplete = onboardingProgress.completed === onboardingProgress.total

  return {
    data: overview,
    stats: overview?.stats ?? null,
    traceVolume: overview?.trace_volume ?? [],
    costByModel: overview?.cost_by_model ?? [],
    recentTraces: overview?.recent_traces ?? [],
    topErrors: overview?.top_errors ?? [],
    scoresSummary: overview?.scores_summary ?? [],
    checklistStatus: overview?.checklist_status ?? null,
    onboardingProgress,
    isOnboardingComplete,
    isOnboardingDismissed,
    dismissOnboarding,
    timeRange,
    setTimeRange: handleTimeRangeChange,
    isLoading,
    isRefetching: isFetching && !isLoading,
    error,
    hasProject: !!projectId,
    refetch,
  }
}
