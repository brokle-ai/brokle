'use client'

import { useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import { useProjectOnly } from '@/features/projects'
import { traceQueryKeys } from './trace-query-keys'
import { getProjectTraces } from '../api/traces-api'
import type { Trace } from '../data/schema'

/**
 * Session data aggregated from traces
 */
export interface Session {
  session_id: string
  trace_count: number
  first_trace: Date
  last_trace: Date
  total_duration: number // nanoseconds
  total_tokens: number
  total_cost: number
  user_ids: string[]
  traces: Trace[]
}

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
    sessionsPage: parseAsInteger.withDefault(1),
    sessionsPageSize: parseAsInteger.withDefault(20),
    sessionSearch: parseAsString,
    sessionsSortBy: parseAsString,
    sessionsSortOrder: parseAsString,
  })

  const setSearch = useCallback(
    (search: string | null) => {
      setQuery({ sessionSearch: search, sessionsPage: 1 })
    },
    [setQuery]
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        sessionsPage: Math.max(1, page),
        ...(pageSize !== undefined && { sessionsPageSize: Math.max(1, pageSize) }),
      })
    },
    [setQuery]
  )

  const setSorting = useCallback(
    (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => {
      setQuery({
        sessionsSortBy: sortBy || null,
        sessionsSortOrder: sortOrder || null,
      })
    },
    [setQuery]
  )

  const resetAll = useCallback(() => {
    setQuery({
      sessionsPage: 1,
      sessionsPageSize: null,
      sessionSearch: null,
      sessionsSortBy: null,
      sessionsSortOrder: null,
    })
  }, [setQuery])

  return {
    page: Math.max(1, query.sessionsPage),
    pageSize: Math.max(1, query.sessionsPageSize),
    search: query.sessionSearch,
    sortBy: query.sessionsSortBy,
    sortOrder: query.sessionsSortOrder as 'asc' | 'desc' | null,

    setSearch,
    setPagination,
    setSorting,
    resetAll,

    hasActiveFilters: !!query.sessionSearch,
  }
}

/**
 * Aggregate traces into sessions
 */
function aggregateTracesToSessions(traces: Trace[]): Session[] {
  const sessionMap = new Map<string, Session>()

  traces.forEach((trace) => {
    const sessionId = trace.session_id || 'no-session'

    if (!sessionMap.has(sessionId)) {
      sessionMap.set(sessionId, {
        session_id: sessionId,
        trace_count: 0,
        first_trace: trace.start_time,
        last_trace: trace.start_time,
        total_duration: 0,
        total_tokens: 0,
        total_cost: 0,
        user_ids: [],
        traces: [],
      })
    }

    const session = sessionMap.get(sessionId)!

    session.trace_count++
    session.traces.push(trace)

    // Update timestamps
    if (trace.start_time < session.first_trace) {
      session.first_trace = trace.start_time
    }
    if (trace.start_time > session.last_trace) {
      session.last_trace = trace.start_time
    }

    // Aggregate metrics
    session.total_duration += trace.duration || 0
    session.total_tokens += trace.tokens || 0
    session.total_cost += trace.cost || 0

    // Collect unique user IDs
    if (trace.user_id && !session.user_ids.includes(trace.user_id)) {
      session.user_ids.push(trace.user_id)
    }
  })

  return Array.from(sessionMap.values())
}

/**
 * Hook to fetch and manage project sessions (traces grouped by session_id).
 *
 * Note: This aggregates traces client-side. Future improvement could add
 * a backend endpoint for server-side session aggregation.
 */
export function useProjectSessions() {
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()
  const tableState = useSessionsTableState()
  const projectId = currentProject?.id

  // Build query params for fetching traces with session_id
  const queryParams = useMemo(() => {
    if (!projectId) return undefined
    return {
      page: 1,
      pageSize: 500, // Fetch larger batch for client-side aggregation
      search: tableState.search || undefined,
    }
  }, [projectId, tableState.search])

  const {
    data: tracesData,
    isLoading: isTracesLoading,
    isFetching: isTracesFetching,
    error,
    refetch,
  } = useQuery({
    queryKey: traceQueryKeys.sessions(projectId!, queryParams),

    queryFn: async () => {
      if (!projectId) throw new Error('No project selected')

      // Fetch traces - we'll aggregate by session_id client-side
      const result = await getProjectTraces({
        projectId,
        page: 1,
        pageSize: 500, // Get enough for aggregation
      })

      return result
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

  // Aggregate traces into sessions
  const sessions = useMemo(() => {
    if (!tracesData?.traces) return []

    // Filter to traces that have a session_id
    const tracesWithSession = tracesData.traces.filter((t) => t.session_id)

    const aggregated = aggregateTracesToSessions(tracesWithSession)

    // Sort sessions
    if (tableState.sortBy) {
      const sortOrder = tableState.sortOrder === 'desc' ? -1 : 1
      aggregated.sort((a, b) => {
        let aVal: any, bVal: any

        switch (tableState.sortBy) {
          case 'trace_count':
            aVal = a.trace_count
            bVal = b.trace_count
            break
          case 'total_duration':
            aVal = a.total_duration
            bVal = b.total_duration
            break
          case 'total_tokens':
            aVal = a.total_tokens
            bVal = b.total_tokens
            break
          case 'total_cost':
            aVal = a.total_cost
            bVal = b.total_cost
            break
          case 'last_trace':
          default:
            aVal = a.last_trace.getTime()
            bVal = b.last_trace.getTime()
            break
        }

        return (aVal - bVal) * sortOrder
      })
    } else {
      // Default sort by most recent activity
      aggregated.sort((a, b) => b.last_trace.getTime() - a.last_trace.getTime())
    }

    return aggregated
  }, [tracesData?.traces, tableState.sortBy, tableState.sortOrder])

  // Paginate sessions client-side
  const paginatedSessions = useMemo(() => {
    const start = (tableState.page - 1) * tableState.pageSize
    return sessions.slice(start, start + tableState.pageSize)
  }, [sessions, tableState.page, tableState.pageSize])

  return {
    data: paginatedSessions,
    totalCount: sessions.length,
    page: tableState.page,
    pageSize: tableState.pageSize,
    totalPages: Math.ceil(sessions.length / tableState.pageSize),

    isLoading: isProjectLoading || isTracesLoading,
    isFetching: isTracesFetching,
    isProjectLoading,
    isSessionsLoading: isTracesLoading,

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
