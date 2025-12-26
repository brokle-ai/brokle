'use client'

import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useProjectOnly } from '@/features/projects'
import { getProjectTraces, getTraceFilterOptions } from '../api/traces-api'
import type { Trace } from '../data/schema'
import type { TraceFilterOptions } from '../api/traces-api'

/**
 * Hook to fetch and manage project traces with filtering, sorting, and pagination
 *
 * Uses React Query for:
 * - Automatic caching (30 seconds stale time)
 * - Loading state management
 * - Error handling
 * - Background refetching
 *
 * Requires:
 * - Project context (from workspace context)
 * - Search params for table state (page, filters, sorting)
 *
 * @returns Traces data, pagination, loading state, and error state
 */
export function useProjectTraces() {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  const { page, pageSize, filter, status, sortBy, sortOrder } =
    useTableSearchParams(searchParams)

  // Extract project ID from current project
  const projectId = currentProject?.id

  // Use React Query for data fetching with automatic caching
  const {
    data,
    isLoading: isTracesLoading,
    isFetching: isTracesFetching,
    error,
    refetch,
  } = useQuery({
    // Query key includes all parameters that affect the data
    queryKey: ['traces', projectId, page, pageSize, filter, status, sortBy, sortOrder],

    // Query function: fetch traces from backend
    queryFn: async () => {
      if (!projectId) {
        throw new Error('No project selected')
      }

      return getProjectTraces({
        projectId,
        page,
        pageSize,
        search: filter || undefined,
        status: status.length > 0 ? status : undefined,
        sortBy: sortBy || undefined,
        sortOrder: sortOrder || undefined,
      })
    },

    // Only fetch when we have a project ID
    enabled: !!projectId && hasProject,

    // Cache configuration
    staleTime: 30_000, // 30 seconds - data is fresh for this duration
    gcTime: 5 * 60 * 1000, // 5 minutes - keep unused data in cache
    refetchOnWindowFocus: true, // Refetch when user returns to tab
    refetchOnReconnect: true, // Refetch when internet reconnects

    // Retry configuration
    retry: 2, // Retry failed requests twice
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000), // Exponential backoff

    // Keep previous data visible while fetching new data (e.g., when filters change)
    placeholderData: keepPreviousData,
  })

  return {
    // Data
    data: data?.traces ?? [],
    totalCount: data?.totalCount ?? 0,
    page: data?.page ?? page,
    pageSize: data?.pageSize ?? pageSize,
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
