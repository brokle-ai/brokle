/**
 * Drilldown Utilities
 *
 * Map a clicked widget data point to a traces-page filter URL. The
 * URL is resolved by the caller — we emit a relative path + query
 * string and let the caller plug it into TanStack's <Link to=...>
 * or router.navigate().
 */

import type { WidgetQuery } from '../types'

export interface DrilldownFilter {
  id: string
  column: string
  operator: '=' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'in'
  value: string | number | string[] | null
}

const DIMENSION_TO_COLUMN_MAP: Record<string, string> = {
  time: 'started_at',
  timestamp: 'started_at',
  model: 'model',
  model_name: 'model',
  provider: 'provider',
  provider_name: 'provider',
  status: 'status',
  trace_name: 'name',
  name: 'name',
  trace_id: 'id',
  evaluation_name: 'name',
  score_name: 'score_name',
  user_id: 'user_id',
  session_id: 'session_id',
  environment: 'environment',
}

function generateFilterId(): string {
  return `drilldown_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
}

export function buildFilterFromDataPoint(
  dimensionName: string,
  dimensionValue: string | number,
): DrilldownFilter {
  const column = DIMENSION_TO_COLUMN_MAP[dimensionName] ?? dimensionName
  return {
    id: generateFilterId(),
    column,
    operator: '=',
    value:
      typeof dimensionValue === 'number'
        ? dimensionValue
        : String(dimensionValue),
  }
}

export function buildFiltersFromDataPoint(
  data: Record<string, unknown>,
  query: WidgetQuery,
): DrilldownFilter[] {
  const filters: DrilldownFilter[] = []
  const dimensions = query.dimensions ?? []

  for (const dimension of dimensions) {
    let value: unknown

    if (dimensions.length === 1 && 'name' in data) {
      value = data.name
    } else {
      value = data[dimension]
    }

    if (value !== undefined && value !== null && value !== '') {
      filters.push(
        buildFilterFromDataPoint(dimension, value as string | number),
      )
    }
  }

  return filters
}

export function encodeFiltersForUrl(filters: DrilldownFilter[]): string {
  if (filters.length === 0) return ''
  return encodeURIComponent(JSON.stringify(filters))
}

export interface DrilldownTimeRange {
  from?: string
  to?: string
  relative?: string
}

/**
 * Build a traces drilldown URL. `projectSlug` here is the
 * `{orgId}/{projectId}` composite slug (web-vite router uses those
 * as path params); we emit a full path that TanStack's navigate can
 * consume.
 */
export function buildDrilldownUrl(
  orgId: string,
  projectId: string,
  filters: DrilldownFilter[],
  timeRange?: DrilldownTimeRange,
): string {
  const params = new URLSearchParams()

  if (filters.length > 0) {
    params.set('filters', JSON.stringify(filters))
  }

  if (timeRange?.relative && timeRange.relative !== 'custom') {
    params.set('time_rel', timeRange.relative)
  } else if (timeRange?.from && timeRange?.to) {
    params.set('time_from', timeRange.from)
    params.set('time_to', timeRange.to)
  }

  const queryString = params.toString()
  return `/o/${orgId}/p/${projectId}/traces${queryString ? `?${queryString}` : ''}`
}

export interface DrilldownContext {
  orgId: string
  projectId: string
  query: WidgetQuery
  timeRange?: DrilldownTimeRange
}

export function createDrilldownHandler(
  context: DrilldownContext,
  navigate: (url: string) => void,
) {
  return (data: Record<string, unknown>) => {
    const filters = buildFiltersFromDataPoint(data, context.query)
    const url = buildDrilldownUrl(
      context.orgId,
      context.projectId,
      filters,
      context.timeRange,
    )
    navigate(url)
  }
}
