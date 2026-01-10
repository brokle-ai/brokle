'use client'

import { useQuery } from '@tanstack/react-query'
import type { TimeRange } from '@/components/shared/time-range-picker'
import {
  getUsageOverview,
  getUsageTimeSeries,
  getUsageByProject,
} from '../api/usage-api'

// Query keys for cache management
export const usageQueryKeys = {
  all: ['usage'] as const,
  overview: (orgId: string) => [...usageQueryKeys.all, 'overview', orgId] as const,
  timeSeries: (orgId: string, timeRange: TimeRange, granularity?: string) =>
    [...usageQueryKeys.all, 'timeseries', orgId, timeRange, granularity] as const,
  byProject: (orgId: string, timeRange: TimeRange) =>
    [...usageQueryKeys.all, 'by-project', orgId, timeRange] as const,
}

export function useUsageOverviewQuery(organizationId: string | undefined) {
  return useQuery({
    queryKey: usageQueryKeys.overview(organizationId ?? ''),
    queryFn: () => getUsageOverview(organizationId!),
    enabled: !!organizationId,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
  })
}

export function useUsageTimeSeriesQuery(
  organizationId: string | undefined,
  timeRange: TimeRange,
  granularity?: 'hourly' | 'daily'
) {
  return useQuery({
    queryKey: usageQueryKeys.timeSeries(organizationId ?? '', timeRange, granularity),
    queryFn: () => getUsageTimeSeries(organizationId!, timeRange, granularity),
    enabled: !!organizationId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useUsageByProjectQuery(
  organizationId: string | undefined,
  timeRange: TimeRange
) {
  return useQuery({
    queryKey: usageQueryKeys.byProject(organizationId ?? '', timeRange),
    queryFn: () => getUsageByProject(organizationId!, timeRange),
    enabled: !!organizationId,
    staleTime: 5 * 60 * 1000,
  })
}
