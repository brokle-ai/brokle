'use client'

import { useQuery } from '@tanstack/react-query'
import { evaluationRulesApi } from '../api/evaluation-rules-api'
import type { RuleAnalyticsParams } from '../types'

export const ruleAnalyticsKeys = {
  all: ['rule-analytics'] as const,
  analytics: (projectId: string, ruleId: string, params?: RuleAnalyticsParams) =>
    [...ruleAnalyticsKeys.all, 'data', projectId, ruleId, params] as const,
}

/**
 * Query hook for fetching rule analytics with time period filtering
 */
export function useRuleAnalyticsQuery(
  projectId: string | undefined,
  ruleId: string | undefined,
  params?: RuleAnalyticsParams
) {
  return useQuery({
    queryKey: ruleAnalyticsKeys.analytics(projectId ?? '', ruleId ?? '', params),
    queryFn: () => evaluationRulesApi.getRuleAnalytics(projectId!, ruleId!, params),
    enabled: !!projectId && !!ruleId,
    staleTime: 60_000, // 1 minute
    gcTime: 5 * 60 * 1000, // 5 minutes
  })
}
