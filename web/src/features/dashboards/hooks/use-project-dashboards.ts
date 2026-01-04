'use client'

import { useMemo } from 'react'
import { useProjectOnly } from '@/features/projects'
import { useDashboardsQuery } from './use-dashboards-queries'
import type { Dashboard, DashboardFilter } from '../types'

interface UseProjectDashboardsOptions {
  filter?: DashboardFilter
}

export function useProjectDashboards(options: UseProjectDashboardsOptions = {}) {
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id

  const {
    data: response,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useDashboardsQuery(projectId, options.filter)

  const dashboards = useMemo(() => {
    return response?.dashboards ?? []
  }, [response?.dashboards])

  const totalCount = response?.total ?? 0

  return {
    data: dashboards as Dashboard[],
    totalCount,
    isLoading,
    isFetching,
    error: error?.message || null,
    hasProject: !!projectId,
    refetch,
    currentProject,
  }
}
