'use client'

import { useQuery } from '@tanstack/react-query'
import { getProjectOverview } from '../api/overview-api'
import type { OverviewTimeRange } from '../types'

/**
 * Query keys for overview queries
 */
export const overviewQueryKeys = {
  all: ['overview'] as const,
  project: (projectId: string, timeRange: OverviewTimeRange) =>
    [...overviewQueryKeys.all, projectId, timeRange] as const,
}

/**
 * Query hook to get project overview data
 */
export function useOverviewQuery(
  projectId: string | undefined,
  timeRange: OverviewTimeRange = '24h',
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
