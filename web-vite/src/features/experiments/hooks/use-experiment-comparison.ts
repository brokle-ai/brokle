import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import {
  compareExperiments,
  experimentCompareKey,
} from '../api/queries'
import type {
  ExperimentComparisonResponse,
  ExperimentComparisonSummary,
  ExperimentScoreDiff,
  ExperimentScoreStats,
  ScoreComparisonRow,
} from '../api/types'

/**
 * Calculate diff between baseline and current experiment client-side so
 * baseline switching doesn't require a server round-trip.
 */
function calculateDiff(
  baseline: ExperimentScoreStats | undefined,
  current: ExperimentScoreStats | undefined,
): ExperimentScoreDiff | undefined {
  if (!baseline || !current) return undefined
  const difference = current.mean - baseline.mean
  return {
    type: 'NUMERIC',
    difference: Math.abs(difference),
    direction: difference >= 0 ? '+' : '-',
  }
}

function transformToScoreRows(
  data: ExperimentComparisonResponse,
  baselineId: string | undefined,
  experimentIds: string[],
): ScoreComparisonRow[] {
  const { scores } = data
  return Object.keys(scores).map((scoreName) => {
    const scoreData = scores[scoreName] ?? {}
    const baselineStats = baselineId ? scoreData[baselineId] : undefined

    const experiments: ScoreComparisonRow['experiments'] = {}
    for (const expId of experimentIds) {
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
    }
    return { scoreName, experiments }
  })
}

export function useExperimentComparisonQuery(
  projectId: string | undefined,
  experimentIds: string[],
  baselineId?: string,
) {
  const query = useQuery({
    queryKey: experimentCompareKey(projectId ?? '', experimentIds, baselineId),
    queryFn: () =>
      compareExperiments(projectId!, experimentIds, baselineId),
    enabled: !!projectId && experimentIds.length >= 2,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })

  // Recompute diffs when baseline changes without hitting the network.
  const scoreRows = useMemo(() => {
    if (!query.data) return []
    return transformToScoreRows(query.data, baselineId, experimentIds)
  }, [query.data, baselineId, experimentIds])

  const experiments: Record<string, ExperimentComparisonSummary> = useMemo(
    () => query.data?.experiments ?? {},
    [query.data?.experiments],
  )

  return {
    ...query,
    scoreRows,
    experiments,
  }
}
