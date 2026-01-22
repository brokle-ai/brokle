import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { traceQueryKeys } from './trace-query-keys'
import {
  listComments,
  getCommentCount,
  createComment,
  updateComment,
  deleteComment,
  toggleReaction,
  createReply,
  type Comment,
  type ListCommentsResponse,
  type CommentCountResponse,
  type CreateCommentRequest,
  type UpdateCommentRequest,
  type ToggleReactionRequest,
  type ReactionSummary,
} from '../api/comments-api'

/**
 * Hook for fetching comments for a trace
 */
export function useComments(projectId: string, traceId: string) {
  return useQuery<ListCommentsResponse>({
    queryKey: traceQueryKeys.comments(projectId, traceId),
    queryFn: () => listComments(projectId, traceId),
    enabled: Boolean(projectId && traceId),
  })
}

/**
 * Hook for fetching comment count for a trace
 */
export function useCommentCount(projectId: string, traceId: string) {
  return useQuery<CommentCountResponse>({
    queryKey: traceQueryKeys.commentCount(projectId, traceId),
    queryFn: () => getCommentCount(projectId, traceId),
    enabled: Boolean(projectId && traceId),
  })
}

/**
 * Hook for creating a new comment
 */
export function useCreateComment(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateCommentRequest) =>
      createComment(projectId, traceId, data),
    onSuccess: () => {
      // Invalidate comments list and count
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.commentCount(projectId, traceId),
      })
      toast.success('Comment added')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add comment')
    },
  })
}

/**
 * Hook for updating a comment
 */
export function useUpdateComment(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      commentId,
      data,
    }: {
      commentId: string
      data: UpdateCommentRequest
    }) => updateComment(projectId, traceId, commentId, data),
    onSuccess: () => {
      // Invalidate comments list
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })
      toast.success('Comment updated')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update comment')
    },
  })
}

/**
 * Hook for deleting a comment
 */
export function useDeleteComment(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (commentId: string) =>
      deleteComment(projectId, traceId, commentId),
    onSuccess: () => {
      // Invalidate comments list and count
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.commentCount(projectId, traceId),
      })
      toast.success('Comment deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete comment')
    },
  })
}

/**
 * Hook for toggling a reaction on a comment with optimistic updates
 */
export function useToggleReaction(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      commentId,
      emoji,
    }: {
      commentId: string
      emoji: string
    }) => toggleReaction(projectId, traceId, commentId, { emoji }),
    onMutate: async ({ commentId, emoji }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })

      // Snapshot previous value
      const previousComments = queryClient.getQueryData<ListCommentsResponse>(
        traceQueryKeys.comments(projectId, traceId)
      )

      // Optimistically update the reaction
      if (previousComments) {
        queryClient.setQueryData<ListCommentsResponse>(
          traceQueryKeys.comments(projectId, traceId),
          {
            ...previousComments,
            comments: updateReactionInComments(
              previousComments.comments,
              commentId,
              emoji
            ),
          }
        )
      }

      return { previousComments }
    },
    onError: (error: Error, _variables, context) => {
      // Rollback on error
      if (context?.previousComments) {
        queryClient.setQueryData(
          traceQueryKeys.comments(projectId, traceId),
          context.previousComments
        )
      }
      toast.error(error.message || 'Failed to toggle reaction')
    },
    onSettled: () => {
      // Refetch after mutation settles
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })
    },
  })
}

/**
 * Helper to update reaction in comments tree (handles nested replies)
 */
function updateReactionInComments(
  comments: Comment[],
  commentId: string,
  emoji: string
): Comment[] {
  return comments.map((comment) => {
    if (comment.id === commentId) {
      return {
        ...comment,
        reactions: toggleReactionInList(comment.reactions, emoji),
      }
    }
    if (comment.replies && comment.replies.length > 0) {
      return {
        ...comment,
        replies: updateReactionInComments(comment.replies, commentId, emoji),
      }
    }
    return comment
  })
}

/**
 * Helper to toggle a reaction in a reaction list (optimistic)
 */
function toggleReactionInList(
  reactions: ReactionSummary[],
  emoji: string
): ReactionSummary[] {
  const existing = reactions.find((r) => r.emoji === emoji)

  if (existing) {
    if (existing.has_user) {
      // User is removing their reaction
      if (existing.count === 1) {
        // Remove the reaction entirely
        return reactions.filter((r) => r.emoji !== emoji)
      }
      // Decrement count
      return reactions.map((r) =>
        r.emoji === emoji
          ? { ...r, count: r.count - 1, has_user: false }
          : r
      )
    } else {
      // User is adding their reaction
      return reactions.map((r) =>
        r.emoji === emoji
          ? { ...r, count: r.count + 1, has_user: true }
          : r
      )
    }
  } else {
    // New reaction type
    return [
      ...reactions,
      { emoji, count: 1, users: [], has_user: true },
    ]
  }
}

/**
 * Hook for creating a reply to a comment
 */
export function useCreateReply(projectId: string, traceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      parentId,
      content,
    }: {
      parentId: string
      content: string
    }) => createReply(projectId, traceId, parentId, { content }),
    onSuccess: () => {
      // Invalidate comments list and count
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.comments(projectId, traceId),
      })
      queryClient.invalidateQueries({
        queryKey: traceQueryKeys.commentCount(projectId, traceId),
      })
      toast.success('Reply added')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add reply')
    },
  })
}
