'use client'

import { useQuery } from '@tanstack/react-query'
import { scoresApi } from '../api/scores-api'
import type { ScoreAnalyticsParams } from '../types'

export const scoreAnalyticsQueryKeys = {
  all: ['score-analytics'] as const,
  analytics: (projectId: string, params: ScoreAnalyticsParams) =>
    [...scoreAnalyticsQueryKeys.all, 'data', projectId, params] as const,
  names: (projectId: string) =>
    [...scoreAnalyticsQueryKeys.all, 'names', projectId] as const,
}

export function useScoreAnalyticsQuery(
  projectId: string | undefined,
  params: ScoreAnalyticsParams | undefined
) {
  return useQuery({
    queryKey: scoreAnalyticsQueryKeys.analytics(projectId ?? '', params ?? { score_name: '' }),
    queryFn: () => scoresApi.getScoreAnalytics(projectId!, params!),
    enabled: !!projectId && !!params?.score_name,
    staleTime: 60_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useScoreNamesQuery(projectId: string | undefined) {
  return useQuery({
    queryKey: scoreAnalyticsQueryKeys.names(projectId ?? ''),
    queryFn: () => scoresApi.getScoreNames(projectId!),
    enabled: !!projectId,
    staleTime: 60_000,
    gcTime: 5 * 60 * 1000,
  })
}
