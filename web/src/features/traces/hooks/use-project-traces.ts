'use client'

import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useProjectOnly } from '@/features/projects'
import { getProjectTraces } from '../api/traces-api'
import type { Trace } from '../data/schema'

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
