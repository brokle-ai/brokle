'use client'

import * as React from 'react'
import { Clock, DollarSign, AlertTriangle, Box, Cpu, Server, Radio, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import type { Span } from '../data/schema'
import { formatDuration } from '../utils/format-helpers'

// ============================================================================
// Types
// ============================================================================

interface SpanTimelineProps {
  spans: Span[]
  onSpanSelect?: (span: Span) => void
  selectedSpanId?: string
  className?: string
}

interface TimelineSpan extends Span {
  depth: number
  startOffsetPct: number
  widthPct: number
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Build hierarchical depth map for all spans
 */
function buildDepthMap(spans: Span[]): Map<string, number> {
  const depthMap = new Map<string, number>()
  const parentMap = new Map<string, string>()

  // Build parent reference map
  spans.forEach((span) => {
    if (span.parent_span_id) {
      parentMap.set(span.span_id, span.parent_span_id)
    }
  })

  // Calculate depth for each span
  function getDepth(spanId: string): number {
    if (depthMap.has(spanId)) {
      return depthMap.get(spanId)!
    }

    const parentId = parentMap.get(spanId)
    if (!parentId) {
      depthMap.set(spanId, 0)
      return 0
    }

    const depth = getDepth(parentId) + 1
    depthMap.set(spanId, depth)
    return depth
  }

  spans.forEach((span) => getDepth(span.span_id))
  return depthMap
}

/**
 * Calculate timeline positions for each span
 */
function calculateTimelinePositions(spans: Span[]): {
  timelineSpans: TimelineSpan[]
  minTime: number
  maxTime: number
  totalDurationMs: number
} {
  if (spans.length === 0) {
    return { timelineSpans: [], minTime: 0, maxTime: 0, totalDurationMs: 0 }
  }

  const depthMap = buildDepthMap(spans)

  // Calculate time bounds
  const startTimes = spans.map((s) => new Date(s.start_time).getTime())
  const endTimes = spans
    .filter((s) => s.end_time)
    .map((s) => new Date(s.end_time!).getTime())

  const minTime = Math.min(...startTimes)
  const maxTime = Math.max(...endTimes, ...startTimes)
  const totalDurationMs = maxTime - minTime

  if (totalDurationMs === 0) {
    // All spans at the same instant
    return {
      timelineSpans: spans.map((span) => ({
        ...span,
        depth: depthMap.get(span.span_id) || 0,
        startOffsetPct: 0,
        widthPct: 100,
      })),
      minTime,
      maxTime,
      totalDurationMs: 1, // Avoid division by zero
    }
  }

  // Calculate position for each span
  const timelineSpans: TimelineSpan[] = spans.map((span) => {
    const startMs = new Date(span.start_time).getTime()
    const endMs = span.end_time
      ? new Date(span.end_time).getTime()
      : startMs + (span.duration ? span.duration / 1_000_000 : 0)

    const startOffsetPct = ((startMs - minTime) / totalDurationMs) * 100
    const durationPct = ((endMs - startMs) / totalDurationMs) * 100
    const widthPct = Math.max(1, durationPct) // Minimum 1% width for visibility

    return {
      ...span,
      depth: depthMap.get(span.span_id) || 0,
      startOffsetPct,
      widthPct,
    }
  })

  // Sort by start time, then by depth for consistent ordering
  timelineSpans.sort((a, b) => {
    const startDiff = new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
    if (startDiff !== 0) return startDiff
    return a.depth - b.depth
  })

  return { timelineSpans, minTime, maxTime, totalDurationMs }
}

/**
 * Format duration from milliseconds for time axis
 */
function formatTimeAxisLabel(ms: number): string {
  if (ms < 1) return `${Math.round(ms * 1000)}µs`
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
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
 * Get color for span bar based on status and depth
 */
function getSpanColor(span: TimelineSpan): string {
  if (span.has_error || span.status_code === 2) {
    return 'bg-red-500 dark:bg-red-600'
  }

  // Rotate colors based on depth for visual distinction
  const colors = [
    'bg-blue-500 dark:bg-blue-600',
    'bg-green-500 dark:bg-green-600',
    'bg-purple-500 dark:bg-purple-600',
    'bg-amber-500 dark:bg-amber-600',
    'bg-teal-500 dark:bg-teal-600',
    'bg-pink-500 dark:bg-pink-600',
  ]

  return colors[span.depth % colors.length]
}

// ============================================================================
// TimelineRow Component
// ============================================================================

interface TimelineRowProps {
  span: TimelineSpan
  onSpanSelect?: (span: Span) => void
  isSelected: boolean
}

function TimelineRow({ span, onSpanSelect, isSelected }: TimelineRowProps) {
  const { icon: KindIcon, label: kindLabel } = getSpanKindInfo(span.span_kind)

  return (
    <TooltipProvider>
      <div
        className={cn(
          'flex items-center gap-2 h-8 border-b border-muted/50 cursor-pointer transition-colors',
          'hover:bg-muted/30',
          isSelected && 'bg-primary/10'
        )}
        onClick={() => onSpanSelect?.(span)}
      >
        {/* Span name column (fixed width) */}
        <div
          className='flex items-center gap-1.5 min-w-[200px] max-w-[200px] px-2 text-sm truncate'
          style={{ paddingLeft: `${span.depth * 12 + 8}px` }}
        >
          {span.depth > 0 && <ChevronRight className='h-3 w-3 text-muted-foreground' />}
          <KindIcon className='h-3.5 w-3.5 flex-shrink-0 text-muted-foreground' />
          <span className='truncate font-medium'>{span.span_name}</span>
          {span.has_error && <AlertTriangle className='h-3 w-3 text-red-500 flex-shrink-0' />}
        </div>

        {/* Timeline bar column (flexible) */}
        <div className='flex-1 h-full relative px-1'>
          <Tooltip>
            <TooltipTrigger asChild>
              <div
                className={cn(
                  'absolute h-5 top-1.5 rounded-sm transition-all',
                  getSpanColor(span),
                  isSelected && 'ring-2 ring-primary ring-offset-1 ring-offset-background'
                )}
                style={{
                  left: `${span.startOffsetPct}%`,
                  width: `${span.widthPct}%`,
                  minWidth: '4px',
                }}
              />
            </TooltipTrigger>
            <TooltipContent side='top' className='max-w-xs'>
              <div className='space-y-1 text-xs'>
                <div className='font-semibold'>{span.span_name}</div>
                <div className='flex items-center gap-1 text-muted-foreground'>
                  <Clock className='h-3 w-3' />
                  {formatDuration(span.duration)}
                </div>
                {span.model_name && (
                  <div className='text-muted-foreground'>
                    Model: {span.model_name}
                  </div>
                )}
                {span.provider_name && (
                  <div className='text-muted-foreground'>
                    Provider: {span.provider_name}
                  </div>
                )}
                {(span.gen_ai_usage_input_tokens || span.gen_ai_usage_output_tokens) && (
                  <div className='text-muted-foreground'>
                    Tokens: {((span.gen_ai_usage_input_tokens || 0) + (span.gen_ai_usage_output_tokens || 0)).toLocaleString()}
                  </div>
                )}
                {span.total_cost && Number(span.total_cost) > 0 && (
                  <div className='flex items-center gap-1 text-muted-foreground'>
                    <DollarSign className='h-3 w-3' />
                    ${Number(span.total_cost).toFixed(6)}
                  </div>
                )}
              </div>
            </TooltipContent>
          </Tooltip>
        </div>

        {/* Duration column (fixed width) */}
        <div className='min-w-[70px] max-w-[70px] px-2 text-xs text-muted-foreground text-right'>
          {formatDuration(span.duration)}
        </div>
      </div>
    </TooltipProvider>
  )
}

// ============================================================================
// TimeAxis Component
// ============================================================================

interface TimeAxisProps {
  totalDurationMs: number
}

function TimeAxis({ totalDurationMs }: TimeAxisProps) {
  // Generate tick marks (roughly 5 ticks)
  const tickCount = 5
  const ticks = Array.from({ length: tickCount + 1 }, (_, i) => {
    const pct = (i / tickCount) * 100
    const timeMs = (i / tickCount) * totalDurationMs
    return { pct, label: formatTimeAxisLabel(timeMs) }
  })

  return (
    <div className='flex items-center h-6 border-b border-muted'>
      {/* Span name column header */}
      <div className='min-w-[200px] max-w-[200px] px-2 text-xs font-medium text-muted-foreground'>
        Span
      </div>

      {/* Timeline axis */}
      <div className='flex-1 relative px-1'>
        {ticks.map(({ pct, label }) => (
          <div
            key={pct}
            className='absolute top-0 bottom-0 flex flex-col items-center'
            style={{ left: `${pct}%`, transform: 'translateX(-50%)' }}
          >
            <div className='h-2 border-l border-muted-foreground/30' />
            <span className='text-[10px] text-muted-foreground whitespace-nowrap'>
              {label}
            </span>
          </div>
        ))}
      </div>

      {/* Duration column header */}
      <div className='min-w-[70px] max-w-[70px] px-2 text-xs font-medium text-muted-foreground text-right'>
        Duration
      </div>
    </div>
  )
}

// ============================================================================
// SpanTimeline Component (Main Export)
// ============================================================================

export function SpanTimeline({ spans, onSpanSelect, selectedSpanId, className }: SpanTimelineProps) {
  const { timelineSpans, totalDurationMs } = React.useMemo(
    () => calculateTimelinePositions(spans),
    [spans]
  )

  if (spans.length === 0) {
    return (
      <div className={cn('flex items-center justify-center py-8 text-muted-foreground', className)}>
        No spans available
      </div>
    )
  }

  return (
    <div className={cn('border rounded-md overflow-hidden', className)}>
      {/* Time axis header */}
      <TimeAxis totalDurationMs={totalDurationMs} />

      {/* Scrollable span rows */}
      <ScrollArea className='h-[400px]'>
        <div className='min-w-[500px]'>
          {timelineSpans.map((span) => (
            <TimelineRow
              key={span.span_id}
              span={span}
              onSpanSelect={onSpanSelect}
              isSelected={selectedSpanId === span.span_id}
            />
          ))}
        </div>
        <ScrollBar orientation='horizontal' />
      </ScrollArea>
    </div>
  )
}
