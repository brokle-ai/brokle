'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  executeDashboardQueries,
  executeWidgetQuery,
  getViewDefinitions,
} from '../api/widget-queries-api'
import type {
  DashboardQueryResults,
  WidgetQueryResult,
  QueryExecutionParams,
  TimeRange,
  VariableValues,
} from '../types'

/**
 * Query keys for widget query execution
 */
export const widgetQueryKeys = {
  all: ['widget-queries'] as const,

  // Dashboard query results
  dashboardResults: () => [...widgetQueryKeys.all, 'dashboard'] as const,
  dashboardResult: (
    projectId: string,
    dashboardId: string,
    timeRange?: TimeRange
  ) =>
    [...widgetQueryKeys.dashboardResults(), projectId, dashboardId, timeRange] as const,

  // Single widget query results
  widgetResults: () => [...widgetQueryKeys.all, 'widget'] as const,
  widgetResult: (
    projectId: string,
    dashboardId: string,
    widgetId: string,
    timeRange?: TimeRange
  ) =>
    [
      ...widgetQueryKeys.widgetResults(),
      projectId,
      dashboardId,
      widgetId,
      timeRange,
    ] as const,

  // View definitions
  viewDefinitions: () => [...widgetQueryKeys.all, 'view-definitions'] as const,
}

/**
 * Options for dashboard queries hook
 */
export interface UseDashboardQueriesOptions {
  /** Whether the query is enabled */
  enabled?: boolean
  /** Time range for queries */
  timeRange?: TimeRange
  /** Variable values to substitute in queries */
  variableValues?: VariableValues
  /** Auto-refresh interval in milliseconds (0 to disable) */
  refetchInterval?: number
}

/**
 * Hook to execute all widget queries for a dashboard
 *
 * This hook fetches query results for all widgets in a dashboard.
 * Results are cached and can be auto-refreshed at a configurable interval.
 *
 * @example
 * ```tsx
 * const { data, isLoading, error, refetch } = useDashboardQueries(
 *   projectId,
 *   dashboardId,
 *   { timeRange: { relative: '24h' }, refetchInterval: 30000 }
 * )
 *
 * // Access widget data
 * const widgetData = data?.results[widgetId]?.data
 * ```
 */
export function useDashboardQueries(
  projectId: string | undefined,
  dashboardId: string | undefined,
  options: UseDashboardQueriesOptions = {}
) {
  const { enabled = true, timeRange, variableValues, refetchInterval = 0 } = options

  return useQuery({
    queryKey: [
      ...widgetQueryKeys.dashboardResult(
        projectId || '',
        dashboardId || '',
        timeRange
      ),
      variableValues, // Include variables in query key for cache invalidation
    ],
    queryFn: async (): Promise<DashboardQueryResults> => {
      if (!projectId || !dashboardId) {
        throw new Error('Project ID and Dashboard ID are required')
      }

      const params: QueryExecutionParams = {}
      if (timeRange) {
        params.time_range = timeRange
      }
      if (variableValues && Object.keys(variableValues).length > 0) {
        params.variable_values = variableValues
      }

      return executeDashboardQueries(projectId, dashboardId, params)
    },
    enabled: !!projectId && !!dashboardId && enabled,
    staleTime: 10_000, // 10 seconds - widget data is fairly fresh
    gcTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: refetchInterval > 0 ? refetchInterval : undefined,
  })
}

/**
 * Hook to execute a single widget query
 *
 * Use this for on-demand widget refresh without refetching all widgets.
 */
export function useWidgetQuery(
  projectId: string | undefined,
  dashboardId: string | undefined,
  widgetId: string | undefined,
  options: UseDashboardQueriesOptions = {}
) {
  const { enabled = true, timeRange } = options

  return useQuery({
    queryKey: widgetQueryKeys.widgetResult(
      projectId || '',
      dashboardId || '',
      widgetId || '',
      timeRange
    ),
    queryFn: async (): Promise<WidgetQueryResult> => {
      if (!projectId || !dashboardId || !widgetId) {
        throw new Error('Project ID, Dashboard ID, and Widget ID are required')
      }

      const params: QueryExecutionParams = {}
      if (timeRange) {
        params.time_range = timeRange
      }

      return executeWidgetQuery(projectId, dashboardId, widgetId, params)
    },
    enabled: !!projectId && !!dashboardId && !!widgetId && enabled,
    staleTime: 10_000,
    gcTime: 2 * 60 * 1000,
  })
}

/**
 * Mutation hook to refresh dashboard queries with force refresh
 *
 * Use this when the user explicitly clicks a refresh button.
 */
export function useRefreshDashboardQueries(
  projectId: string,
  dashboardId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params?: QueryExecutionParams) => {
      return executeDashboardQueries(projectId, dashboardId, {
        ...params,
        force_refresh: true,
      })
    },
    onSuccess: (data) => {
      // Update the cache with fresh results
      queryClient.setQueryData(
        widgetQueryKeys.dashboardResult(
          projectId,
          dashboardId,
          data.executed_at ? undefined : undefined
        ),
        data
      )

      // Invalidate to ensure any filtered queries are also updated
      queryClient.invalidateQueries({
        queryKey: widgetQueryKeys.dashboardResults(),
        predicate: (query) => {
          const key = query.queryKey as string[]
          return key[2] === projectId && key[3] === dashboardId
        },
      })
    },
  })
}

/**
 * Hook to get view definitions for the query builder
 *
 * Returns available measures and dimensions for each view type.
 * This data rarely changes so it has a long cache time.
 */
export function useViewDefinitions(options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: widgetQueryKeys.viewDefinitions(),
    queryFn: getViewDefinitions,
    enabled: options.enabled ?? true,
    staleTime: 5 * 60 * 1000, // 5 minutes - view definitions rarely change
    gcTime: 30 * 60 * 1000, // 30 minutes
  })
}
