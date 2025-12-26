'use client'

import { useMemo } from 'react'
import { useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useEvaluationRulesQuery } from './use-evaluation-rules'
import type { EvaluationRule } from '../types'

export interface UseProjectRulesReturn {
  data: EvaluationRule[]
  totalCount: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
}

export function useProjectRules(): UseProjectRulesReturn {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { filter: searchFilter } = useTableSearchParams(searchParams)

  const projectId = currentProject?.id

  const {
    data: rulesResponse,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useEvaluationRulesQuery(projectId)

  // Client-side filtering since the API doesn't support filter param
  const rules = rulesResponse?.rules
  const filteredData = useMemo(() => {
    if (!rules) return []
    if (!searchFilter) return rules

    const lowerFilter = searchFilter.toLowerCase()
    return rules.filter(
      (rule) =>
        rule.name.toLowerCase().includes(lowerFilter) ||
        rule.description?.toLowerCase().includes(lowerFilter)
    )
  }, [rules, searchFilter])

  return {
    data: filteredData,
    totalCount: filteredData.length,
    isLoading: isProjectLoading || isQueryLoading,
    isFetching,
    error: error?.message ?? null,
    refetch,
    hasProject,
  }
}
