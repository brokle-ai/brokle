'use client'

import { useQuery } from '@tanstack/react-query'
import { evaluatorsApi } from '../api/evaluators-api'
import type { EvaluatorAnalyticsParams } from '../types'

export const evaluatorAnalyticsKeys = {
  all: ['evaluator-analytics'] as const,
  analytics: (projectId: string, evaluatorId: string, params?: EvaluatorAnalyticsParams) =>
    [...evaluatorAnalyticsKeys.all, 'data', projectId, evaluatorId, params] as const,
}

/**
 * Query hook for fetching evaluator analytics with time period filtering
 */
export function useEvaluatorAnalyticsQuery(
  projectId: string | undefined,
  evaluatorId: string | undefined,
  params?: EvaluatorAnalyticsParams
) {
  return useQuery({
    queryKey: evaluatorAnalyticsKeys.analytics(projectId ?? '', evaluatorId ?? '', params),
    queryFn: () => evaluatorsApi.getEvaluatorAnalytics(projectId!, evaluatorId!, params),
    enabled: !!projectId && !!evaluatorId,
    staleTime: 60_000, // 1 minute
    gcTime: 5 * 60 * 1000, // 5 minutes
  })
}
