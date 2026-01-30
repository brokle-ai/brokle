'use client'

import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryState, parseAsStringLiteral } from 'nuqs'
import { format, parseISO } from 'date-fns'
import { useProjectOnly } from '@/features/projects'
import { getProjectOverview } from '@/features/overview/api/overview-api'
import type { OverviewResponse, TimeSeriesPoint, CostByModel } from '@/features/overview/types'
import type { TimeRange as OverviewTimeRange } from '@/components/shared/time-range-picker'

/**
 * Time range options for metrics view
 * Note: 'all' now maps to the backend 'all' option (capped at 365 days)
 */
export const TIME_RANGES = ['24h', '7d', '30d', 'all'] as const
export type TimeRange = (typeof TIME_RANGES)[number]

/**
 * Map frontend time range to backend TimeRange format
 */
function mapToOverviewTimeRange(range: TimeRange): OverviewTimeRange {
  switch (range) {
    case '24h':
      return { relative: '24h' }
    case '7d':
      return { relative: '7d' }
    case '30d':
      return { relative: '30d' }
    case 'all':
      return { relative: 'all' }
  }
}

/**
 * Aggregated metrics data
 * Uses the overview endpoint's server-side aggregated data for accuracy.
 */
export interface TraceMetrics {
  // Summary stats (accurate - server-side aggregated)
  totalTraces: number
  totalTokens: number
  totalCost: number
  averageLatency: number // ms
  errorRate: number // percentage

  // By model breakdown with tokens
  byModel: Array<{
    model: string
    count: number
    tokens: number
    cost: number
  }>

  // By provider breakdown (derived from model names)
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
 * Extract provider name from model name
 */
function extractProvider(modelName: string): string {
  const lowerModel = modelName.toLowerCase()
  if (lowerModel.includes('gpt') || lowerModel.includes('o1') || lowerModel.includes('o3') || lowerModel.includes('openai')) {
    return 'OpenAI'
  }
  if (lowerModel.includes('claude') || lowerModel.includes('anthropic')) {
    return 'Anthropic'
  }
  if (lowerModel.includes('gemini') || lowerModel.includes('google') || lowerModel.includes('palm')) {
    return 'Google'
  }
  if (lowerModel.includes('llama') || lowerModel.includes('meta')) {
    return 'Meta'
  }
  if (lowerModel.includes('mistral') || lowerModel.includes('mixtral')) {
    return 'Mistral'
  }
  if (lowerModel.includes('cohere') || lowerModel.includes('command')) {
    return 'Cohere'
  }
  return 'Other'
}

/**
 * Transform overview response to TraceMetrics format
 */
function transformOverviewToMetrics(overview: OverviewResponse): TraceMetrics {
  const { stats, trace_volume, cost_time_series, token_time_series, error_time_series, cost_by_model } = overview

  // Transform cost_by_model to byModel format with tokens and count
  const byModel: TraceMetrics['byModel'] = (cost_by_model || []).map((item: CostByModel) => ({
    model: item.model,
    count: item.count,
    tokens: item.tokens,
    cost: item.cost,
  }))

  // Aggregate by provider
  const providerMap = new Map<string, { count: number; tokens: number; cost: number }>()
  for (const item of byModel) {
    const provider = extractProvider(item.model)
    const existing = providerMap.get(provider) || { count: 0, tokens: 0, cost: 0 }
    existing.count += item.count
    existing.tokens += item.tokens
    existing.cost += item.cost
    providerMap.set(provider, existing)
  }

  const byProvider: TraceMetrics['byProvider'] = Array.from(providerMap.entries())
    .map(([provider, data]) => ({
      provider,
      ...data,
    }))
    .sort((a, b) => b.cost - a.cost)

  // Build time series by aligning all data points
  // Create a map from timestamp to metrics
  const timeSeriesMap = new Map<string, { traces: number; tokens: number; cost: number; errors: number }>()

  // Process trace volume
  for (const point of trace_volume || []) {
    const key = point.timestamp
    const existing = timeSeriesMap.get(key) || { traces: 0, tokens: 0, cost: 0, errors: 0 }
    existing.traces = point.value
    timeSeriesMap.set(key, existing)
  }

  // Process cost time series
  for (const point of cost_time_series || []) {
    const key = point.timestamp
    const existing = timeSeriesMap.get(key) || { traces: 0, tokens: 0, cost: 0, errors: 0 }
    existing.cost = point.value
    timeSeriesMap.set(key, existing)
  }

  // Process token time series
  for (const point of token_time_series || []) {
    const key = point.timestamp
    const existing = timeSeriesMap.get(key) || { traces: 0, tokens: 0, cost: 0, errors: 0 }
    existing.tokens = point.value
    timeSeriesMap.set(key, existing)
  }

  // Process error time series
  for (const point of error_time_series || []) {
    const key = point.timestamp
    const existing = timeSeriesMap.get(key) || { traces: 0, tokens: 0, cost: 0, errors: 0 }
    existing.errors = point.value
    timeSeriesMap.set(key, existing)
  }

  // Convert map to sorted array
  const timeSeries: TraceMetrics['timeSeries'] = Array.from(timeSeriesMap.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([timestamp, data]) => {
      // Parse timestamp and format for display
      let dateLabel: string
      try {
        const date = parseISO(timestamp)
        dateLabel = format(date, 'MMM d')
      } catch {
        dateLabel = timestamp
      }

      return {
        date: dateLabel,
        ...data,
      }
    })

  return {
    totalTraces: stats.traces_count,
    totalTokens: stats.total_tokens,
    totalCost: stats.total_cost,
    averageLatency: stats.avg_latency_ms,
    errorRate: stats.error_rate,
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
    'range',
    parseAsStringLiteral(TIME_RANGES).withDefault('7d')
  )

  return { timeRange, setTimeRange }
}

/**
 * Hook to fetch trace metrics using the overview endpoint.
 *
 * This uses server-side aggregation via the overview API, providing accurate
 * totals across ALL traces in the project, not limited to a sample size.
 */
export function useTraceMetrics() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { timeRange, setTimeRange } = useMetricsState()
  const projectId = currentProject?.id

  const {
    data: overviewData,
    isLoading: isOverviewLoading,
    isFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: ['projects', projectId, 'overview', timeRange],

    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')

      // Use the overview endpoint for accurate server-side aggregation
      const overviewTimeRange = mapToOverviewTimeRange(timeRange)
      return getProjectOverview(projectId, overviewTimeRange)
    },

    enabled: !!projectId && hasProject,

    staleTime: 60_000, // 1 minute - metrics can be slightly stale
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,

    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  // Transform overview data to TraceMetrics format
  const metrics = useMemo((): TraceMetrics | null => {
    if (!overviewData) return null
    return transformOverviewToMetrics(overviewData)
  }, [overviewData])

  return {
    metrics,
    timeRange,
    setTimeRange,

    isLoading: isProjectLoading || isOverviewLoading,
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
