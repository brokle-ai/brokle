'use client'

import { useMemo } from 'react'
import { useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useExperimentsQuery } from './use-experiments'
import type { Experiment } from '../types'

export interface UseProjectExperimentsReturn {
  data: Experiment[]
  totalCount: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
}

export function useProjectExperiments(): UseProjectExperimentsReturn {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { filter: searchFilter } = useTableSearchParams(searchParams)

  const projectId = currentProject?.id

  const {
    data: experiments,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useExperimentsQuery(projectId)

  // Client-side filtering since the API doesn't support filter param
  const filteredData = useMemo(() => {
    if (!experiments) return []
    if (!searchFilter) return experiments

    const lowerFilter = searchFilter.toLowerCase()
    return experiments.filter(
      (experiment) =>
        experiment.name.toLowerCase().includes(lowerFilter) ||
        experiment.description?.toLowerCase().includes(lowerFilter)
    )
  }, [experiments, searchFilter])

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
