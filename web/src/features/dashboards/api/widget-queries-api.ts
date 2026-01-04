/**
 * Widget Queries API Client
 *
 * API functions for executing dashboard widget queries using BrokleAPIClient.
 */

import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  DashboardQueryResults,
  WidgetQueryResult,
  QueryExecutionParams,
  ViewDefinitionsResponse,
} from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Execute all widget queries for a dashboard
 */
export const executeDashboardQueries = async (
  projectId: string,
  dashboardId: string,
  params?: QueryExecutionParams
): Promise<DashboardQueryResults> => {
  return client.post<DashboardQueryResults>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/execute`,
    params ?? {}
  )
}

/**
 * Execute a single widget query
 */
export const executeWidgetQuery = async (
  projectId: string,
  dashboardId: string,
  widgetId: string,
  params?: QueryExecutionParams
): Promise<WidgetQueryResult> => {
  return client.post<WidgetQueryResult>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/widgets/${widgetId}/execute`,
    params ?? {}
  )
}

/**
 * Get available view definitions for the query builder
 * Returns measures and dimensions for each view type (traces, spans, scores)
 */
export const getViewDefinitions = async (): Promise<ViewDefinitionsResponse> => {
  return client.get<ViewDefinitionsResponse>('/v1/dashboards/view-definitions')
}
