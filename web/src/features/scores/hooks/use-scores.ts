'use client'

import { useQuery } from '@tanstack/react-query'
import { scoresApi } from '../api/scores-api'
import type { ScoreListParams } from '../types'

export const scoreQueryKeys = {
  all: ['scores'] as const,
  list: (projectId: string, params?: ScoreListParams) =>
    [...scoreQueryKeys.all, 'list', projectId, params] as const,
  detail: (projectId: string, scoreId: string) =>
    [...scoreQueryKeys.all, 'detail', projectId, scoreId] as const,
}

export function useScoresQuery(
  projectId: string | undefined,
  params?: ScoreListParams
) {
  return useQuery({
    queryKey: scoreQueryKeys.list(projectId ?? '', params),
    queryFn: () => scoresApi.listScores(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useScoreQuery(
  projectId: string | undefined,
  scoreId: string | undefined
) {
  return useQuery({
    queryKey: scoreQueryKeys.detail(projectId ?? '', scoreId ?? ''),
    queryFn: () => scoresApi.getScore(projectId!, scoreId!),
    enabled: !!projectId && !!scoreId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}
