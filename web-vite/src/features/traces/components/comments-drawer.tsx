import { useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Loader2, MessageCircle, Send } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import {
  createTraceComment,
  createTraceCommentReply,
  toggleTraceCommentReaction,
  traceCommentsQueryOptions,
  tracesKeys,
} from '../api/queries'
import { CommentList } from './comment-list'

// Max content length matches the backend binding
// (`comment.CreateCommentRequest.Content: min=1,max=10000`).
const MAX_CONTENT_CHARS = 10_000

interface CommentsDrawerProps {
  projectId: string
  traceId: string
  className?: string
}

export function CommentsDrawer({
  projectId,
  traceId,
  className,
}: CommentsDrawerProps) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [content, setContent] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const commentsQuery = useQuery(
    traceCommentsQueryOptions(projectId, traceId),
  )

  const invalidateComments = () =>
    queryClient.invalidateQueries({
      queryKey: tracesKeys.comments(traceId, projectId),
    })

  const createMutation = useMutation({
    mutationFn: (value: string) =>
      createTraceComment(projectId, traceId, { content: value }),
    onSuccess: () => {
      setContent('')
      invalidateComments()
    },
    onError: (err) => {
      toast.error('Failed to post comment', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const replyMutation = useMutation({
    mutationFn: ({ parentId, value }: { parentId: string; value: string }) =>
      createTraceCommentReply(projectId, traceId, parentId, {
        content: value,
      }),
    onSuccess: invalidateComments,
    onError: (err) => {
      toast.error('Failed to post reply', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const reactionMutation = useMutation({
    mutationFn: ({
      commentId,
      emoji,
    }: {
      commentId: string
      emoji: string
    }) =>
      toggleTraceCommentReaction(projectId, traceId, commentId, { emoji }),
    onSuccess: invalidateComments,
    onError: (err) => {
      toast.error('Failed to toggle reaction', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const handleSubmit = () => {
    const trimmed = content.trim()
    if (!trimmed) return
    if (trimmed.length > MAX_CONTENT_CHARS) return
    createMutation.mutate(trimmed)
  }

  const total = commentsQuery.data?.total ?? 0
  const comments = commentsQuery.data?.comments ?? []
  const badgeText = total > 99 ? '99+' : String(total)

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        // Focus the textarea when the sheet opens so the reviewer can
        // start typing without an extra click. Timeout lets the sheet
        // animation settle before focus moves — keeps iOS VoiceOver
        // from reading the header twice.
        if (next) {
          window.setTimeout(() => textareaRef.current?.focus(), 100)
        }
      }}
    >
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn('relative h-8 gap-1 px-2', className)}
        >
          <MessageCircle className="h-4 w-4" />
          <span className="text-xs">Comments</span>
          {total > 0 ? (
            <Badge
              variant="secondary"
              className="ml-1 h-4 min-w-4 px-1 text-[10px]"
            >
              {badgeText}
            </Badge>
          ) : null}
        </Button>
      </SheetTrigger>
      <SheetContent
        side="right"
        className="flex w-full flex-col gap-0 p-0 sm:max-w-md lg:max-w-lg"
      >
        <SheetHeader className="border-b px-4 py-3">
          <SheetTitle className="flex items-center gap-2 text-base">
            <MessageCircle className="h-4 w-4" />
            Comments
            {total > 0 ? (
              <Badge variant="secondary">{total}</Badge>
            ) : null}
          </SheetTitle>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-1 py-3">
          <CommentList
            comments={comments}
            onReply={(parentId, value) =>
              replyMutation.mutate({ parentId, value })
            }
            onToggleReaction={(commentId, emoji) =>
              reactionMutation.mutate({ commentId, emoji })
            }
            replyPendingId={
              replyMutation.isPending
                ? replyMutation.variables?.parentId
                : undefined
            }
            reactionPendingId={
              reactionMutation.isPending
                ? reactionMutation.variables?.commentId
                : undefined
            }
            isLoading={commentsQuery.isLoading}
          />
        </div>

        <div className="border-t px-4 py-3">
          <Textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Add a comment..."
            rows={3}
            className="resize-none text-sm"
            disabled={createMutation.isPending}
            onKeyDown={(e) => {
              // Cmd/Ctrl+Enter is the standard "send" accelerator used
              // across Slack, Linear, and GitHub issue threads.
              if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                e.preventDefault()
                handleSubmit()
              }
            }}
          />
          <div className="mt-2 flex items-center justify-between">
            <span
              className={cn(
                'text-xs text-muted-foreground',
                content.length > MAX_CONTENT_CHARS && 'text-destructive',
              )}
            >
              {content.length.toLocaleString()}/
              {MAX_CONTENT_CHARS.toLocaleString()}
            </span>
            <Button
              size="sm"
              onClick={handleSubmit}
              disabled={
                !content.trim() ||
                content.length > MAX_CONTENT_CHARS ||
                createMutation.isPending
              }
            >
              {createMutation.isPending ? (
                <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />
              ) : (
                <Send className="mr-1 h-3.5 w-3.5" />
              )}
              Send
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
