'use client'

import * as React from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ChevronRight, Clock, DollarSign, AlertTriangle, Box, Cpu, Server, Radio } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import type { Span } from '../data/schema'
import { formatDuration, formatCostDetailed } from '../utils/format-helpers'
import { ItemBadge } from './item-badge'
import type { TraceDisplaySettings, ObservationLevel } from './peek-sheet/trace-settings-dropdown'

// ============================================================================
// Types
// ============================================================================

export interface SpanTreeProps {
  spans: Span[]
  onSpanSelect?: (span: Span) => void
  selectedSpanId?: string
  className?: string
  /** Display settings from parent */
  displaySettings?: TraceDisplaySettings
  /** Collapsed node IDs (controlled from parent for expand/collapse all) */
  collapsedNodes?: Set<string>
  /** Callback when a node is toggled */
  onToggleNode?: (spanId: string) => void
}

interface FlattenedSpan {
  span: Span
  depth: number
  hasChildren: boolean
  isCollapsed: boolean
  parentTotalCost?: number
  parentDuration?: number
}

// ============================================================================
// Constants
// ============================================================================

const ROW_HEIGHT = 40 // Estimated row height for virtualization
const OVERSCAN = 5 // Number of items to render outside viewport

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Build hierarchical tree from flat span array
 */
function buildSpanTree(spans: Span[]): Span[] {
  const spanMap = new Map<string, Span>()
  const childrenMap = new Map<string, Span[]>()

  spans.forEach((span) => {
    spanMap.set(span.span_id, { ...span, child_spans: [] })
    childrenMap.set(span.span_id, [])
  })

  const rootSpans: Span[] = []

  spans.forEach((span) => {
    const spanWithChildren = spanMap.get(span.span_id)!
    if (span.parent_span_id && spanMap.has(span.parent_span_id)) {
      childrenMap.get(span.parent_span_id)!.push(spanWithChildren)
    } else {
      rootSpans.push(spanWithChildren)
    }
  })

  spanMap.forEach((span, id) => {
    span.child_spans = childrenMap.get(id)
  })

  const sortByStartTime = (a: Span, b: Span) =>
    new Date(a.start_time).getTime() - new Date(b.start_time).getTime()

  rootSpans.sort(sortByStartTime)
  spanMap.forEach((span) => {
    span.child_spans?.sort(sortByStartTime)
  })

  return rootSpans
}

/**
 * Flatten tree for virtualization, respecting collapsed state
 */
function flattenTree(
  spans: Span[],
  collapsedNodes: Set<string>,
  minLevel: ObservationLevel,
  depth: number = 0,
  parentCost?: number,
  parentDuration?: number
): FlattenedSpan[] {
  const result: FlattenedSpan[] = []

  for (const span of spans) {
    // Apply level filter
    if (!passesLevelFilter(span, minLevel)) continue

    const hasChildren = Boolean(span.child_spans && span.child_spans.length > 0)
    const isCollapsed = collapsedNodes.has(span.span_id)

    result.push({
      span,
      depth,
      hasChildren,
      isCollapsed,
      parentTotalCost: parentCost,
      parentDuration: parentDuration,
    })

    // Recursively add children if not collapsed
    if (hasChildren && !isCollapsed) {
      const childItems = flattenTree(
        span.child_spans!,
        collapsedNodes,
        minLevel,
        depth + 1,
        span.total_cost ?? undefined,
        span.duration ?? undefined
      )
      result.push(...childItems)
    }
  }

  return result
}

/**
 * Check if span passes level filter
 */
function passesLevelFilter(span: Span, minLevel: ObservationLevel): boolean {
  if (minLevel === 'all') return true

  const levelPriority: Record<string, number> = {
    debug: 0,
    default: 1,
    info: 1,
    warning: 2,
    warn: 2,
    error: 3,
  }

  const spanLevel = (span.level || 'default').toLowerCase()
  const spanPriority = levelPriority[spanLevel] ?? 1
  const minPriority =
    minLevel === 'default' ? 1 : minLevel === 'warning' ? 2 : minLevel === 'error' ? 3 : 0

  return spanPriority >= minPriority
}

/**
 * Get the total duration of all spans
 */
function getTotalDuration(spans: Span[]): number {
  if (spans.length === 0) return 0

  const startTimes = spans.map((s) => new Date(s.start_time).getTime())
  const endTimes = spans.filter((s) => s.end_time).map((s) => new Date(s.end_time!).getTime())

  const minStart = Math.min(...startTimes)
  const maxEnd = Math.max(...endTimes, ...startTimes)

  return maxEnd - minStart
}

/**
 * Get span kind icon and label
 */
function getSpanKindInfo(kind: number): { icon: React.ElementType; label: string } {
  switch (kind) {
    case 0:
      return { icon: Box, label: 'Unspecified' }
    case 1:
      return { icon: Cpu, label: 'Internal' }
    case 2:
      return { icon: Server, label: 'Server' }
    case 3:
      return { icon: Radio, label: 'Client' }
    case 4:
      return { icon: Radio, label: 'Producer' }
    case 5:
      return { icon: Radio, label: 'Consumer' }
    default:
      return { icon: Box, label: 'Unknown' }
  }
}

