import * as React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  ArrowLeft,
  Copy,
  Check,
  ListTree,
  Star,
  ExternalLink,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Span, TraceDetail } from '../api/types'
import { tracesKeys, updateTraceBookmark } from '../api/queries'
import { TraceTagsEditor } from './trace-tags-editor'
import { AnnotationsDrawer } from './annotations-drawer'
import { CommentsDrawer } from './comments-drawer'
import {
  formatCost,
  formatDuration,
  formatTokens,
} from '../utils/format-helpers'

interface TraceDetailHeaderProps {
  trace: TraceDetail
  spans: Span[]
  orgId: string
  projectId: string
  className?: string
}

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1200)
    } catch {
      // clipboard permission denied — silent
    }
  }
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className="h-6 w-6"
      onClick={copy}
      aria-label="Copy trace ID"
    >
      {copied ? (
        <Check className="h-3.5 w-3.5 text-green-600" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </Button>
  )
}

function StatusBadge({ trace }: { trace: TraceDetail }) {
  if (trace.has_error || trace.status_code === 2) {
    return <Badge variant="destructive">Error</Badge>
  }
  if (trace.status_code === 1) {
    return <Badge variant="secondary">OK</Badge>
  }
  return <Badge variant="outline">Unset</Badge>
}

function MetaPair({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
        {label}
      </span>
      <span className="text-sm">{value}</span>
    </div>
  )
}

/**
 * TraceDetailHeader - Header for trace detail view.
 *
 * Page-mode only (no peek context) — back link to the traces list,
 * copy/bookmark/open-in-new-tab controls, and the embedded
 * annotations + comments drawer triggers. The tag row sits below the
 * primary control strip so the meta pairs aren't displaced when the
 * tag list wraps.
 */
export function TraceDetailHeader({
  trace,
  spans: _spans,
  orgId,
  projectId,
  className,
}: TraceDetailHeaderProps) {
  const queryClient = useQueryClient()

  const bookmarkMutation = useMutation({
    mutationFn: (bookmarked: boolean) =>
      updateTraceBookmark(projectId, trace.trace_id, bookmarked),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: tracesKeys.detail(trace.trace_id),
      })
      queryClient.invalidateQueries({ queryKey: tracesKeys.lists() })
    },
    onError: (err) => {
      toast.error('Failed to update bookmark', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const handleOpenInNewTab = () => {
    window.open(window.location.href, '_blank', 'noopener,noreferrer')
  }

  return (
    <div className={cn('border-b bg-background', className)}>
      <div className="flex items-center justify-between gap-3 px-4 py-3">
        <div className="flex min-w-0 items-center gap-2">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  asChild
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 mr-1"
                >
                  <Link
                    to="/o/$orgId/p/$projectId/traces"
                    params={{ orgId, projectId }}
                    search={{
                      page: 1,
                      limit: 20,
                      q: undefined,
                      status: undefined,
                      range: 'all',
                      model: undefined,
                    }}
                    aria-label="Back to traces"
                  >
                    <ArrowLeft className="h-4 w-4" />
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent>Back to traces</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <div className="flex items-center gap-1.5">
            <ListTree className="h-4 w-4" />
            <span className="text-sm font-medium">Trace</span>
          </div>

          <span className="truncate text-sm font-mono font-medium">
            {trace.trace_id}
          </span>
          <CopyButton value={trace.trace_id} />

          <h1 className="ml-3 truncate text-lg font-semibold">{trace.name}</h1>
          <StatusBadge trace={trace} />
        </div>

        <div className="flex shrink-0 items-center gap-1">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() =>
                    bookmarkMutation.mutate(!trace.bookmarked)
                  }
                  disabled={bookmarkMutation.isPending}
                  aria-label={
                    trace.bookmarked ? 'Remove bookmark' : 'Bookmark trace'
                  }
                >
                  <Star
                    className={cn(
                      'h-4 w-4',
                      trace.bookmarked && 'fill-yellow-400 text-yellow-400',
                    )}
                  />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {trace.bookmarked ? 'Remove bookmark' : 'Bookmark trace'}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <CommentsDrawer projectId={projectId} traceId={trace.trace_id} />
          <AnnotationsDrawer projectId={projectId} traceId={trace.trace_id} />

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleOpenInNewTab}
                  aria-label="Open in new tab"
                >
                  <ExternalLink className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Open in new tab</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-2 px-4 pb-3 text-sm">
        <MetaPair label="Duration" value={formatDuration(trace.duration)} />
        <MetaPair label="Tokens" value={formatTokens(trace.total_tokens)} />
        <MetaPair label="Cost" value={formatCost(trace.total_cost)} />
        <MetaPair label="Spans" value={trace.span_count.toLocaleString()} />
        {trace.model_name && (
          <MetaPair label="Model" value={trace.model_name} />
        )}
      </div>

      <div className="px-4 pb-3">
        <TraceTagsEditor
          projectId={projectId}
          traceId={trace.trace_id}
          tags={trace.tags || []}
        />
      </div>
    </div>
  )
}
