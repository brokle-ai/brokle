'use client'

import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useProjectOnly } from '@/features/projects'
import { datasetsApi } from '../api/datasets-api'
import { datasetQueryKeys } from './use-datasets'

import type { Dataset } from '../types'

export interface UseProjectDatasetsReturn {
  data: Dataset[]
  totalCount: number
  page: number
  pageSize: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
}

export function useProjectDatasets(): UseProjectDatasetsReturn {
  const searchParams = useSearchParams()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  const { filter, page, pageSize } = useTableSearchParams(searchParams)
  const projectId = currentProject?.id

  const {
    data,
    isLoading: isDatasetsLoading,
    isFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: [...datasetQueryKeys.list(projectId ?? ''), filter, page, pageSize],
    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')
      return datasetsApi.listDatasets(projectId, {
        page,
        limit: pageSize,
        search: filter || undefined,
      })
    },
    enabled: !!projectId && hasProject,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    retry: 2,
    placeholderData: keepPreviousData,
  })

  return {
    data: data?.data ?? [],
    totalCount: data?.pagination?.total ?? 0,
    page,
    pageSize,
    isLoading: isProjectLoading || isDatasetsLoading,
    isFetching,
    error: error instanceof Error ? error.message : error ? String(error) : null,
    refetch,
    hasProject,
    currentProject,
  }
}
