import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { traceQueryKeys } from './trace-query-keys'
import { deleteTrace } from '../api/traces-api'

/**
 * Hook for deleting a trace with optimistic updates and cache invalidation
 */
export function useDeleteTrace(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (traceId: string) => deleteTrace(projectId, traceId),
    onSuccess: (_data, traceId) => {
      // Invalidate traces list to trigger refetch
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.list(projectId) })

      // Remove detail queries for the deleted trace (no refetch - data no longer exists)
      // This also removes spans and scores queries as they use detail as a prefix
      queryClient.removeQueries({ queryKey: traceQueryKeys.detail(projectId, traceId) })

      toast.success('Trace deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete trace')
    },
  })
}

/**
 * Delete multiple traces
 *
 * Executes individual deletions in parallel
 * Note: This is a client-side batch operation until a bulk delete API is available
 */
async function deleteMultipleTraces(
  projectId: string,
  traceIds: string[]
): Promise<{ succeeded: string[]; failed: string[] }> {
  const results = await Promise.allSettled(
    traceIds.map((traceId) => deleteTrace(projectId, traceId))
  )

  const succeeded: string[] = []
  const failed: string[] = []

  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      succeeded.push(traceIds[index])
    } else {
      failed.push(traceIds[index])
    }
  })

  return { succeeded, failed }
}

/**
 * Hook for deleting multiple traces
 */
export function useDeleteMultipleTraces(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (traceIds: string[]) => deleteMultipleTraces(projectId, traceIds),
    onSuccess: ({ succeeded, failed }) => {
      // Invalidate traces list
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.list(projectId) })

      // Remove detail queries for each successfully deleted trace
      // This also removes spans and scores queries as they use detail as a prefix
      succeeded.forEach((traceId) => {
        queryClient.removeQueries({ queryKey: traceQueryKeys.detail(projectId, traceId) })
      })

      if (failed.length === 0) {
        toast.success(`${succeeded.length} trace${succeeded.length !== 1 ? 's' : ''} deleted`)
      } else if (succeeded.length === 0) {
        toast.error('Failed to delete traces')
      } else {
        toast.warning(
          `${succeeded.length} deleted, ${failed.length} failed`
        )
      }
    },
    onError: () => {
      toast.error('Failed to delete traces')
    },
  })
}
