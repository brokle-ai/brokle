import { useState } from 'react'
import {
  ChevronUp,
  ChevronDown,
  Copy,
  Check,
  ExternalLink,
  ListTree,
  Maximize2,
  X,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { AddToQueueButton } from '@/features/annotation-queues'
import type { TraceDetail } from '../../api/types'

// ----------------------------------------------------------------------
// Formatters — mirror the TraceDetail / TracesTable rendering contract
// so peek and full-page show identical duration/cost/token shapes.
// ----------------------------------------------------------------------

function formatDuration(ns: number | undefined): string {
  if (ns === undefined || ns === null) return '—'
  if (ns < 1_000) return `${ns}ns`
  const us = ns / 1_000
  if (us < 1_000) return `${us.toFixed(1)}µs`
  const ms = us / 1_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1_000).toFixed(2)}s`
}

function formatCost(costStr: string | undefined): string {
  if (!costStr) return '—'
  const n = Number(costStr)
  if (!Number.isFinite(n)) return '—'
  if (n === 0) return '$0.00'
  if (n < 0.01) return `$${n.toFixed(6)}`
  return `$${n.toFixed(4)}`
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined || tokens === null) return '—'
  return tokens.toLocaleString()
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

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1200)
    } catch {
      // clipboard permission denied — silent (no toast) to match
      // the copy behaviour in TracesTable / TraceDetail.
    }
  }
  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-5 w-5"
      onClick={copy}
      aria-label="Copy trace ID"
    >
      {copied ? (
        <Check className="h-3 w-3 text-green-500" />
      ) : (
        <Copy className="h-3 w-3 text-muted-foreground hover:text-foreground" />
      )}
    </Button>
  )
}

interface PeekSheetHeaderProps {
  trace: TraceDetail
  orgId: string
  projectId: string
  onPrevious: () => void
  onNext: () => void
  onExpand: () => void
  onExpandNewTab: () => void
  onClose: () => void
  hasPrevious: boolean
  hasNext: boolean
  className?: string
}

// PeekSheetHeader — two-row header:
//   Row 1: trace icon + name + status badge + per-page nav + add-to-queue
//          + expand + close
//   Row 2: metadata summary (duration / tokens / cost / trace ID)
export function PeekSheetHeader({
  trace,
  orgId,
  projectId,
  onPrevious,
  onNext,
  onExpand,
  onExpandNewTab,
  onClose,
  hasPrevious,
  hasNext,
  className,
}: PeekSheetHeaderProps) {
  return (
    <div className={cn('border-b bg-background', className)}>
      <div className="flex items-center justify-between gap-2 px-4 py-3">
        <div className="flex min-w-0 items-center gap-2">
          <ListTree className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
          <span className="truncate text-sm font-medium" title={trace.name}>
            {trace.name}
          </span>
          <StatusBadge trace={trace} />
        </div>

        <div className="flex flex-shrink-0 items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onPrevious}
            disabled={!hasPrevious}
            aria-label="Previous trace"
            title="Previous trace (←)"
          >
            <ChevronUp className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onNext}
            disabled={!hasNext}
            aria-label="Next trace"
            title="Next trace (→)"
          >
            <ChevronDown className="h-4 w-4" />
          </Button>

          <div className="mx-1 h-6 w-px bg-border" />

          <AddToQueueButton
            orgId={orgId}
            projectId={projectId}
            objectId={trace.trace_id}
            objectType="trace"
          />

          <div className="mx-1 h-6 w-px bg-border" />

          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onExpand}
            aria-label="Open in current tab"
            title="Open in current tab"
          >
            <Maximize2 className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onExpandNewTab}
            aria-label="Open in new tab"
            title="Open in new tab"
          >
            <ExternalLink className="h-4 w-4" />
          </Button>

          <div className="mx-1 h-6 w-px bg-border" />

          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onClose}
            aria-label="Close peek"
            title="Close (Esc)"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-border/50 px-4 py-2 text-xs text-muted-foreground">
        <span>Duration {formatDuration(trace.duration)}</span>
        <span>Tokens {formatTokens(trace.total_tokens)}</span>
        <span>Cost {formatCost(trace.total_cost)}</span>
        <span className="flex items-center gap-1">
          <span>ID</span>
          <span className="font-mono">{trace.trace_id.slice(0, 12)}…</span>
          <CopyButton value={trace.trace_id} />
        </span>
      </div>
    </div>
  )
}
