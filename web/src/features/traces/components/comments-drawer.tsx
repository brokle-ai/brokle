'use client'

import * as React from 'react'
import { useSearchParams, usePathname, useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { MessageCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCurrentUser } from '@/features/authentication'
import {
  useComments,
  useCommentCount,
  useCreateComment,
  useUpdateComment,
  useDeleteComment,
  useToggleReaction,
  useCreateReply,
} from '../hooks/use-comments'
import { CommentList } from './comment-list'
import { CommentForm } from './comment-form'

interface CommentsDrawerProps {
  projectId: string
  traceId: string
  className?: string
}

/**
 * CommentsDrawer - Drawer with badge count for trace comments
 *
 * Features:
 * - Sheet component (side="right", width 400-540px)
 * - Trigger button with MessageCircle icon + badge count (capped at "99+")
 * - Header with title and comment count
 * - Scrollable comment list
 * - Fixed comment form at bottom
 * - Deep linking support: ?comments=open query param
 */
export function CommentsDrawer({
  projectId,
  traceId,
  className,
}: CommentsDrawerProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // Deep linking: use URL as source of truth for open state
  const isOpenFromUrl = searchParams.get('comments') === 'open'

  // Handle open/close by updating URL (URL is the source of truth)
  const handleOpenChange = React.useCallback(
    (newOpen: boolean) => {
      const params = new URLSearchParams(searchParams.toString())
      if (newOpen) {
        params.set('comments', 'open')
      } else {
        params.delete('comments')
      }

      const newUrl = params.toString()
        ? `${pathname}?${params.toString()}`
        : pathname
      router.replace(newUrl, { scroll: false })
    },
    [pathname, router, searchParams]
  )

  // Get current user for ownership checks
  const { data: currentUser } = useCurrentUser()

  // Queries
  const { data: commentsData, isLoading: isLoadingComments } = useComments(
    projectId,
    traceId
  )
  const { data: countData } = useCommentCount(projectId, traceId)

  // Mutations
  const createMutation = useCreateComment(projectId, traceId)
  const updateMutation = useUpdateComment(projectId, traceId)
  const deleteMutation = useDeleteComment(projectId, traceId)
  const reactionMutation = useToggleReaction(projectId, traceId)
  const replyMutation = useCreateReply(projectId, traceId)

  const comments = commentsData?.comments || []
  const commentCount = countData?.count || 0

  const handleCreate = (content: string) => {
    createMutation.mutate({ content })
  }

  const handleUpdate = (commentId: string, content: string) => {
    updateMutation.mutate({ commentId, data: { content } })
  }

  const handleDelete = (commentId: string) => {
    deleteMutation.mutate(commentId)
  }

  const handleToggleReaction = (commentId: string, emoji: string) => {
    reactionMutation.mutate({ commentId, emoji })
  }

  const handleReply = (parentId: string, content: string) => {
    replyMutation.mutate({ parentId, content })
  }

  // Format badge count
  const badgeText = commentCount > 99 ? '99+' : commentCount.toString()

  return (
    <Sheet open={isOpenFromUrl} onOpenChange={handleOpenChange}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <SheetTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className={cn('h-8 w-8 relative', className)}
              >
                <MessageCircle className='h-4 w-4' />
                {commentCount > 0 && (
                  <Badge
                    variant='secondary'
                    className='absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center'
                  >
                    {badgeText}
                  </Badge>
                )}
                <span className='sr-only'>Comments</span>
              </Button>
            </SheetTrigger>
          </TooltipTrigger>
          <TooltipContent>
            {commentCount === 0
              ? 'Add a comment'
              : `${commentCount} comment${commentCount === 1 ? '' : 's'}`}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <SheetContent
        side='right'
        className='w-full sm:max-w-md lg:max-w-lg flex flex-col'
        hideCloseButton={false}
      >
        <SheetHeader className='border-b pb-4'>
          <SheetTitle className='flex items-center gap-2'>
            <MessageCircle className='h-5 w-5' />
            Comments
            {commentCount > 0 && (
              <Badge variant='secondary' className='ml-1'>
                {commentCount}
              </Badge>
            )}
          </SheetTitle>
        </SheetHeader>

        {/* Scrollable comment list */}
        <div className='flex-1 overflow-y-auto py-4'>
          <CommentList
            comments={comments}
            currentUserId={currentUser?.id}
            onUpdate={handleUpdate}
            onDelete={handleDelete}
            onToggleReaction={handleToggleReaction}
            onReply={handleReply}
            updatingCommentId={
              updateMutation.isPending
                ? updateMutation.variables?.commentId
                : undefined
            }
            deletingCommentId={
              deleteMutation.isPending ? deleteMutation.variables : undefined
            }
            replyingToCommentId={
              replyMutation.isPending
                ? replyMutation.variables?.parentId
                : undefined
            }
            isLoading={isLoadingComments}
          />
        </div>

        {/* Fixed comment form at bottom */}
        <div className='border-t pt-4'>
          <CommentForm
            onSubmit={handleCreate}
            isSubmitting={createMutation.isPending}
            placeholder='Add a comment...'
          />
        </div>
      </SheetContent>
    </Sheet>
  )
}
