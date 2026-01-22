import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { traceQueryKeys } from './trace-query-keys'
import { updateTraceTags } from '../api/traces-api'

/**
 * Hook for updating trace tags with cache invalidation
 *
 * Tags are normalized on the backend (lowercase, unique, sorted)
 */
export function useUpdateTraceTags(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ traceId, tags }: { traceId: string; tags: string[] }) =>
      updateTraceTags(projectId, traceId, tags),
    onSuccess: (_, { traceId }) => {
      // Invalidate traces list to refresh tags in table view
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.list(projectId) })
      // Invalidate specific trace detail to refresh tags in detail view
      queryClient.invalidateQueries({ queryKey: traceQueryKeys.detail(projectId, traceId) })
      toast.success('Tags updated')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update tags')
    },
  })
}
