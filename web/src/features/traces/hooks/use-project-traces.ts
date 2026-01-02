'use client'

import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useProjectOnly } from '@/features/projects'
import { useTracesTableState } from './use-traces-table-state'
import { getProjectTraces, getTraceFilterOptions } from '../api/traces-api'
import type { Trace } from '../data/schema'
import type { TraceFilterOptions, GetTracesParams } from '../api/traces-api'
import type { UseTracesTableStateReturn } from './use-traces-table-state'

/**
 * Maps frontend column IDs (from TanStack Table) to backend sort field names.
 * Only columns with valid backend mappings are sortable.
 */
const SORT_FIELD_MAP: Record<string, string> = {
  // Direct matches
  model_name: 'model_name',
  service_name: 'service_name',

  // Renamed fields
  duration: 'trace_duration_nano',
  cost: 'total_cost',
  tokens: 'total_tokens',
  spanCount: 'span_count',
  start_time: 'trace_start',

  // Additional sortable fields
  input_tokens: 'input_tokens',
  output_tokens: 'output_tokens',
  end_time: 'trace_end',
  error_span_count: 'error_span_count',
}

/**
 * Hook to fetch and manage project traces with filtering, sorting, and pagination.
 *
 * Uses the centralized useTracesTableState hook for URL state management.
 * All filters, pagination, sorting, and search state is persisted in the URL.
 *
 * @returns Traces data, pagination, loading state, error state, and table state
 */
export function useProjectTraces() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = useTracesTableState()
  const projectId = currentProject?.id

  // Map filter conditions to API params
  const apiParams = useMemo((): GetTracesParams | null => {
    if (!projectId) return null

    const params: GetTracesParams = {
      projectId,
      page: tableState.page,
      pageSize: tableState.pageSize,
      search: tableState.search || undefined,
      searchType: tableState.searchType || undefined,
      sortBy: tableState.sortBy ? SORT_FIELD_MAP[tableState.sortBy] : undefined,
      sortOrder: tableState.sortOrder || undefined,
    }

    // Map structured filters to API params
    const statusValues: string[] = []
    const statusNotValues: string[] = []

    tableState.filters.forEach((filter) => {
      // Handle status_code separately - it supports IN/NOT IN with multiple values
      if (filter.column === 'status_code') {
        const statusMap: Record<string, string> = {
          '0': 'unset',
          '1': 'ok',
          '2': 'error',
        }

        // Normalize to array for consistent handling (supports both single and multi-select)
        const values = Array.isArray(filter.value)
          ? filter.value
          : filter.value !== null
            ? [filter.value]
            : []

        const isExclusion = filter.operator === '!=' || filter.operator === 'NOT IN'
        const targetArray = isExclusion ? statusNotValues : statusValues

        values.forEach((v) => {
          const mapped = statusMap[String(v)] || String(v)
          if (!targetArray.includes(mapped)) {
            targetArray.push(mapped)
          }
        })
        return // Skip to next filter
      }

      // For all other columns, use single value (they don't support multi-select)
      const value = Array.isArray(filter.value) ? filter.value[0] : filter.value

      switch (filter.column) {
        // String equals filters
        case 'model_name':
          params.modelName = String(value)
          break
        case 'provider_name':
          params.providerName = String(value)
          break
        case 'service_name':
          params.serviceName = String(value)
          break
        case 'session_id':
          params.sessionId = String(value)
          break
        case 'user_id':
          params.userId = String(value)
          break

        // Numeric range filters - map operator to min/max
        case 'total_cost':
          if (filter.operator === '>' || filter.operator === '>=') {
            params.minCost = Number(value)
          } else if (filter.operator === '<' || filter.operator === '<=') {
            params.maxCost = Number(value)
          } else if (filter.operator === '=') {
            params.minCost = Number(value)
            params.maxCost = Number(value)
          }
          break
        case 'total_tokens':
          if (filter.operator === '>' || filter.operator === '>=') {
            params.minTokens = Number(value)
          } else if (filter.operator === '<' || filter.operator === '<=') {
            params.maxTokens = Number(value)
          } else if (filter.operator === '=') {
            params.minTokens = Number(value)
            params.maxTokens = Number(value)
          }
          break
        case 'input_tokens':
          // Map to total tokens for now (backend uses total_tokens)
          if (filter.operator === '>' || filter.operator === '>=') {
            params.minTokens = Number(value)
          } else if (filter.operator === '<' || filter.operator === '<=') {
            params.maxTokens = Number(value)
          }
          break
        case 'output_tokens':
          // Map to total tokens for now (backend uses total_tokens)
          if (filter.operator === '>' || filter.operator === '>=') {
            params.minTokens = Number(value)
          } else if (filter.operator === '<' || filter.operator === '<=') {
            params.maxTokens = Number(value)
          }
          break
        case 'duration_nano':
          if (filter.operator === '>' || filter.operator === '>=') {
            params.minDuration = Number(value)
          } else if (filter.operator === '<' || filter.operator === '<=') {
            params.maxDuration = Number(value)
          } else if (filter.operator === '=') {
            params.minDuration = Number(value)
            params.maxDuration = Number(value)
          }
          break

        // Boolean filter
        case 'has_error':
          params.hasError = String(value) === 'true'
          break

        // DateTime filters
        case 'start_time':
          if (filter.operator === '>' || filter.operator === '>=') {
            params.startTime = new Date(String(value))
          }
          break
        case 'end_time':
          if (filter.operator === '<' || filter.operator === '<=') {
            params.endTime = new Date(String(value))
          }
          break
      }
    })

    if (statusValues.length > 0) {
      params.status = statusValues
    }
    if (statusNotValues.length > 0) {
      params.statusNot = statusNotValues
    }

    return params
  }, [projectId, tableState.page, tableState.pageSize, tableState.search, tableState.searchType, tableState.sortBy, tableState.sortOrder, tableState.filters])

  const {
    data,
    isLoading: isTracesLoading,
    isFetching: isTracesFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: ['traces', apiParams],

    queryFn: () => {
      if (!apiParams) {
        throw new Error('No project selected')
      }
      return getProjectTraces(apiParams)
    },

    // Only fetch when we have valid API params
    enabled: !!apiParams && hasProject,

    // Cache configuration
    staleTime: 30_000, // 30 seconds - data is fresh for this duration
    gcTime: 5 * 60 * 1000, // 5 minutes - keep unused data in cache
    refetchOnWindowFocus: true, // Refetch when user returns to tab
    refetchOnReconnect: true, // Refetch when internet reconnects

    // Retry configuration
    retry: 2, // Retry failed requests twice
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000), // Exponential backoff

    // NOTE: We intentionally do NOT use keepPreviousData here.
    // Instead, we show loading state inside the table body (Langfuse pattern).
    // This prevents blinking while providing clear feedback during data fetches.
  })

  return {
    // Data
    data: data?.traces ?? [],
    totalCount: data?.totalCount ?? 0,
    page: data?.page ?? tableState.page,
    pageSize: data?.pageSize ?? tableState.pageSize,
    totalPages: data?.totalPages ?? 0,

    // Loading states
    isLoading: isProjectLoading || isTracesLoading,
    isFetching: isTracesFetching,
    isProjectLoading,
    isTracesLoading,

    // Error state
    error: error instanceof Error ? error.message : error ? String(error) : null,

    // Actions
    refetch,

    // Project context
    hasProject,
    currentProject,

    // Table state (for components to use)
    tableState,
  }
}

