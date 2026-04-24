import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { MessageSquare, Reply as ReplyIcon } from 'lucide-react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import type { TraceComment } from '../api/types'
import { ReactionBar } from './reaction-bar'

// Thread renderer. Top-level comments are rendered with replies
// nested one level below; deeper nesting isn't supported v1 (and the
// backend's `parent_id` is a single-level reference anyway — the
// reply endpoint rejects replies-to-replies). Reactions live inline
// on every comment, including replies.

interface CommentListProps {
  comments: TraceComment[]
  onReply: (parentId: string, content: string) => void
  onToggleReaction: (commentId: string, emoji: string) => void
  replyPendingId?: string
  reactionPendingId?: string
  isLoading?: boolean
}

export function CommentList({
  comments,
  onReply,
  onToggleReaction,
  replyPendingId,
  reactionPendingId,
  isLoading,
}: CommentListProps) {
  if (isLoading) {
    return (
      <div className="space-y-3 px-4">
        {[0, 1, 2].map((i) => (
          <div key={i} className="flex gap-3">
            <div className="h-8 w-8 shrink-0 animate-pulse rounded-full bg-muted" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-32 animate-pulse rounded bg-muted" />
              <div className="h-4 w-full animate-pulse rounded bg-muted" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (comments.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <MessageSquare className="mb-3 h-10 w-10 text-muted-foreground/40" />
        <p className="text-sm text-muted-foreground">No comments yet.</p>
        <p className="text-xs text-muted-foreground">
          Start the conversation.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-1">
      {comments.map((c) => (
        <CommentNode
          key={c.id}
          comment={c}
          onReply={onReply}
          onToggleReaction={onToggleReaction}
          replyPendingId={replyPendingId}
          reactionPendingId={reactionPendingId}
        />
      ))}
    </div>
  )
}

interface CommentNodeProps {
  comment: TraceComment
  onReply: (parentId: string, content: string) => void
  onToggleReaction: (commentId: string, emoji: string) => void
  replyPendingId?: string
  reactionPendingId?: string
  isReply?: boolean
}

function CommentNode({
  comment,
  onReply,
  onToggleReaction,
  replyPendingId,
  reactionPendingId,
  isReply,
}: CommentNodeProps) {
  const [replyOpen, setReplyOpen] = useState(false)
  const [replyContent, setReplyContent] = useState('')

  const author = comment.author
  const authorName = author?.name ?? 'Unknown user'
  const initials = authorName
    .split(' ')
    .map((part) => part[0])
    .filter(Boolean)
    .join('')
    .toUpperCase()
    .slice(0, 2)

  const handleSubmitReply = () => {
    const trimmed = replyContent.trim()
    if (!trimmed) return
    onReply(comment.id, trimmed)
    setReplyContent('')
    setReplyOpen(false)
  }

  // Nested replies are one-level. `isReply` is inherited from the
  // parent render so we don't render a Reply button on already-nested
  // items — that keeps the thread from accidentally growing two
  // levels when the backend would reject it anyway.
  const canReply = !isReply
  const hasReplies = comment.replies && comment.replies.length > 0

  const timeAgo = (() => {
    const d = new Date(comment.created_at)
    if (Number.isNaN(d.getTime())) return comment.created_at
    return formatDistanceToNow(d, { addSuffix: true })
  })()

  return (
    <div
      className={cn(
        'group rounded-md p-3 transition-colors hover:bg-muted/30',
      )}
    >
      <div className="flex gap-3">
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarImage src={author?.avatar_url} alt={authorName} />
          <AvatarFallback className="text-xs">{initials}</AvatarFallback>
        </Avatar>
        <div className="min-w-0 flex-1">
          <div className="flex items-baseline gap-2">
            <span className="truncate text-sm font-medium">
              {authorName}
            </span>
            <span className="shrink-0 text-xs text-muted-foreground">
              {timeAgo}
            </span>
            {comment.is_edited ? (
              <span className="shrink-0 text-xs text-muted-foreground">
                (edited)
              </span>
            ) : null}
          </div>
          <p className="mt-1 whitespace-pre-wrap break-words text-sm">
            {comment.content}
          </p>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <ReactionBar
              reactions={comment.reactions ?? []}
              onToggle={(emoji) => onToggleReaction(comment.id, emoji)}
              disabled={reactionPendingId === comment.id}
            />
            {canReply ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-muted-foreground hover:text-foreground"
                onClick={() => setReplyOpen((v) => !v)}
                disabled={replyPendingId === comment.id}
              >
                <ReplyIcon className="mr-1 h-3.5 w-3.5" />
                Reply
              </Button>
            ) : null}
          </div>

          {replyOpen ? (
            <div className="mt-3 space-y-2 border-l-2 border-muted pl-3">
              <Textarea
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                placeholder="Write a reply..."
                rows={2}
                className="resize-none text-sm"
                autoFocus
                onKeyDown={(e) => {
                  // Cmd/Ctrl+Enter submits; Esc cancels. Same pattern as
                  // the top-level comment form.
                  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                    e.preventDefault()
                    handleSubmitReply()
                  } else if (e.key === 'Escape') {
                    e.preventDefault()
                    setReplyOpen(false)
                    setReplyContent('')
                  }
                }}
              />
              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setReplyOpen(false)
                    setReplyContent('')
                  }}
                >
                  Cancel
                </Button>
                <Button
                  type="button"
                  size="sm"
                  onClick={handleSubmitReply}
                  disabled={
                    !replyContent.trim() || replyPendingId === comment.id
                  }
                >
                  Reply
                </Button>
              </div>
            </div>
          ) : null}

          {hasReplies ? (
            <div className="mt-3 space-y-1 border-l-2 border-muted pl-3">
              {comment.replies!.map((r) => (
                <CommentNode
                  key={r.id}
                  comment={r}
                  onReply={onReply}
                  onToggleReaction={onToggleReaction}
                  replyPendingId={replyPendingId}
                  reactionPendingId={reactionPendingId}
                  isReply
                />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}
