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

export const widgetQueryKeys = {
  all: ['widget-queries'] as const,
  dashboardResults: () => [...widgetQueryKeys.all, 'dashboard'] as const,
  dashboardResult: (
    projectId: string,
    dashboardId: string,
    timeRange?: TimeRange,
  ) =>
    [
      ...widgetQueryKeys.dashboardResults(),
      projectId,
      dashboardId,
      timeRange,
    ] as const,
  widgetResults: () => [...widgetQueryKeys.all, 'widget'] as const,
  widgetResult: (
    projectId: string,
    dashboardId: string,
    widgetId: string,
    timeRange?: TimeRange,
  ) =>
    [
      ...widgetQueryKeys.widgetResults(),
      projectId,
      dashboardId,
      widgetId,
      timeRange,
    ] as const,
  viewDefinitions: () => [...widgetQueryKeys.all, 'view-definitions'] as const,
}

export interface UseDashboardQueriesOptions {
  enabled?: boolean
  timeRange?: TimeRange
  variableValues?: VariableValues
  refetchInterval?: number
}

export function useDashboardQueries(
  projectId: string | undefined,
  dashboardId: string | undefined,
  options: UseDashboardQueriesOptions = {},
) {
  const {
    enabled = true,
    timeRange,
    variableValues,
    refetchInterval = 0,
  } = options

  return useQuery({
    queryKey: [
      ...widgetQueryKeys.dashboardResult(
        projectId ?? '',
        dashboardId ?? '',
        timeRange,
      ),
      variableValues,
    ],
    queryFn: async (): Promise<DashboardQueryResults> => {
      if (!projectId || !dashboardId) {
        throw new Error('Project ID and Dashboard ID are required')
      }
      const params: QueryExecutionParams = {}
      if (timeRange) params.time_range = timeRange
      if (variableValues && Object.keys(variableValues).length > 0) {
        params.variable_values = variableValues
      }
      return executeDashboardQueries(projectId, dashboardId, params)
    },
    enabled: !!projectId && !!dashboardId && enabled,
    staleTime: 10_000,
    gcTime: 2 * 60_000,
    refetchInterval: refetchInterval > 0 ? refetchInterval : undefined,
  })
}

export function useWidgetQuery(
  projectId: string | undefined,
  dashboardId: string | undefined,
  widgetId: string | undefined,
  options: UseDashboardQueriesOptions = {},
) {
  const { enabled = true, timeRange } = options

  return useQuery({
    queryKey: widgetQueryKeys.widgetResult(
      projectId ?? '',
      dashboardId ?? '',
      widgetId ?? '',
      timeRange,
    ),
    queryFn: async (): Promise<WidgetQueryResult> => {
      if (!projectId || !dashboardId || !widgetId) {
        throw new Error(
          'Project ID, Dashboard ID, and Widget ID are required',
        )
      }
      const params: QueryExecutionParams = {}
      if (timeRange) params.time_range = timeRange
      return executeWidgetQuery(projectId, dashboardId, widgetId, params)
    },
    enabled: !!projectId && !!dashboardId && !!widgetId && enabled,
    staleTime: 10_000,
    gcTime: 2 * 60_000,
  })
}

export function useRefreshDashboardQueries(
  projectId: string,
  dashboardId: string,
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (params?: QueryExecutionParams) =>
      executeDashboardQueries(projectId, dashboardId, {
        ...params,
        force_refresh: true,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: widgetQueryKeys.dashboardResults(),
        predicate: (query) => {
          const key = query.queryKey as unknown as string[]
          return key[2] === projectId && key[3] === dashboardId
        },
      })
    },
  })
}

export function useViewDefinitions(options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: widgetQueryKeys.viewDefinitions(),
    queryFn: getViewDefinitions,
    enabled: options.enabled ?? true,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
  })
}
