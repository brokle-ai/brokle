'use client'

import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useProjectOnly } from '@/features/projects'
import { datasetsApi } from '../api/datasets-api'
import { datasetQueryKeys } from './use-datasets'

export interface UseProjectDatasetsReturn {
  data: Awaited<ReturnType<typeof datasetsApi.listDatasets>>
  totalCount: number
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

  const { filter } = useTableSearchParams(searchParams)
  const projectId = currentProject?.id

  const {
    data,
    isLoading: isDatasetsLoading,
    isFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: [...datasetQueryKeys.list(projectId ?? ''), filter],
    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')
      return datasetsApi.listDatasets(projectId)
    },
    enabled: !!projectId && hasProject,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    retry: 2,
    placeholderData: keepPreviousData,
  })

  // Client-side filtering based on search term
  const filteredData = data?.filter((dataset) => {
    if (!filter) return true
    const searchLower = filter.toLowerCase()
    return (
      dataset.name.toLowerCase().includes(searchLower) ||
      dataset.description?.toLowerCase().includes(searchLower)
    )
  }) ?? []

  return {
    data: filteredData,
    totalCount: filteredData.length,
    isLoading: isProjectLoading || isDatasetsLoading,
    isFetching,
    error: error instanceof Error ? error.message : error ? String(error) : null,
    refetch,
    hasProject,
    currentProject,
  }
}
