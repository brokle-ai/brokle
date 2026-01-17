'use client'

import * as React from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { MoreHorizontal, Pencil, Trash2, MessageSquare, ChevronDown, ChevronRight } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { cn } from '@/lib/utils'
import type { Comment } from '../api/comments-api'
import { CommentForm } from './comment-form'
import { ReactionBar } from './reaction-bar'

interface CommentItemProps {
  comment: Comment
  currentUserId?: string
  onUpdate: (commentId: string, content: string) => void
  onDelete: (commentId: string) => void
  onToggleReaction: (commentId: string, emoji: string) => void
  onReply?: (parentId: string, content: string) => void
  isUpdating?: boolean
  isDeleting?: boolean
  isReplying?: boolean
  /** Whether this is a nested reply (affects styling and disables reply button) */
  isReply?: boolean
}

/**
 * CommentItem - Single comment display with edit/delete actions
 *
 * Features:
 * - Avatar with fallback initials
 * - Author name and relative timestamp
 * - Comment content (preserves whitespace)
 * - Edited indicator
 * - Actions dropdown (owner only)
 * - Inline edit mode
 * - Emoji reactions with toggle
 * - Reply button and nested replies (top-level only)
 */
export function CommentItem({
  comment,
  currentUserId,
  onUpdate,
  onDelete,
  onToggleReaction,
  onReply,
  isUpdating = false,
  isDeleting = false,
  isReplying = false,
  isReply = false,
}: CommentItemProps) {
  const [isEditing, setIsEditing] = React.useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = React.useState(false)
  const [isReplyFormOpen, setIsReplyFormOpen] = React.useState(false)
  const [repliesExpanded, setRepliesExpanded] = React.useState(true)

  const isOwner = currentUserId === comment.created_by
  const authorName = comment.author?.name || 'Unknown User'
  const authorInitials = authorName
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  const handleUpdate = (content: string) => {
    onUpdate(comment.id, content)
    setIsEditing(false)
  }

  const handleDelete = () => {
    onDelete(comment.id)
    setShowDeleteDialog(false)
  }

  const handleReply = (content: string) => {
    onReply?.(comment.id, content)
    setIsReplyFormOpen(false)
  }

  const createdAt = new Date(comment.created_at)
  const timeAgo = formatDistanceToNow(createdAt, { addSuffix: true })
  const hasReplies = comment.replies && comment.replies.length > 0
  const canReply = !isReply && onReply

  if (isEditing) {
    return (
      <div className='p-4 bg-muted/50 rounded-lg'>
        <CommentForm
          initialContent={comment.content}
          onSubmit={handleUpdate}
          onCancel={() => setIsEditing(false)}
          isSubmitting={isUpdating}
          placeholder='Edit your comment...'
          submitLabel='Save'
        />
      </div>
    )
  }

  return (
    <>
      <div className='group flex gap-3 p-4 rounded-lg hover:bg-muted/30 transition-colors'>
        {/* Avatar */}
        <Avatar className='h-8 w-8 shrink-0'>
          <AvatarImage src={comment.author?.avatar_url || undefined} alt={authorName} />
          <AvatarFallback className='text-xs'>{authorInitials}</AvatarFallback>
        </Avatar>

        {/* Content */}
        <div className='flex-1 min-w-0'>
          {/* Header row */}
          <div className='flex items-center justify-between gap-2'>
            <div className='flex items-center gap-2 min-w-0'>
              <span className='text-sm font-medium truncate'>{authorName}</span>
              <span className='text-xs text-muted-foreground shrink-0'>{timeAgo}</span>
              {comment.is_edited && (
                <span className='text-xs text-muted-foreground shrink-0'>(edited)</span>
              )}
            </div>

            {/* Actions dropdown - only for owner */}
            {isOwner && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant='ghost'
                    size='icon'
                    className={cn(
                      'h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity',
                      (isUpdating || isDeleting) && 'opacity-50 pointer-events-none'
                    )}
                  >
                    <MoreHorizontal className='h-4 w-4' />
                    <span className='sr-only'>Comment actions</span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align='end'>
                  <DropdownMenuItem onClick={() => setIsEditing(true)}>
                    <Pencil className='h-4 w-4 mr-2' />
                    Edit
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className='text-destructive focus:text-destructive'
                    onClick={() => setShowDeleteDialog(true)}
                  >
                    <Trash2 className='h-4 w-4 mr-2' />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          {/* Comment content */}
          <p className='text-sm mt-1 whitespace-pre-wrap break-words'>{comment.content}</p>

          {/* Reaction bar and reply button */}
          <div className='flex items-center gap-2 mt-2'>
            <ReactionBar
              reactions={comment.reactions || []}
              onToggleReaction={(emoji) => onToggleReaction(comment.id, emoji)}
              disabled={isUpdating || isDeleting}
            />
            {canReply && (
              <Button
                variant='ghost'
                size='sm'
                className='h-7 px-2 text-muted-foreground hover:text-foreground'
                onClick={() => setIsReplyFormOpen(true)}
                disabled={isReplying}
              >
                <MessageSquare className='h-4 w-4 mr-1' />
                Reply
              </Button>
            )}
          </div>

          {/* Inline reply form */}
          {isReplyFormOpen && (
            <div className='mt-3 pl-4 border-l-2 border-muted'>
              <CommentForm
                onSubmit={handleReply}
                onCancel={() => setIsReplyFormOpen(false)}
                isSubmitting={isReplying}
                placeholder='Write a reply...'
                submitLabel='Reply'
                autoFocus
              />
            </div>
          )}

          {/* Nested replies */}
          {hasReplies && !isReply && (
            <Collapsible
              open={repliesExpanded}
              onOpenChange={setRepliesExpanded}
              className='mt-3'
            >
              <CollapsibleTrigger asChild>
                <Button
                  variant='ghost'
                  size='sm'
                  className='h-7 px-2 text-muted-foreground hover:text-foreground'
                >
                  {repliesExpanded ? (
                    <ChevronDown className='h-4 w-4 mr-1' />
                  ) : (
                    <ChevronRight className='h-4 w-4 mr-1' />
                  )}
                  {comment.reply_count} {comment.reply_count === 1 ? 'reply' : 'replies'}
                </Button>
              </CollapsibleTrigger>
              <CollapsibleContent className='mt-2 pl-4 border-l-2 border-muted space-y-1'>
                {comment.replies!.map((reply) => (
                  <CommentItem
                    key={reply.id}
                    comment={reply}
                    currentUserId={currentUserId}
                    onUpdate={onUpdate}
                    onDelete={onDelete}
                    onToggleReaction={onToggleReaction}
                    isUpdating={isUpdating}
                    isDeleting={isDeleting}
                    isReply
                  />
                ))}
              </CollapsibleContent>
            </Collapsible>
          )}
        </div>
      </div>

      {/* Delete confirmation dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete comment?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete your comment.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
              disabled={isDeleting}
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
