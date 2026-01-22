'use client'

import * as React from 'react'
import { Skeleton } from '@/components/ui/skeleton'
import { MessageSquare } from 'lucide-react'
import type { Comment } from '../api/comments-api'
import { CommentItem } from './comment-item'

interface CommentListProps {
  comments: Comment[]
  currentUserId?: string
  onUpdate: (commentId: string, content: string) => void
  onDelete: (commentId: string) => void
  onToggleReaction: (commentId: string, emoji: string) => void
  onReply: (parentId: string, content: string) => void
  updatingCommentId?: string
  deletingCommentId?: string
  replyingToCommentId?: string
  isLoading?: boolean
}

/**
 * CommentList - Renders list of comments with loading and empty states
 *
 * Features:
 * - Chronological order (oldest first)
 * - Auto-scroll to latest on new comment
 * - Loading skeleton (3 items)
 * - Empty state message
 */
export function CommentList({
  comments,
  currentUserId,
  onUpdate,
  onDelete,
  onToggleReaction,
  onReply,
  updatingCommentId,
  deletingCommentId,
  replyingToCommentId,
  isLoading = false,
}: CommentListProps) {
  const listEndRef = React.useRef<HTMLDivElement>(null)
  const prevCommentsLength = React.useRef(comments.length)

  // Auto-scroll to latest comment when new comment is added
  React.useEffect(() => {
    if (comments.length > prevCommentsLength.current) {
      listEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
    prevCommentsLength.current = comments.length
  }, [comments.length])

  if (isLoading) {
    return (
      <div className='space-y-4'>
        {[1, 2, 3].map((i) => (
          <div key={i} className='flex gap-3 p-4'>
            <Skeleton className='h-8 w-8 rounded-full shrink-0' />
            <div className='flex-1 space-y-2'>
              <div className='flex items-center gap-2'>
                <Skeleton className='h-4 w-24' />
                <Skeleton className='h-3 w-16' />
              </div>
              <Skeleton className='h-4 w-full' />
              <Skeleton className='h-4 w-3/4' />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (comments.length === 0) {
    return (
      <div className='flex flex-col items-center justify-center py-12 text-center'>
        <MessageSquare className='h-12 w-12 text-muted-foreground/50 mb-4' />
        <p className='text-sm text-muted-foreground'>No comments yet.</p>
        <p className='text-sm text-muted-foreground'>Start the conversation!</p>
      </div>
    )
  }

  return (
    <div className='space-y-1'>
      {comments.map((comment) => (
        <CommentItem
          key={comment.id}
          comment={comment}
          currentUserId={currentUserId}
          onUpdate={onUpdate}
          onDelete={onDelete}
          onToggleReaction={onToggleReaction}
          onReply={onReply}
          isUpdating={updatingCommentId === comment.id}
          isDeleting={deletingCommentId === comment.id}
          isReplying={replyingToCommentId === comment.id}
        />
      ))}
      <div ref={listEndRef} />
    </div>
  )
}
