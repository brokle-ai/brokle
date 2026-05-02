import { useMemo, useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { ScoreTagList } from '@/features/scores'
import type { Score } from '@/features/scores/types'
import type { Span } from '../api/types'

interface SpanNode extends Span {
  children: SpanNode[]
}

interface FlatRow {
  node: SpanNode
  depth: number
  hasChildren: boolean
  isCollapsed: boolean
}

// Duration arrives in ns. Kept separate from the list-table helper so
// the tree has zero cross-component imports.
function formatDuration(ns: number | undefined): string {
  if (ns === undefined || ns === null) return '—'
  if (ns < 1_000) return `${ns}ns`
  const us = ns / 1_000
  if (us < 1_000) return `${us.toFixed(1)}µs`
  const ms = us / 1_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1_000).toFixed(2)}s`
}

// Build hierarchical tree from the flat span array by threading
// `parent_span_id`. Orphans (parent not in the batch) are promoted to
// roots so a mis-ordered response still renders.
function buildTree(spans: Span[]): SpanNode[] {
  const byId = new Map<string, SpanNode>()
  for (const s of spans) byId.set(s.span_id, { ...s, children: [] })

  const roots: SpanNode[] = []
  for (const s of spans) {
    const node = byId.get(s.span_id)
    if (!node) continue
    const parent = s.parent_span_id ? byId.get(s.parent_span_id) : undefined
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }

  const byStart = (a: SpanNode, b: SpanNode) =>
    new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
  const sortRecursive = (nodes: SpanNode[]) => {
    nodes.sort(byStart)
    for (const n of nodes) sortRecursive(n.children)
  }
  sortRecursive(roots)
  return roots
}

// Flatten tree for straightforward rendering, respecting collapsed IDs.
function flatten(roots: SpanNode[], collapsed: Set<string>): FlatRow[] {
  const out: FlatRow[] = []
  const walk = (node: SpanNode, depth: number) => {
    const isCollapsed = collapsed.has(node.span_id)
    const hasChildren = node.children.length > 0
    out.push({ node, depth, hasChildren, isCollapsed })
    if (hasChildren && !isCollapsed) {
      for (const child of node.children) walk(child, depth + 1)
    }
  }
  for (const r of roots) walk(r, 0)
  return out
}

interface SpanTreeProps {
  spans: Span[]
  selectedSpanId?: string
  onSpanSelect: (span: Span) => void
  /**
   * Scores keyed by span_id — drives the inline `<ScoreTagList>` chip
   * row on each span. Optional: when omitted, no score badges render
   * (used by callers that don't yet have scores fetched, e.g. the
   * peek sheet's collapsed view).
   */
  scoresBySpanId?: Map<string, Score[]>
}

export function SpanTree({
  spans,
  selectedSpanId,
  onSpanSelect,
  scoresBySpanId,
}: SpanTreeProps) {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const rows = useMemo(() => {
    const tree = buildTree(spans)
    return flatten(tree, collapsed)
  }, [spans, collapsed])

  if (spans.length === 0) {
    return (
      <div className="rounded-md border p-6 text-center text-sm text-muted-foreground">
        No spans available for this trace.
      </div>
    )
  }

  const toggle = (spanId: string) =>
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(spanId)) next.delete(spanId)
      else next.add(spanId)
      return next
    })

  return (
    <div className="rounded-md border divide-y">
      {rows.map(({ node, depth, hasChildren, isCollapsed }) => {
        const isSelected = node.span_id === selectedSpanId
        const statusVariant =
          node.has_error || node.status_code === 2
            ? 'destructive'
            : node.status_code === 1
              ? 'secondary'
              : 'outline'
        const statusLabel =
          node.has_error || node.status_code === 2
            ? 'Error'
            : node.status_code === 1
              ? 'OK'
              : 'Unset'

        return (
          <div
            key={node.span_id}
            role="button"
            tabIndex={0}
            aria-selected={isSelected}
            onClick={() => onSpanSelect(node)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                onSpanSelect(node)
              }
            }}
            className={cn(
              'flex cursor-pointer items-center gap-2 px-2 py-1.5 text-sm outline-none transition-colors',
              'hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring',
              isSelected && 'bg-primary/10 ring-1 ring-primary/30',
            )}
            style={{ paddingLeft: `${depth * 20 + 8}px` }}
          >
            {hasChildren ? (
              <button
                type="button"
                aria-label={isCollapsed ? 'Expand span' : 'Collapse span'}
                className="rounded p-0.5 text-muted-foreground hover:bg-muted"
                onClick={(e) => {
                  e.stopPropagation()
                  toggle(node.span_id)
                }}
              >
                <ChevronRight
                  className={cn(
                    'h-4 w-4 transition-transform',
                    !isCollapsed && 'rotate-90',
                  )}
                />
              </button>
            ) : (
              <span className="w-5" />
            )}

            <Badge variant={statusVariant} className="shrink-0">
              {statusLabel}
            </Badge>

            <span className="flex-1 truncate font-medium">{node.span_name}</span>

            {(node.model_name || node.provider_name) && (
              <span className="shrink-0 text-xs text-muted-foreground">
                {node.provider_name ?? ''}
                {node.provider_name && node.model_name ? ' / ' : ''}
                {node.model_name ?? ''}
              </span>
            )}

            <span className="w-20 shrink-0 text-right text-xs text-muted-foreground">
              {formatDuration(node.duration)}
            </span>

            {scoresBySpanId && scoresBySpanId.get(node.span_id)?.length ? (
              <div
                className="ml-2 max-w-[160px] shrink-0"
                onClick={(e) => e.stopPropagation()}
              >
                <ScoreTagList
                  scores={scoresBySpanId.get(node.span_id) ?? []}
                  maxVisible={2}
                />
              </div>
            ) : null}
          </div>
        )
      })}
    </div>
  )
}
