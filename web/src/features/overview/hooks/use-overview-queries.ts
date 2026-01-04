'use client'

import { useQuery } from '@tanstack/react-query'
import { getProjectOverview } from '../api/overview-api'
import type { TimeRange } from '@/components/shared/time-range-picker'

// Default time range
const DEFAULT_TIME_RANGE: TimeRange = { relative: '24h' }

/**
 * Create a stable query key from TimeRange
 */
function createTimeRangeKey(timeRange: TimeRange): string {
  if (timeRange.relative === 'custom' && timeRange.from && timeRange.to) {
    return `custom:${timeRange.from}:${timeRange.to}`
  }
  return timeRange.relative || '24h'
}

/**
 * Query keys for overview queries
 */
export const overviewQueryKeys = {
  all: ['overview'] as const,
  project: (projectId: string, timeRange: TimeRange) =>
    [...overviewQueryKeys.all, projectId, createTimeRangeKey(timeRange)] as const,
}

/**
 * Query hook to get project overview data
 */
export function useOverviewQuery(
  projectId: string | undefined,
  timeRange: TimeRange = DEFAULT_TIME_RANGE,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: overviewQueryKeys.project(projectId || '', timeRange),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return getProjectOverview(projectId, timeRange)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 60_000, // 1 minute - overview data updates frequently
    gcTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 60_000, // Auto-refresh every minute
  })
}
