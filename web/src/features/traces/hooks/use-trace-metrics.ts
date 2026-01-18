'use client'

import { useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryState, parseAsStringLiteral } from 'nuqs'
import { subDays, subHours, startOfDay, format, isAfter } from 'date-fns'
import { useProjectOnly } from '@/features/projects'
import { traceQueryKeys } from './trace-query-keys'
import { getProjectTraces } from '../api/traces-api'
import type { Trace } from '../data/schema'

/**
 * Time range options for metrics
 */
export const TIME_RANGES = ['24h', '7d', '30d', 'all'] as const
export type TimeRange = (typeof TIME_RANGES)[number]

/**
 * Aggregated metrics data
 */
export interface TraceMetrics {
  // Summary stats
  totalTraces: number
  totalTokens: number
  totalCost: number
  averageLatency: number // ms
  errorRate: number // percentage

  // By model breakdown
  byModel: Array<{
    model: string
    count: number
    tokens: number
    cost: number
  }>

  // By provider breakdown
  byProvider: Array<{
    provider: string
    count: number
    tokens: number
    cost: number
  }>

  // Time series data for charts
  timeSeries: Array<{
    date: string
    traces: number
    tokens: number
    cost: number
    errors: number
  }>
}

/**
 * Get start date for time range
 */
function getStartDate(range: TimeRange): Date | null {
  const now = new Date()
  switch (range) {
    case '24h':
      return subHours(now, 24)
    case '7d':
      return subDays(now, 7)
    case '30d':
      return subDays(now, 30)
    case 'all':
      return null
  }
}

/**
 * Group traces into daily buckets
 */
function groupByDay(traces: Trace[], startDate: Date | null): Map<string, Trace[]> {
  const groups = new Map<string, Trace[]>()

  traces.forEach((trace) => {
    if (startDate && trace.start_time < startDate) return

    const dateKey = format(trace.start_time, 'yyyy-MM-dd')

    if (!groups.has(dateKey)) {
      groups.set(dateKey, [])
    }
    groups.get(dateKey)!.push(trace)
  })

  return groups
}

/**
 * Calculate metrics from traces
 */
function calculateMetrics(traces: Trace[], timeRange: TimeRange): TraceMetrics {
  const startDate = getStartDate(timeRange)

  // Filter traces by time range
  const filteredTraces = startDate
    ? traces.filter((t) => isAfter(t.start_time, startDate))
    : traces

  // Calculate summary stats
  const totalTraces = filteredTraces.length
  const totalTokens = filteredTraces.reduce((sum, t) => sum + (t.tokens || 0), 0)
  const totalCost = filteredTraces.reduce((sum, t) => sum + (t.cost || 0), 0)
  const totalDuration = filteredTraces.reduce((sum, t) => sum + (t.duration || 0), 0)
  const errorCount = filteredTraces.filter((t) => t.has_error).length

  const averageLatency = totalTraces > 0 ? totalDuration / totalTraces / 1_000_000 : 0 // Convert ns to ms
  const errorRate = totalTraces > 0 ? (errorCount / totalTraces) * 100 : 0

  // Group by model
  const modelMap = new Map<string, { count: number; tokens: number; cost: number }>()
  filteredTraces.forEach((trace) => {
    const model = trace.model_name || 'unknown'
    if (!modelMap.has(model)) {
      modelMap.set(model, { count: 0, tokens: 0, cost: 0 })
    }
    const entry = modelMap.get(model)!
    entry.count++
    entry.tokens += trace.tokens || 0
    entry.cost += trace.cost || 0
  })

  const byModel = Array.from(modelMap.entries())
    .map(([model, stats]) => ({ model, ...stats }))
    .sort((a, b) => b.count - a.count)

  // Group by provider
  const providerMap = new Map<string, { count: number; tokens: number; cost: number }>()
  filteredTraces.forEach((trace) => {
    const provider = trace.provider_name || 'unknown'
    if (!providerMap.has(provider)) {
      providerMap.set(provider, { count: 0, tokens: 0, cost: 0 })
    }
    const entry = providerMap.get(provider)!
    entry.count++
    entry.tokens += trace.tokens || 0
    entry.cost += trace.cost || 0
  })

  const byProvider = Array.from(providerMap.entries())
    .map(([provider, stats]) => ({ provider, ...stats }))
    .sort((a, b) => b.count - a.count)

  // Build time series
  const dailyGroups = groupByDay(filteredTraces, startDate)
  const timeSeries: TraceMetrics['timeSeries'] = []

  // Generate date range for time series
  const today = new Date()
  const rangeStart = startDate || subDays(today, 30)
  let currentDate = startOfDay(rangeStart)

  while (currentDate <= today) {
    const dateKey = format(currentDate, 'yyyy-MM-dd')
    const dayTraces = dailyGroups.get(dateKey) || []

    timeSeries.push({
      date: format(currentDate, 'MMM d'),
      traces: dayTraces.length,
      tokens: dayTraces.reduce((sum, t) => sum + (t.tokens || 0), 0),
      cost: dayTraces.reduce((sum, t) => sum + (t.cost || 0), 0),
      errors: dayTraces.filter((t) => t.has_error).length,
    })

    currentDate = new Date(currentDate)
    currentDate.setDate(currentDate.getDate() + 1)
  }

  return {
    totalTraces,
    totalTokens,
    totalCost,
    averageLatency,
    errorRate,
    byModel,
    byProvider,
    timeSeries,
  }
}

export interface UseMetricsStateReturn {
  timeRange: TimeRange
  setTimeRange: (range: TimeRange) => void
}

/**
 * Hook to manage metrics time range via URL state
 */
export function useMetricsState(): UseMetricsStateReturn {
  const [timeRange, setTimeRange] = useQueryState(
    'metricsRange',
    parseAsStringLiteral(TIME_RANGES).withDefault('7d')
  )

  return { timeRange, setTimeRange }
}

/**
 * Hook to fetch and compute trace metrics
 */
export function useTraceMetrics() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { timeRange, setTimeRange } = useMetricsState()
  const projectId = currentProject?.id

  const {
    data: tracesData,
    isLoading: isTracesLoading,
    isFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: traceQueryKeys.metrics(projectId!, { timeRange }),

    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')

      // Fetch traces for the time range
      // For metrics, we need a larger batch
      const startDate = getStartDate(timeRange)

      const result = await getProjectTraces({
        projectId,
        page: 1,
        pageSize: 1000, // Fetch larger batch for metrics
        startTime: startDate || undefined,
      })

      return result
    },

    enabled: !!projectId && hasProject,

    staleTime: 60_000, // 1 minute - metrics can be slightly stale
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,

    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  // Calculate metrics from traces
  const metrics = useMemo((): TraceMetrics | null => {
    if (!tracesData?.traces) return null
    return calculateMetrics(tracesData.traces, timeRange)
  }, [tracesData?.traces, timeRange])

  return {
    metrics,
    timeRange,
    setTimeRange,

    isLoading: isProjectLoading || isTracesLoading,
    isFetching,

    error: error instanceof Error ? error.message : error ? String(error) : null,

    refetch,

    hasProject,
    currentProject,
  }
}

export interface UseTraceMetricsReturn {
  metrics: TraceMetrics | null
  timeRange: TimeRange
  setTimeRange: (range: TimeRange) => void
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
}
