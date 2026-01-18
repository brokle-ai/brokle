'use client'

import { useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import { useProjectOnly } from '@/features/projects'
import { traceQueryKeys } from './trace-query-keys'
import { getSpans } from '../api/traces-api'
import type { Span } from '../data/schema'
import type { GetSpansParams } from '../api/traces-api'

/**
 * Maps frontend column IDs to backend sort field names for spans.
 */
const SPAN_SORT_FIELD_MAP: Record<string, string> = {
  span_name: 'span_name',
  start_time: 'start_time',
  duration: 'duration',
  model_name: 'model_name',
  total_cost: 'total_cost',
}

export interface UseSpansTableStateReturn {
  // State (read from URL)
  page: number
  pageSize: number
  spanType: string | null
  model: string | null
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null

  // Setters
  setSpanType: (type: string | null) => void
  setModel: (model: string | null) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  // Computed
  hasActiveFilters: boolean
}

/**
 * Hook to manage spans table URL state.
 * Uses nuqs for type-safe URL synchronization.
 */
export function useSpansTableState(): UseSpansTableStateReturn {
  const [query, setQuery] = useQueryStates({
    spansPage: parseAsInteger.withDefault(1),
    spansPageSize: parseAsInteger.withDefault(20),
    spanType: parseAsString,
    spanModel: parseAsString,
    spansSortBy: parseAsString,
    spansSortOrder: parseAsString,
  })

  const setSpanType = useCallback(
    (type: string | null) => {
      setQuery({ spanType: type, spansPage: 1 })
    },
    [setQuery]
  )

  const setModel = useCallback(
    (model: string | null) => {
      setQuery({ spanModel: model, spansPage: 1 })
    },
    [setQuery]
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        spansPage: Math.max(1, page),
        ...(pageSize !== undefined && { spansPageSize: Math.max(1, pageSize) }),
      })
    },
    [setQuery]
  )

  const setSorting = useCallback(
    (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => {
      setQuery({
        spansSortBy: sortBy || null,
        spansSortOrder: sortOrder || null,
      })
    },
    [setQuery]
  )

  const resetAll = useCallback(() => {
    setQuery({
      spansPage: 1,
      spansPageSize: null,
      spanType: null,
      spanModel: null,
      spansSortBy: null,
      spansSortOrder: null,
    })
  }, [setQuery])

  return {
    page: Math.max(1, query.spansPage),
    pageSize: Math.max(1, query.spansPageSize),
    spanType: query.spanType,
    model: query.spanModel,
    sortBy: query.spansSortBy,
    sortOrder: query.spansSortOrder as 'asc' | 'desc' | null,

    setSpanType,
    setModel,
    setPagination,
    setSorting,
    resetAll,

    hasActiveFilters: !!query.spanType || !!query.spanModel,
  }
}

/**
 * Hook to fetch and manage project spans with filtering, sorting, and pagination.
 *
 * @returns Spans data, pagination, loading state, error state, and table state
 */
export function useProjectSpans() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = useSpansTableState()
  const projectId = currentProject?.id

  // Map filter conditions to API params
  const apiParams = useMemo((): GetSpansParams | null => {
    if (!projectId) return null

    const params: GetSpansParams = {
      projectId,
      page: tableState.page,
      pageSize: tableState.pageSize,
      type: tableState.spanType || undefined,
      model: tableState.model || undefined,
      sortBy: tableState.sortBy ? SPAN_SORT_FIELD_MAP[tableState.sortBy] : undefined,
      sortOrder: tableState.sortOrder || undefined,
    }

    return params
  }, [
    projectId,
    tableState.page,
    tableState.pageSize,
    tableState.spanType,
    tableState.model,
    tableState.sortBy,
    tableState.sortOrder,
  ])

  // Extract params without projectId for query key
  const queryParams = useMemo(() => {
    if (!apiParams) return undefined
    const { projectId: _, ...rest } = apiParams
    return rest
  }, [apiParams])

  const {
    data,
    isLoading: isSpansLoading,
    isFetching: isSpansFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: traceQueryKeys.spans(projectId!, queryParams),

    queryFn: () => {
      if (!apiParams) {
        throw new Error('No project selected')
      }
      return getSpans(apiParams)
    },

    enabled: !!apiParams && hasProject,

    // Cache configuration - 30s auto-refetch like Opik
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    refetchInterval: 30_000, // Auto-refetch every 30s

    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  return {
    // Data
    data: data?.spans ?? [],
    totalCount: data?.totalCount ?? 0,
    page: data?.page ?? tableState.page,
    pageSize: data?.pageSize ?? tableState.pageSize,
    totalPages: data?.totalPages ?? 0,

    // Loading states
    isLoading: isProjectLoading || isSpansLoading,
    isFetching: isSpansFetching,
    isProjectLoading,
    isSpansLoading,

    // Error state
    error: error instanceof Error ? error.message : error ? String(error) : null,

    // Actions
    refetch,

    // Project context
    hasProject,
    currentProject,

    // Table state
    tableState,
  }
}

export interface UseProjectSpansReturn {
  data: Span[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  isLoading: boolean
  isFetching: boolean
  isProjectLoading: boolean
  isSpansLoading: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
  tableState: UseSpansTableStateReturn
}
