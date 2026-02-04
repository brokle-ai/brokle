'use client'

import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useProjectOnly } from '@/features/projects'
import { getPrompts } from '../api/prompts-api'
import { promptQueryKeys } from './use-prompts-queries'
import { usePromptsTableState } from './use-prompts-table-state'
import type { PromptListItem } from '../types'

/**
 * Hook to fetch and manage project prompts with filtering and pagination
 *
 * Uses React Query for:
 * - Automatic caching (30 seconds stale time)
 * - Loading state management
 * - Error handling
 * - Background refetching
 *
 * Requires:
 * - Project context (from workspace context)
 * - URL state from nuqs (page, filters)
 *
 * @returns Prompts data, pagination, loading state, error state, and table state
 */
export function useProjectPrompts() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = usePromptsTableState()

  // Extract values from table state
  const { page, pageSize, search: searchFilter, types } = tableState

  // Convert to API format
  const search = searchFilter || undefined
  // API currently accepts single type, use first selected type
  const type = types.length > 0 ? types[0] : undefined
  const limit = pageSize

  const projectId = currentProject?.id

  const filters = { page, limit, search, type, tags: undefined }

  const {
    data,
    isLoading: isPromptsLoading,
    isFetching: isPromptsFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: promptQueryKeys.list(projectId || '', filters),

    queryFn: async () => {
      if (!projectId) {
        throw new Error('No project selected')
      }

      return getPrompts({
        projectId,
        ...filters,
      })
    },

    enabled: !!projectId && hasProject,

    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

    // Keep previous data visible while fetching new data (e.g., when filters change)
    placeholderData: keepPreviousData,
  })

  return {
    // Data
    data: data?.data ?? [],
    totalCount: data?.pagination?.total ?? 0,
    page: data?.pagination?.page ?? page,
    pageSize: data?.pagination?.limit ?? limit,
    totalPages: data?.pagination?.totalPages ?? 0,

    // Loading states
    isLoading: isProjectLoading || isPromptsLoading,
    isFetching: isPromptsFetching,
    isProjectLoading,
    isPromptsLoading,

    // Error state
    error: error instanceof Error ? error.message : error ? String(error) : null,

    // Actions
    refetch,

    // Project context
    hasProject,
    currentProject,

    // Table state (for URL-based state management)
    tableState,
  }
}

/**
 * Return type for useProjectPrompts hook
 */
export interface UseProjectPromptsReturn {
  // Data
  data: PromptListItem[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number

  // Loading states
  isLoading: boolean
  isFetching: boolean
  isProjectLoading: boolean
  isPromptsLoading: boolean

  // Error state
  error: string | null

  // Actions
  refetch: () => void

  // Project context
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']

  // Table state (for URL-based state management)
  tableState: ReturnType<typeof usePromptsTableState>
}