/**
 * Get heatmap color class based on value relative to parent
 */
function getHeatmapColor(value: number | undefined, parentValue: number | undefined): string {
  if (!value || !parentValue || parentValue === 0) return ''

  const ratio = value / parentValue
  if (ratio >= 0.66) return 'text-red-600 dark:text-red-400'
  if (ratio >= 0.33) return 'text-yellow-600 dark:text-yellow-400'
  return 'text-green-600 dark:text-green-400'
}

/**
 * Get status color class
 */
function getStatusColorClass(statusCode: number): string {
  switch (statusCode) {
    case 0:
      return 'text-muted-foreground'
    case 1:
      return 'text-green-600 dark:text-green-400'
    case 2:
      return 'text-red-600 dark:text-red-400'
    default:
      return 'text-muted-foreground'
  }
}

// ============================================================================
// SpanRow Component (Virtualized Row)
// ============================================================================

interface SpanRowProps {
  item: FlattenedSpan
  onSpanSelect?: (span: Span) => void
  selectedSpanId?: string
  traceDuration: number
  displaySettings?: TraceDisplaySettings
  onToggle: () => void
}

const SpanRow = React.memo(function SpanRow({
  item,
  onSpanSelect,
  selectedSpanId,
  traceDuration,
  displaySettings,
  onToggle,
}: SpanRowProps) {
  const { span, depth, hasChildren, isCollapsed, parentTotalCost, parentDuration } = item
  const isSelected = selectedSpanId === span.span_id

  const { icon: KindIcon, label: kindLabel } = getSpanKindInfo(span.span_kind)
  const statusColor = getStatusColorClass(span.status_code)

  // Settings with defaults
  const showDuration = displaySettings?.showDuration ?? true
  const showCostTokens = displaySettings?.showCostTokens ?? true
  const colorCodeMetrics = displaySettings?.colorCodeMetrics ?? false

  // Calculate heatmap colors if enabled
  const durationColor = colorCodeMetrics
    ? getHeatmapColor(span.duration ?? undefined, parentDuration)
    : ''
  const costColor = colorCodeMetrics
    ? getHeatmapColor(span.total_cost ?? undefined, parentTotalCost)
    : ''

  // Duration bar percentage
  const durationPct =
    traceDuration > 0 && span.duration
      ? Math.max(2, Math.round(((span.duration / 1_000_000) / traceDuration) * 100))
      : 2

  return (
    <div
      className={cn(
        'flex items-center gap-2 py-1.5 px-2 rounded-md cursor-pointer transition-colors',
        'hover:bg-muted/50',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      style={{ paddingLeft: `${depth * 20 + 8}px` }}
      onClick={() => onSpanSelect?.(span)}
    >
      {/* Expand/Collapse Button */}
      {hasChildren ? (
        <Button
          variant='ghost'
          size='icon'
          className='h-5 w-5 p-0'
          onClick={(e) => {
            e.stopPropagation()
            onToggle()
          }}
        >
          <ChevronRight
            className={cn('h-4 w-4 transition-transform', !isCollapsed && 'rotate-90')}
          />
        </Button>
      ) : (
        <div className='w-5' />
      )}

      {/* Span Type Badge */}
      <ItemBadge spanType={span.span_type} showLabel={false} isSmall />

      {/* Span Kind Icon */}
      <Tooltip>
        <TooltipTrigger asChild>
          <KindIcon className={cn('h-4 w-4 flex-shrink-0', statusColor)} />
        </TooltipTrigger>
        <TooltipContent side='top' className='text-xs'>
          {kindLabel} ({span.status_code === 2 ? 'Error' : span.status_code === 1 ? 'OK' : 'Unset'})
        </TooltipContent>
      </Tooltip>

      {/* Span Name */}
      <div className='flex-1 min-w-0'>
        <div className='flex items-center gap-2'>
          <span className='font-medium text-sm truncate'>{span.span_name}</span>
          {span.has_error && (
            <AlertTriangle className='h-3.5 w-3.5 text-red-500 flex-shrink-0' />
          )}
        </div>
        {(span.model_name || span.provider_name) && (
          <div className='text-xs text-muted-foreground truncate'>
            {span.provider_name && <span>{span.provider_name}</span>}
            {span.provider_name && span.model_name && <span> / </span>}
            {span.model_name && <span>{span.model_name}</span>}
          </div>
        )}
      </div>

      {/* Tokens Badge */}
      {showCostTokens && (span.gen_ai_usage_input_tokens || span.gen_ai_usage_output_tokens) && (
        <Badge variant='outline' className='text-xs h-5 px-1.5 flex-shrink-0'>
          {((span.gen_ai_usage_input_tokens || 0) + (span.gen_ai_usage_output_tokens || 0)).toLocaleString()} tok
        </Badge>
      )}

      {/* Cost */}
      {showCostTokens && span.total_cost && span.total_cost > 0 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={cn('flex items-center gap-1 text-xs flex-shrink-0', costColor || 'text-muted-foreground')}>
              <DollarSign className='h-3 w-3' />
              {formatCostDetailed(span.total_cost)}
            </div>
          </TooltipTrigger>
          <TooltipContent side='top' className='text-xs'>
            Total Cost
          </TooltipContent>
        </Tooltip>
      )}

      {/* Duration */}
      {showDuration && (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={cn('flex items-center gap-1 text-xs flex-shrink-0 min-w-[60px] justify-end', durationColor || 'text-muted-foreground')}>
              <Clock className='h-3 w-3' />
              {formatDuration(span.duration)}
            </div>
          </TooltipTrigger>
          <TooltipContent side='top' className='text-xs'>
            Duration: {formatDuration(span.duration)}
          </TooltipContent>
        </Tooltip>
      )}

      {/* Mini duration bar */}
      {showDuration && (
        <div className='w-16 h-1.5 bg-muted rounded-full overflow-hidden flex-shrink-0'>
          <div
            className={cn(
              'h-full rounded-full',
              span.has_error ? 'bg-red-500' : colorCodeMetrics && durationColor ? 'bg-current' : 'bg-primary'
            )}
            style={{ width: `${Math.min(durationPct, 100)}%` }}
          />
        </div>
      )}
    </div>
  )
})

// ============================================================================
// SpanTree Component (Main Export)
// ============================================================================

export function SpanTree({
  spans,
  onSpanSelect,
  selectedSpanId,
  className,
  displaySettings,
  collapsedNodes: externalCollapsedNodes,
  onToggleNode,
}: SpanTreeProps) {
  // Internal collapsed state if not controlled from parent
  const [internalCollapsedNodes, setInternalCollapsedNodes] = React.useState<Set<string>>(new Set())
  const collapsedNodes = externalCollapsedNodes ?? internalCollapsedNodes

  // Ref for virtualized container
  const parentRef = React.useRef<HTMLDivElement>(null)

  // Build hierarchical tree
  const spanTree = React.useMemo(() => buildSpanTree(spans), [spans])

  // Calculate total duration
  const totalDurationMs = React.useMemo(() => getTotalDuration(spans), [spans])

  // Flatten tree for virtualization
  const flattenedItems = React.useMemo(
    () => flattenTree(spanTree, collapsedNodes, displaySettings?.minLevel ?? 'all'),
    [spanTree, collapsedNodes, displaySettings?.minLevel]
  )

  // Set up virtualizer
  const virtualizer = useVirtualizer({
    count: flattenedItems.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  })

  // Toggle node collapse
  const handleToggle = React.useCallback(
    (spanId: string) => {
      if (onToggleNode) {
        onToggleNode(spanId)
      } else {
        setInternalCollapsedNodes((prev) => {
          const next = new Set(prev)
          if (next.has(spanId)) {
            next.delete(spanId)
          } else {
            next.add(spanId)
          }
          return next
        })
      }
    },
    [onToggleNode]
  )

  if (spans.length === 0) {
    return (
      <div className={cn('flex items-center justify-center py-8 text-muted-foreground', className)}>
        No spans available
      </div>
    )
  }

  if (flattenedItems.length === 0) {
    return (
      <div className={cn('flex items-center justify-center py-8 text-muted-foreground', className)}>
        No spans match the current filter
      </div>
    )
  }

  return (
    <TooltipProvider>
      <div
        ref={parentRef}
        className={cn('h-full overflow-auto', className)}
      >
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const item = flattenedItems[virtualItem.index]
            return (
              <div
                key={item.span.span_id}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: `${virtualItem.size}px`,
                  transform: `translateY(${virtualItem.start}px)`,
                }}
              >
                <SpanRow
                  item={item}
                  onSpanSelect={onSpanSelect}
                  selectedSpanId={selectedSpanId}
                  traceDuration={totalDurationMs}
                  displaySettings={displaySettings}
                  onToggle={() => handleToggle(item.span.span_id)}
                />
              </div>
            )
          })}
        </div>
      </div>
    </TooltipProvider>
  )
}

// ============================================================================
// Utility Exports for Parent Components
// ============================================================================

/**
 * Get all span IDs in a tree (for expand/collapse all)
 */
export function getAllSpanIds(spans: Span[]): string[] {
  const ids: string[] = []
  const tree = buildSpanTree(spans)

  function collectIds(nodes: Span[]) {
    for (const node of nodes) {
      ids.push(node.span_id)
      if (node.child_spans) {
        collectIds(node.child_spans)
      }
    }
  }

  collectIds(tree)
  return ids
}

/**
 * Get only parent span IDs (spans with children)
 */
export function getParentSpanIds(spans: Span[]): string[] {
  const ids: string[] = []
  const tree = buildSpanTree(spans)

  function collectParentIds(nodes: Span[]) {
    for (const node of nodes) {
      if (node.child_spans && node.child_spans.length > 0) {
        ids.push(node.span_id)
        collectParentIds(node.child_spans)
      }
    }
  }

  collectParentIds(tree)
  return ids
}
