'use client'

import { useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useExperimentsQuery } from './use-experiments'
import type { Experiment } from '../types'

export interface UseProjectExperimentsReturn {
  data: Experiment[]
  totalCount: number
  page: number
  pageSize: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
}

export function useProjectExperiments(): UseProjectExperimentsReturn {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const { filter: searchFilter, page, pageSize } = useTableSearchParams(searchParams)

  const projectId = currentProject?.id

  const {
    data: experiments,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useExperimentsQuery(projectId, {
    page,
    limit: pageSize,
    search: searchFilter || undefined,
  })

  return {
    data: experiments?.data ?? [],
    totalCount: experiments?.pagination?.total ?? 0,
    page,
    pageSize,
    isLoading: isProjectLoading || isQueryLoading,
    isFetching,
    error: error?.message ?? null,
    refetch,
    hasProject,
  }
}
