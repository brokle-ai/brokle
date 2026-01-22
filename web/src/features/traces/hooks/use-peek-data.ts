'use client'

import { useSearchParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { useProjectOnly } from '@/features/projects'
import { traceQueryKeys } from './trace-query-keys'
import { getTraceById } from '../api/traces-api'
import type { Trace } from '../data/schema'

/**
 * Hook to fetch trace details for peek view
 * Only fetches when peek parameter exists in URL
 *
 * Uses React Query for:
 * - Automatic caching (shared with detail page)
 * - Loading state management
 * - Error handling
 * - Background refetching
 */
export function usePeekData() {
  const searchParams = useSearchParams()
  const peekId = searchParams.get('peek')
  const { currentProject, hasProject } = useProjectOnly()

  // Extract project ID from current project
  const projectId = currentProject?.id

  // Fetch trace data with React Query
  const {
    data: trace,
    isLoading,
    error,
  } = useQuery({
    // Query key matches detail page for cache sharing
    queryKey: traceQueryKeys.detail(projectId!, peekId!),

    // Query function: fetch trace from backend
    queryFn: async () => {
      if (!projectId) {
        throw new Error('No project selected')
      }
      if (!peekId) {
        throw new Error('No trace ID provided')
      }
      return getTraceById(projectId, peekId)
    },

    // Only fetch when we have both project ID and peek ID
    enabled: !!projectId && !!peekId && hasProject,

    // Retry configuration
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

    // Cache shares with detail page - if user opens peek then navigates to detail,
    // the data is already cached!
    staleTime: 30_000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes
  })

  return {
    trace: trace ?? null,
    isLoading,
    error: error instanceof Error ? error : null,
    peekId,
    projectId: projectId ?? null,
  }
}