/**
 * Return type for useProjectTraces hook
 */
export interface UseProjectTracesReturn {
  // Data
  data: Trace[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number

  // Loading states
  isLoading: boolean
  isFetching: boolean
  isProjectLoading: boolean
  isTracesLoading: boolean

  // Error state
  error: string | null

  // Actions
  refetch: () => void

  // Project context
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']

  // Table state
  tableState: UseTracesTableStateReturn
}

/**
 * Hook to fetch filter options for traces
 *
 * Returns available values for filter dropdowns (models, providers, services, etc.)
 * and min/max ranges for numeric filters (cost, tokens, duration).
 *
 * Used to populate the advanced filter UI dynamically based on actual data.
 *
 * @returns Filter options data, loading state, and error state
 */
export function useTraceFilterOptions() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  const projectId = currentProject?.id

  const {
    data,
    isLoading: isOptionsLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['traceFilterOptions', projectId],

    queryFn: async () => {
      if (!projectId) {
        throw new Error('No project selected')
      }

      return getTraceFilterOptions(projectId)
    },

    enabled: !!projectId && hasProject,

    // Longer cache since filter options don't change frequently
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 15 * 60 * 1000, // 15 minutes
    refetchOnWindowFocus: false, // Don't refetch on focus since options are stable

    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  return {
    // Data - provide sensible defaults
    filterOptions: data ?? {
      models: [],
      providers: [],
      services: [],
      environments: [],
      users: [],
      sessions: [],
      costRange: null,
      tokenRange: null,
      durationRange: null,
    },

    // Loading states
    isLoading: isProjectLoading || isOptionsLoading,
    isProjectLoading,
    isOptionsLoading,

    // Error state
    error: error instanceof Error ? error.message : error ? String(error) : null,

    // Actions
    refetch,

    // Project context
    hasProject,
    currentProject,
  }
}

/**
 * Return type for useTraceFilterOptions hook
 */
export interface UseTraceFilterOptionsReturn {
  filterOptions: TraceFilterOptions
  isLoading: boolean
  isProjectLoading: boolean
  isOptionsLoading: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
}
