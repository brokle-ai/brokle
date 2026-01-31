'use client'

import { useMemo } from 'react'
import { useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useEvaluatorsQuery } from './use-evaluators'
import type { Evaluator } from '../types'

export interface UseProjectEvaluatorsReturn {
  data: Evaluator[]
  totalCount: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
}

export function useProjectEvaluators(): UseProjectEvaluatorsReturn {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { filter: searchFilter } = useTableSearchParams(searchParams)

  const projectId = currentProject?.id

  const {
    data: evaluatorsResponse,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useEvaluatorsQuery(projectId)

  // Client-side filtering since the API doesn't support filter param
  const evaluators = evaluatorsResponse?.data
  const filteredData = useMemo(() => {
    if (!evaluators) return []
    if (!searchFilter) return evaluators

    const lowerFilter = searchFilter.toLowerCase()
    return evaluators.filter(
      (evaluator: Evaluator) =>
        evaluator.name.toLowerCase().includes(lowerFilter) ||
        evaluator.description?.toLowerCase().includes(lowerFilter)
    )
  }, [evaluators, searchFilter])

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
