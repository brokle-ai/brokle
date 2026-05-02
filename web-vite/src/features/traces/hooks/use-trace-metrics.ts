import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format, isAfter, startOfDay, subDays, subHours } from 'date-fns'
import { useProjectOnly } from '@/features/projects'
import { traceListQueryOptions } from '../api/queries'
import type { TraceSummary } from '../api/types'

/**
 * Time range options for the metrics view. `all` skips the start
 * bound; the rest translate to a unix-seconds floor sent to the
 * backend `start_time` query param.
 */
export const TIME_RANGES = ['24h', '7d', '30d', 'all'] as const
export type TimeRange = (typeof TIME_RANGES)[number]

/**
 * Aggregated metrics derived from a batch of traces. The web-vite
 * variant wires off `TraceSummary` (the list-endpoint wire shape) so
 * cost is parsed from a decimal string (`shopspring/decimal` JSON form)
 * and durations stay in nanoseconds until the format helpers convert
 * to milliseconds at the render boundary.
 */
export interface TraceMetrics {
  totalTraces: number
  totalTokens: number
  totalCost: number
  averageLatency: number // ms
  errorRate: number // percentage 0-100

  byModel: Array<{
    model: string
    count: number
    tokens: number
    cost: number
  }>

  byProvider: Array<{
    provider: string
    count: number
    tokens: number
    cost: number
  }>

  timeSeries: Array<{
    date: string
    traces: number
    tokens: number
    cost: number
    errors: number
  }>
}

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

function parseCost(value: string | undefined | null): number {
  if (!value) return 0
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

function groupByDay(
  traces: TraceSummary[],
  startDate: Date | null,
): Map<string, TraceSummary[]> {
  const groups = new Map<string, TraceSummary[]>()
  for (const trace of traces) {
    const start = new Date(trace.start_time)
    if (startDate && start < startDate) continue
    const dateKey = format(start, 'yyyy-MM-dd')
    const bucket = groups.get(dateKey)
    if (bucket) bucket.push(trace)
    else groups.set(dateKey, [trace])
  }
  return groups
}

function calculateMetrics(
  traces: TraceSummary[],
  timeRange: TimeRange,
): TraceMetrics {
  const startDate = getStartDate(timeRange)

  const filtered = startDate
    ? traces.filter((t) => isAfter(new Date(t.start_time), startDate))
    : traces

  const totalTraces = filtered.length
  const totalTokens = filtered.reduce((sum, t) => sum + (t.total_tokens || 0), 0)
  const totalCost = filtered.reduce((sum, t) => sum + parseCost(t.total_cost), 0)
  const totalDurationNs = filtered.reduce((sum, t) => sum + (t.duration || 0), 0)
  const errorCount = filtered.filter((t) => t.has_error).length

  const averageLatency =
    totalTraces > 0 ? totalDurationNs / totalTraces / 1_000_000 : 0
  const errorRate = totalTraces > 0 ? (errorCount / totalTraces) * 100 : 0

  const modelMap = new Map<
    string,
    { count: number; tokens: number; cost: number }
  >()
  for (const trace of filtered) {
    const model = trace.model_name || 'unknown'
    const entry = modelMap.get(model) ?? { count: 0, tokens: 0, cost: 0 }
    entry.count += 1
    entry.tokens += trace.total_tokens || 0
    entry.cost += parseCost(trace.total_cost)
    modelMap.set(model, entry)
  }
  const byModel = Array.from(modelMap.entries())
    .map(([model, stats]) => ({ model, ...stats }))
    .sort((a, b) => b.count - a.count)

  const providerMap = new Map<
    string,
    { count: number; tokens: number; cost: number }
  >()
  for (const trace of filtered) {
    const provider = trace.provider_name || 'unknown'
    const entry = providerMap.get(provider) ?? { count: 0, tokens: 0, cost: 0 }
    entry.count += 1
    entry.tokens += trace.total_tokens || 0
    entry.cost += parseCost(trace.total_cost)
    providerMap.set(provider, entry)
  }
  const byProvider = Array.from(providerMap.entries())
    .map(([provider, stats]) => ({ provider, ...stats }))
    .sort((a, b) => b.count - a.count)

  const dailyGroups = groupByDay(filtered, startDate)
  const timeSeries: TraceMetrics['timeSeries'] = []
  const today = new Date()
  const rangeStart = startDate ?? subDays(today, 30)
  const cursor = new Date(startOfDay(rangeStart))
  while (cursor <= today) {
    const dateKey = format(cursor, 'yyyy-MM-dd')
    const dayTraces = dailyGroups.get(dateKey) ?? []
    timeSeries.push({
      date: format(cursor, 'MMM d'),
      traces: dayTraces.length,
      tokens: dayTraces.reduce((sum, t) => sum + (t.total_tokens || 0), 0),
      cost: dayTraces.reduce((sum, t) => sum + parseCost(t.total_cost), 0),
      errors: dayTraces.filter((t) => t.has_error).length,
    })
    cursor.setDate(cursor.getDate() + 1)
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

export interface UseTraceMetricsReturn {
  metrics: TraceMetrics | null
  timeRange: TimeRange
  setTimeRange: (range: TimeRange) => void
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
}

/**
 * Drives the metrics view. Fetches up to 1000 traces in the selected
 * window via the existing list endpoint and aggregates client-side.
 * The backend has no dedicated metrics endpoint yet, so this matches
 * web/'s strategy: trade off precision (capped at 1000) for zero
 * additional API surface.
 */
export function useTraceMetrics(): UseTraceMetricsReturn {
  const { currentProject, hasProject } = useProjectOnly()
  const projectId = currentProject?.id ?? ''
  const [timeRange, setTimeRange] = useState<TimeRange>('7d')

  const {
    data,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useQuery({
    ...traceListQueryOptions(projectId, {
      page: 1,
      limit: 100,
      range: timeRange,
    }),
    enabled: hasProject && projectId.length > 0,
  })

  const metrics = useMemo(() => {
    if (!data?.data) return null
    return calculateMetrics(data.data, timeRange)
  }, [data, timeRange])

  return {
    metrics,
    timeRange,
    setTimeRange,
    isLoading,
    isFetching,
    error: error instanceof Error ? error.message : error ? String(error) : null,
    refetch: () => {
      void refetch()
    },
    hasProject,
  }
}
