'use client'

import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { experimentsApi } from '../api/experiments-api'
import { experimentQueryKeys } from './use-experiments'
import type {
  ExperimentComparisonResponse,
  ExperimentScoreStats,
  ExperimentScoreDiff,
  ScoreComparisonRow,
  ExperimentComparisonSummary,
} from '../types'

/**
 * Calculate diff between baseline and current experiment
 */
function calculateDiff(
  baseline: ExperimentScoreStats | undefined,
  current: ExperimentScoreStats | undefined
): ExperimentScoreDiff | undefined {
  if (!baseline || !current) return undefined

  const difference = current.mean - baseline.mean
  return {
    type: 'NUMERIC',
    difference: Math.abs(difference),
    direction: difference >= 0 ? '+' : '-',
  }
}

/**
 * Transform API response to score rows with client-side diff calculation
 */
function transformToScoreRows(
  data: ExperimentComparisonResponse,
  baselineId: string | undefined,
  experimentIds: string[]
): ScoreComparisonRow[] {
  const { scores } = data
  const scoreNames = Object.keys(scores)

  return scoreNames.map((scoreName) => {
    const scoreData = scores[scoreName]
    const baselineStats = baselineId ? scoreData[baselineId] : undefined

    const experiments: ScoreComparisonRow['experiments'] = {}

    experimentIds.forEach((expId) => {
      const stats = scoreData[expId]
      if (stats) {
        experiments[expId] = {
          stats,
          diff:
            expId !== baselineId
              ? calculateDiff(baselineStats, stats)
              : undefined,
        }
      }
    })

    return { scoreName, experiments }
  })
}

export function useExperimentComparisonQuery(
  projectId: string | undefined,
  experimentIds: string[],
  baselineId?: string
) {
  const query = useQuery({
    queryKey: [
      ...experimentQueryKeys.all,
      'compare',
      projectId,
      [...experimentIds].sort().join(','),
      baselineId,
    ],
    queryFn: () =>
      experimentsApi.compareExperiments(projectId!, experimentIds, baselineId),
    enabled: !!projectId && experimentIds.length >= 2,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })

  // Allows re-calculating diffs when baseline changes without new API call
  const scoreRows = useMemo(() => {
    if (!query.data) return []
    return transformToScoreRows(query.data, baselineId, experimentIds)
  }, [query.data, baselineId, experimentIds])

  const experiments: Record<string, ExperimentComparisonSummary> = useMemo(
    () => query.data?.experiments ?? {},
    [query.data?.experiments]
  )

  return {
    ...query,
    scoreRows,
    experiments,
  }
}
