'use client'

import { useMemo } from 'react'
import { useProjectOnly } from '@/features/projects'
import { useExperimentsQuery } from './use-experiments'
import { useExperimentsTableState } from './use-experiments-table-state'
import type { Experiment } from '../types'

export interface UseProjectExperimentsReturn {
  data: Experiment[]
  totalCount: number
  isLoading: boolean
  isFetching: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  projectSlug: string | undefined
}

export function useProjectExperiments(): UseProjectExperimentsReturn {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = useExperimentsTableState()

  const projectId = currentProject?.id
  const projectSlug = currentProject?.slug

  const {
    data: experiments,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useExperimentsQuery(projectId)

  // Destructure for stable references
  const { search, status, sortBy, sortOrder, page, pageSize } = tableState

  // Client-side filtering since the API doesn't support filter params yet
  const filteredData = useMemo(() => {
    if (!experiments) return []

    let filtered = experiments

    // Filter by search
    if (search) {
      const lowerFilter = search.toLowerCase()
      filtered = filtered.filter(
        (experiment) =>
          experiment.name.toLowerCase().includes(lowerFilter) ||
          experiment.description?.toLowerCase().includes(lowerFilter)
      )
    }

    // Filter by status
    if (status) {
      filtered = filtered.filter(
        (experiment) => experiment.status === status
      )
    }

    // Sort
    filtered = [...filtered].sort((a, b) => {
      let comparison = 0

      switch (sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'status':
          comparison = a.status.localeCompare(b.status)
          break
        case 'created_at':
          comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
          break
        case 'updated_at':
          comparison = new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime()
          break
        default:
          comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      }

      return sortOrder === 'desc' ? -comparison : comparison
    })

    return filtered
  }, [experiments, search, status, sortBy, sortOrder])

  // Paginate
  const paginatedData = useMemo(() => {
    const start = (page - 1) * pageSize
    return filteredData.slice(start, start + pageSize)
  }, [filteredData, page, pageSize])

  return {
    data: paginatedData,
    totalCount: filteredData.length,
    isLoading: isProjectLoading || isQueryLoading,
    isFetching,
    error: error?.message ?? null,
    refetch,
    hasProject,
    projectSlug,
  }
}
