'use client'

import { useQuery } from '@tanstack/react-query'
import { getScoresForTrace } from '../api/traces-api'
import type { Score } from '../data/schema'

/**
 * Query keys for trace scores
 */
export const traceScoresQueryKeys = {
  all: ['trace-scores'] as const,
  forTrace: (projectId: string, traceId: string) =>
    [...traceScoresQueryKeys.all, projectId, traceId] as const,
}

/**
 * Hook to fetch scores for a trace
 *
 * Uses React Query for:
 * - Automatic caching
 * - Loading state management
 * - Error handling
 * - Background refetching
 *
 * @param projectId - Project ULID (optional, query disabled if not provided)
 * @param traceId - Trace ID (optional, query disabled if not provided)
 * @returns Query result with scores array
 */
export function useTraceScoresQuery(
  projectId: string | undefined,
  traceId: string | undefined
) {
  return useQuery({
    queryKey: traceScoresQueryKeys.forTrace(projectId ?? '', traceId ?? ''),
    queryFn: () => getScoresForTrace(projectId!, traceId!),
    enabled: !!projectId && !!traceId,
    staleTime: 30_000, // 30 seconds - scores don't change frequently
    gcTime: 5 * 60 * 1000, // 5 minutes garbage collection
  })
}
