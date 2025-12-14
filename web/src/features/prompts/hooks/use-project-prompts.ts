'use client'

import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { getPrompts } from '../api/prompts-api'
import { promptQueryKeys } from './use-prompts-queries'
import type { PromptListItem, PromptType } from '../types'

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
 * - Search params for table state (page, filters)
 *
 * @returns Prompts data, pagination, loading state, and error state
 */
export function useProjectPrompts() {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  // Parse search params
  const page = parseInt(searchParams.get('page') || '1', 10)
  const limit = parseInt(searchParams.get('limit') || '50', 10)
  const search = searchParams.get('search') || undefined
  const type = (searchParams.get('type') as PromptType) || undefined
  const tagsParam = searchParams.get('tags')
  const tags = tagsParam ? tagsParam.split(',').filter(Boolean) : undefined

  const projectId = currentProject?.id

  const filters = { page, limit, search, type, tags }

  const {
    data,
    isLoading: isPromptsLoading,
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
  })

  return {
    // Data
    data: data?.prompts ?? [],
    totalCount: data?.totalCount ?? 0,
    page: data?.page ?? page,
    pageSize: data?.pageSize ?? limit,
    totalPages: data?.totalPages ?? 0,

    // Loading states
    isLoading: isProjectLoading || isPromptsLoading,
    isProjectLoading,
    isPromptsLoading,

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
  isProjectLoading: boolean
  isPromptsLoading: boolean

  // Error state
  error: string | null

  // Actions
  refetch: () => void

  // Project context
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
}
