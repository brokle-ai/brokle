import type { GetTracesParams, GetSpansParams } from '../api/traces-api'

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
 * - ['spans', 'list', projectId, params?] - flat spans list queries
 * - ['sessions', 'list', projectId, params?] - sessions list queries
 * - ['metrics', projectId] - trace metrics queries
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

  // Spans for a trace - ['traces', 'detail', projectId, traceId, 'spans']
  traceSpans: (projectId: string, traceId: string) =>
    [...traceQueryKeys.detail(projectId, traceId), 'spans'] as const,

  // Flat spans list - ['spans', 'list', projectId, params?]
  spansBase: () => ['spans'] as const,
  spansList: () => [...traceQueryKeys.spansBase(), 'list'] as const,
  spans: (projectId: string, params?: Omit<GetSpansParams, 'projectId'>) =>
    params
      ? ([...traceQueryKeys.spansList(), projectId, params] as const)
      : ([...traceQueryKeys.spansList(), projectId] as const),

  // Sessions list - ['sessions', 'list', projectId, params?]
  sessionsBase: () => ['sessions'] as const,
  sessionsList: () => [...traceQueryKeys.sessionsBase(), 'list'] as const,
  sessions: (projectId: string, params?: Record<string, any>) =>
    params
      ? ([...traceQueryKeys.sessionsList(), projectId, params] as const)
      : ([...traceQueryKeys.sessionsList(), projectId] as const),

  // Trace metrics - ['traces', 'metrics', projectId, params?]
  metricsBase: () => [...traceQueryKeys.all, 'metrics'] as const,
  metrics: (projectId: string, params?: Record<string, any>) =>
    params
      ? ([...traceQueryKeys.metricsBase(), projectId, params] as const)
      : ([...traceQueryKeys.metricsBase(), projectId] as const),

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
