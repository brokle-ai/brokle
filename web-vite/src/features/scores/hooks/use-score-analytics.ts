import { useQuery } from '@tanstack/react-query'
import {
  scoreAnalyticsQueryOptions,
  scoreNamesQueryOptions,
} from '../api/queries'
import type { ScoreAnalyticsParams } from '../api/types'

/**
 * Score analytics data for the dashboard. Disabled until both a
 * project id and a primary `score_name` are available — without the
 * latter the backend has nothing to aggregate.
 */
export function useScoreAnalyticsQuery(
  projectId: string | undefined,
  params: ScoreAnalyticsParams | undefined,
) {
  return useQuery({
    ...scoreAnalyticsQueryOptions(
      projectId ?? '',
      params ?? { score_name: '' },
    ),
    enabled: !!projectId && !!params?.score_name,
  })
}

/**
 * Distinct score names recorded in this project. Drives the score
 * selector on the analytics page.
 */
export function useScoreNamesQuery(projectId: string | undefined) {
  return useQuery({
    ...scoreNamesQueryOptions(projectId ?? ''),
    enabled: !!projectId,
  })
}
