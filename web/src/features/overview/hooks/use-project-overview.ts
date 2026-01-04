'use client'

import { useState, useMemo, useCallback, useEffect } from 'react'
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
  const [isOnboardingDismissed, setIsOnboardingDismissed] = useState(false)

  // Load dismissed state from localStorage
  useEffect(() => {
    if (projectId && typeof window !== 'undefined') {
      const dismissed = localStorage.getItem(`${ONBOARDING_DISMISSED_KEY}:${projectId}`)
      setIsOnboardingDismissed(dismissed === 'true')
    }
  }, [projectId])

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
      setIsOnboardingDismissed(true)
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
  }, [overview?.checklist_status])

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
