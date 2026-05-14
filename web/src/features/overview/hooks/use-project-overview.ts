'use client'

import { useMemo } from 'react'
import { useOverviewQuery } from './use-overview-queries'
import type { TimeRange } from '@/components/shared/time-range-picker'
import type { OverviewResponse } from '../types'

// Default time range
const DEFAULT_TIME_RANGE: TimeRange = { relative: '24h' }

interface UseProjectOverviewOptions {
  /** Time range for the overview data query */
  timeRange?: TimeRange
}

export function useProjectOverview(
  projectId: string | undefined,
  options: UseProjectOverviewOptions = {}
) {
  const { timeRange = DEFAULT_TIME_RANGE } = options

  const {
    data,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useOverviewQuery(projectId, timeRange)

  // Memoize the overview data
  const overview = useMemo<OverviewResponse | null>(() => {
    return data ?? null
  }, [data])

  return {
    data: overview,
    stats: overview?.stats ?? null,
    traceVolume: overview?.trace_volume ?? [],
    costByModel: overview?.cost_by_model ?? [],
    recentTraces: overview?.recent_traces ?? [],
    topErrors: overview?.top_errors ?? [],
    scoresSummary: overview?.scores_summary ?? [],
    checklistStatus: overview?.checklist_status ?? null,
    isLoading,
    isRefetching: isFetching && !isLoading,
    error,
    hasProject: !!projectId,
    refetch,
  }
}
