import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { traceQueryKeys } from './trace-query-keys'
import { updateTraceBookmark } from '../api/traces-api'

/**
 * Hook for updating trace bookmark status with cache invalidation
 */
export function useUpdateTraceBookmark(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ traceId, bookmarked }: { traceId: string; bookmarked: boolean }) =>
      updateTraceBookmark(projectId, traceId, bookmarked),
    onSuccess: (_, { traceId, bookmarked }) => {
      // Invalidate traces list to refresh bookmark status in table view
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.list(projectId) })
      // Invalidate specific trace detail to refresh bookmark in detail view
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.detail(projectId, traceId) })
      toast.success(bookmarked ? 'Trace bookmarked' : 'Bookmark removed')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update bookmark')
    },
  })
}
