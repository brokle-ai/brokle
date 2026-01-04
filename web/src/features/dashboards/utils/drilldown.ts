/**
 * Drilldown Utilities
 *
 * Functions to build filters from widget data points and generate
 * drilldown URLs to the traces page.
 */

import type { WidgetQuery } from '../types'

/**
 * Filter condition for traces page (matches traces API format)
 */
export interface DrilldownFilter {
  id: string
  column: string
  operator: '=' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'in'
  value: string | number | string[] | null
}

/**
 * Map dashboard dimension names to trace filter columns
 * This handles the mapping between widget query dimensions and
 * the actual filterable columns in the traces table.
 */
const DIMENSION_TO_COLUMN_MAP: Record<string, string> = {
  // Time-based
  time: 'started_at',
  timestamp: 'started_at',

  // Model dimensions
  model: 'model',
  model_name: 'model',
  provider: 'provider',
  provider_name: 'provider',

  // Trace dimensions
  status: 'status',
  trace_name: 'name',
  name: 'name',
  trace_id: 'id',

  // Evaluation dimensions
  evaluation_name: 'name',
  score_name: 'score_name',

  // User dimensions
  user_id: 'user_id',
  session_id: 'session_id',

  // Environment
  environment: 'environment',
}

/**
 * Generate a unique ID for filter conditions
 */
function generateFilterId(): string {
  return `drilldown_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
}

/**
 * Build a filter condition from a data point
 */
export function buildFilterFromDataPoint(
  dimensionName: string,
  dimensionValue: string | number
): DrilldownFilter {
  // Map dimension name to column name
  const column = DIMENSION_TO_COLUMN_MAP[dimensionName] || dimensionName

  return {
    id: generateFilterId(),
    column,
    operator: '=',
    value: typeof dimensionValue === 'number' ? dimensionValue : String(dimensionValue),
  }
}

/**
 * Build multiple filters from a widget's dimensions and clicked data point
 *
 * @param data - The clicked data point (typically has name/value structure)
 * @param query - The widget's query configuration
 * @returns Array of filter conditions for the traces page
 */
export function buildFiltersFromDataPoint(
  data: Record<string, unknown>,
  query: WidgetQuery
): DrilldownFilter[] {
  const filters: DrilldownFilter[] = []
  const dimensions = query.dimensions || []

  // The data point typically has dimension values as string fields
  // and measure values as number fields
  for (const dimension of dimensions) {
    // Try to find the dimension value in the data point
    // It might be stored as 'name' for single dimension charts
    // or directly by the dimension key
    let value: unknown

    if (dimensions.length === 1 && 'name' in data) {
      // Single dimension chart - value is in 'name' field
      value = data.name
    } else {
      // Multi-dimension chart - look for the dimension key
      value = data[dimension]
    }

    if (value !== undefined && value !== null && value !== '') {
      filters.push(buildFilterFromDataPoint(dimension, value as string | number))
    }
  }

  return filters
}

/**
 * Encode filters as a URL query string for the traces page
 */
export function encodeFiltersForUrl(filters: DrilldownFilter[]): string {
  if (filters.length === 0) return ''
  return encodeURIComponent(JSON.stringify(filters))
}

/**
 * Build the full drilldown URL to the traces page
 *
 * @param projectSlug - The project slug for the URL
 * @param filters - Filter conditions to apply
 * @param timeRange - Optional time range to preserve
 * @returns Full URL path to traces page with filters
 */
export function buildDrilldownUrl(
  projectSlug: string,
  filters: DrilldownFilter[],
  timeRange?: { from?: string; to?: string; relative?: string }
): string {
  const params = new URLSearchParams()

  // Add filters
  if (filters.length > 0) {
    params.set('filters', JSON.stringify(filters))
  }

  // Add time range if present
  if (timeRange?.relative && timeRange.relative !== 'custom') {
    params.set('time_rel', timeRange.relative)
  } else if (timeRange?.from && timeRange?.to) {
    params.set('time_from', timeRange.from)
    params.set('time_to', timeRange.to)
  }

  const queryString = params.toString()
  return `/projects/${projectSlug}/traces${queryString ? `?${queryString}` : ''}`
}

/**
 * Hook-friendly type for drilldown click handlers
 */
export interface DrilldownContext {
  projectSlug: string
  query: WidgetQuery
  timeRange?: { from?: string; to?: string; relative?: string }
}

/**
 * Create a click handler function for chart data points
 *
 * @param context - Drilldown context with project and query info
 * @param navigate - Function to perform navigation (e.g., router.push)
 * @returns Click handler for chart elements
 */
export function createDrilldownHandler(
  context: DrilldownContext,
  navigate: (url: string) => void
) {
  return (data: Record<string, unknown>) => {
    const filters = buildFiltersFromDataPoint(data, context.query)
    const url = buildDrilldownUrl(context.projectSlug, filters, context.timeRange)
    navigate(url)
  }
}
