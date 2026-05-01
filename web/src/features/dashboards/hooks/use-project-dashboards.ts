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

  // useDashboardsQuery returns the full {data, pagination} envelope.
  // Surface dashboards (the page items) and totalCount (the server's
  // authoritative total) — derived from page length would be wrong as
  // soon as the caller passes limit/offset.
  const dashboards = useMemo(() => response?.data ?? [], [response?.data])
  const totalCount = response?.pagination?.total ?? 0

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
