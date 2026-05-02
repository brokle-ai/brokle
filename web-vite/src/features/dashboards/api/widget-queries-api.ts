/**
 * Widget Queries API
 *
 * Executes widget queries server-side. Returns results keyed by
 * widget ID for batch execution and a single result for per-widget
 * execution.
 */

import { rawFetch } from '@/lib/api/client'
import type {
  DashboardQueryResults,
  WidgetQueryResult,
  QueryExecutionParams,
  ViewDefinitionsResponse,
} from '../types'

async function requestJson<T>(
  input: string,
  init: RequestInit = {},
): Promise<T> {
  const resp = await rawFetch(input, init)
  return (await resp.json()) as T
}

export const executeDashboardQueries = async (
  projectId: string,
  dashboardId: string,
  params?: QueryExecutionParams,
): Promise<DashboardQueryResults> => {
  return requestJson<DashboardQueryResults>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/execute`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params ?? {}),
    },
  )
}

export const executeWidgetQuery = async (
  projectId: string,
  dashboardId: string,
  widgetId: string,
  params?: QueryExecutionParams,
): Promise<WidgetQueryResult> => {
  return requestJson<WidgetQueryResult>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/widgets/${encodeURIComponent(widgetId)}/execute`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params ?? {}),
    },
  )
}

export const getViewDefinitions =
  async (): Promise<ViewDefinitionsResponse> => {
    return requestJson<ViewDefinitionsResponse>(
      '/api/v1/dashboards/view-definitions',
      { method: 'GET' },
    )
  }
