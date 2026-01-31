'use client'

import { useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import { useProjectOnly } from '@/features/projects'
import { getProjectSessions, type Session } from '../api/sessions-api'

export type { Session }

export interface UseSessionsTableStateReturn {
  page: number
  pageSize: number
  search: string | null
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null

  setSearch: (search: string | null) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void

  hasActiveFilters: boolean
}

/**
 * Hook to manage sessions table URL state.
 */
export function useSessionsTableState(): UseSessionsTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(20),
    search: parseAsString,
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  const setSearch = useCallback(
    (search: string | null) => {
      setQuery({ search: search, page: 1 })
    },
    [setQuery]
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        page: Math.max(1, page),
        ...(pageSize !== undefined && { pageSize: Math.max(1, pageSize) }),
      })
    },
    [setQuery]
  )

  const setSorting = useCallback(
    (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => {
      setQuery({
        sortBy: sortBy || null,
        sortOrder: sortOrder || null,
      })
    },
    [setQuery]
  )

  const resetAll = useCallback(() => {
    setQuery({
      page: 1,
      pageSize: null,
      search: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  return {
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    sortBy: query.sortBy,
    sortOrder: query.sortOrder as 'asc' | 'desc' | null,

    setSearch,
    setPagination,
    setSorting,
    resetAll,

    hasActiveFilters: !!query.search,
  }
}

/**
 * Hook to fetch project sessions using the server-side aggregated sessions endpoint.
 *
 * This uses the backend sessions endpoint which performs server-side GROUP BY
 * aggregation in ClickHouse, providing accurate session data across ALL traces
 * in the project, not limited to a sample size.
 */
export function useProjectSessions() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = useSessionsTableState()
  const projectId = currentProject?.id

  const {
    data: sessionsData,
    isLoading: isSessionsLoading,
    isFetching: isSessionsFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: [
      'projects',
      projectId,
      'sessions',
      {
        page: tableState.page,
        pageSize: tableState.pageSize,
        search: tableState.search,
        sortBy: tableState.sortBy,
        sortOrder: tableState.sortOrder,
      },
    ],

    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')

      // Use the server-side sessions endpoint for accurate aggregation
      return getProjectSessions({
        projectId,
        page: tableState.page,
        pageSize: tableState.pageSize,
        search: tableState.search || undefined,
        sortBy: tableState.sortBy || undefined,
        sortOrder: tableState.sortOrder || undefined,
      })
    },

    enabled: !!projectId && hasProject,

    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    refetchInterval: 30_000,

    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  return {
    data: sessionsData?.sessions ?? [],
    totalCount: sessionsData?.totalCount ?? 0,
    page: tableState.page,
    pageSize: tableState.pageSize,
    totalPages: sessionsData?.totalPages ?? 0,

    isLoading: isProjectLoading || isSessionsLoading,
    isFetching: isSessionsFetching,
    isProjectLoading,
    isSessionsLoading,

    error: error instanceof Error ? error.message : error ? String(error) : null,

    refetch,

    hasProject,
    currentProject,

    tableState,
  }
}

export interface UseProjectSessionsReturn {
  data: Session[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  isLoading: boolean
  isFetching: boolean
  isProjectLoading: boolean
  isSessionsLoading: boolean
  error: string | null
  refetch: () => void
  hasProject: boolean
  currentProject: ReturnType<typeof useProjectOnly>['currentProject']
  tableState: UseSessionsTableStateReturn
}
