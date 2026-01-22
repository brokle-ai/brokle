'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { traceQueryKeys } from './trace-query-keys'
import {
  getTraceScores,
  createAnnotation,
  deleteAnnotation,
  type Annotation,
  type CreateAnnotationRequest,
} from '../api/scores-api'

/**
 * Hook for fetching all scores (annotations + automated) for a trace
 */
export function useAnnotations(projectId: string, traceId: string) {
  return useQuery<Annotation[]>({
    queryKey: traceQueryKeys.scores(projectId, traceId),
    queryFn: () => getTraceScores(projectId, traceId),
    enabled: Boolean(projectId && traceId),
  })
}

/**
 * Hook for fetching only human annotations for a trace
 * Filters to only source = 'annotation'
 */
export function useHumanAnnotations(projectId: string, traceId: string) {
  const query = useAnnotations(projectId, traceId)

  return {
    ...query,
    data: query.data?.filter(score => score.source === 'annotation'),
  }
}

/**
 * Hook for fetching only automated scores (api, eval) for a trace
 */
export function useAutomatedScores(projectId: string, traceId: string) {
  const query = useAnnotations(projectId, traceId)

  return {
    ...query,
    data: query.data?.filter(score => score.source !== 'annotation'),
  }
}

/**
 * Hook for getting annotation count for a trace
 */
export function useAnnotationCount(projectId: string, traceId: string) {
  const query = useAnnotations(projectId, traceId)

  return {
    ...query,
    data: query.data?.length ?? 0,
  }
}

/**
 * Hook for creating a new annotation
 */
export function useCreateAnnotation(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateAnnotationRequest) =>
      createAnnotation(projectId, traceId, data),
    onSuccess: () => {
      // Invalidate scores list
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.scores(projectId, traceId),
      })
      toast.success('Annotation added')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add annotation')
    },
  })
}

/**
 * Hook for deleting an annotation
 */
export function useDeleteAnnotation(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (scoreId: string) =>
      deleteAnnotation(projectId, traceId, scoreId),
    onSuccess: () => {
      // Invalidate scores list
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.scores(projectId, traceId),
      })
      toast.success('Annotation deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete annotation')
    },
  })
}
