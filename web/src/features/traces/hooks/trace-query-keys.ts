import type { GetTracesParams } from '../api/traces-api'

/**
 * Query key factory for trace-related queries
 *
 * Follows the hierarchical pattern used by other features (budgetQueryKeys, promptQueryKeys)
 * to enable prefix-based cache invalidation.
 *
 * Key structure:
 * - ['traces'] - base prefix for all trace queries
 * - ['traces', 'list', projectId, params?] - list queries
 * - ['traces', 'detail', projectId, traceId] - detail queries
 * - ['traces', 'detail', projectId, traceId, 'spans'] - spans for a trace
 * - ['traces', 'filterOptions', projectId] - filter options
 *
 * This enables:
 * - Invalidating all lists: queryClient.invalidateQueries({ queryKey: traceQueryKeys.list(projectId) })
 * - Invalidating specific detail: queryClient.invalidateQueries({ queryKey: traceQueryKeys.detail(projectId, traceId) })
 * - Invalidating everything: queryClient.invalidateQueries({ queryKey: traceQueryKeys.all })
 */
export const traceQueryKeys = {
  // Base key for all trace queries
  all: ['traces'] as const,

  // List queries - ['traces', 'list', projectId, params?]
  lists: () => [...traceQueryKeys.all, 'list'] as const,
  list: (projectId: string, params?: Omit<GetTracesParams, 'projectId'>) =>
    params
      ? ([...traceQueryKeys.lists(), projectId, params] as const)
      : ([...traceQueryKeys.lists(), projectId] as const),

  // Detail queries - ['traces', 'detail', projectId, traceId]
  details: () => [...traceQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, traceId: string) =>
    [...traceQueryKeys.details(), projectId, traceId] as const,

  // Spans - ['traces', 'detail', projectId, traceId, 'spans']
  spans: (projectId: string, traceId: string) =>
    [...traceQueryKeys.detail(projectId, traceId), 'spans'] as const,

  // Filter options - ['traces', 'filterOptions', projectId]
  filterOptions: (projectId: string) =>
    [...traceQueryKeys.all, 'filterOptions', projectId] as const,

  // Scores - ['traces', 'detail', projectId, traceId, 'scores']
  scores: (projectId: string, traceId: string) =>
    [...traceQueryKeys.detail(projectId, traceId), 'scores'] as const,

  // Comments - ['traces', 'detail', projectId, traceId, 'comments']
  comments: (projectId: string, traceId: string) =>
    [...traceQueryKeys.detail(projectId, traceId), 'comments'] as const,

  // Comment count - ['traces', 'detail', projectId, traceId, 'comments', 'count']
  commentCount: (projectId: string, traceId: string) =>
    [...traceQueryKeys.comments(projectId, traceId), 'count'] as const,
}
