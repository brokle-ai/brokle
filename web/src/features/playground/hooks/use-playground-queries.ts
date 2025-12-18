'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  createSession,
  getSession,
  listSessions,
  updateSession,
  deleteSession,
} from '../api/playground-api'
import type {
  PlaygroundSessionSummary,
  CreateSessionRequest,
  UpdateSessionRequest,
} from '../types'

// ============================================================================
// Query Keys Factory
// ============================================================================

export const playgroundQueryKeys = {
  all: ['playground'] as const,

  // Sessions
  sessions: () => [...playgroundQueryKeys.all, 'session'] as const,
  session: (projectId: string, sessionId: string) =>
    [...playgroundQueryKeys.sessions(), projectId, sessionId] as const,

  // Sessions list
  sessionsList: (projectId: string, filters?: SessionsFilters) =>
    [...playgroundQueryKeys.sessions(), 'list', projectId, filters] as const,
}

// ============================================================================
// Types
// ============================================================================

export interface SessionsFilters {
  limit?: number
  tags?: string[]
}

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Query hook to get a single playground session by ID
 */
export function useSessionQuery(
  projectId: string | undefined,
  sessionId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: playgroundQueryKeys.session(projectId || '', sessionId || ''),
    queryFn: async () => {
      if (!projectId || !sessionId) {
        throw new Error('Project ID and Session ID are required')
      }
      return getSession(projectId, sessionId)
    },
    enabled: !!projectId && !!sessionId && (options.enabled ?? true),
    staleTime: 30_000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Query hook to list sessions for a project
 */
export function useSessionsQuery(
  projectId: string | undefined,
  filters?: SessionsFilters,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: playgroundQueryKeys.sessionsList(projectId || '', filters),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return listSessions(projectId, filters)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

// ============================================================================
// Mutation Hooks
// ============================================================================

/**
 * Mutation hook to create a new playground session
 */
export function useCreateSessionMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateSessionRequest) => {
      return createSession(projectId, data)
    },
    onSuccess: (newSession) => {
      // Add the new session to the cache
      queryClient.setQueryData(
        playgroundQueryKeys.session(projectId, newSession.id),
        newSession
      )
      // Invalidate sessions list to refresh sidebar
      queryClient.invalidateQueries({
        queryKey: playgroundQueryKeys.sessions(),
      })
      toast.success('Session Saved', {
        description: `"${newSession.name}" has been saved.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Session', {
        description: apiError?.message || 'Could not create playground session. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to update a playground session
 */
export function useUpdateSessionMutation(projectId: string, sessionId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: UpdateSessionRequest) => {
      return updateSession(projectId, sessionId, data)
    },
    onSuccess: (updatedSession) => {
      // Update the cache
      queryClient.setQueryData(
        playgroundQueryKeys.session(projectId, sessionId),
        updatedSession
      )
      // Invalidate sessions list in case name/description changed
      queryClient.invalidateQueries({
        queryKey: playgroundQueryKeys.sessions(),
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Session', {
        description: apiError?.message || 'Could not update playground session. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to delete a playground session
 */
export function useDeleteSessionMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      sessionId,
      sessionName,
    }: {
      sessionId: string
      sessionName?: string
    }) => {
      await deleteSession(projectId, sessionId)
      return { sessionId, sessionName }
    },
    onMutate: async ({ sessionId }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({
        queryKey: playgroundQueryKeys.sessions(),
      })

      // Snapshot the previous value
      const previousSessions = queryClient.getQueryData<PlaygroundSessionSummary[]>(
        playgroundQueryKeys.sessionsList(projectId)
      )

      // Optimistically update sessions list
      if (previousSessions) {
        queryClient.setQueryData<PlaygroundSessionSummary[]>(
          playgroundQueryKeys.sessionsList(projectId),
          previousSessions.filter((s) => s.id !== sessionId)
        )
      }

      return { previousSessions }
    },
    onSuccess: (_data, variables) => {
      // Remove session from cache
      queryClient.removeQueries({
        queryKey: playgroundQueryKeys.session(projectId, variables.sessionId),
      })
      // Invalidate sessions list
      queryClient.invalidateQueries({
        queryKey: playgroundQueryKeys.sessions(),
      })
      toast.success('Session Deleted', {
        description: variables.sessionName
          ? `"${variables.sessionName}" has been deleted.`
          : 'Session has been deleted.',
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback on error
      if (context?.previousSessions) {
        queryClient.setQueryData(
          playgroundQueryKeys.sessionsList(projectId),
          context.previousSessions
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Session', {
        description: apiError?.message || 'Could not delete session. Please try again.',
      })
    },
  })
}
