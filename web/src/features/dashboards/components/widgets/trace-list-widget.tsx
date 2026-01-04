'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { ExternalLink, AlertCircle, CheckCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { Widget } from '../../types'

interface TraceListWidgetProps {
  widget: Widget
  data: TraceListData | null
  isLoading: boolean
  error?: string
  projectSlug?: string
}

// Backend returns flat array with these fields
interface RawTraceItem {
  trace_id: string
  name: string
  service_name?: string
  duration_nano?: number  // Backend sends nanoseconds
  duration_ms?: number    // Or milliseconds (legacy)
  start_time: string
  status_code?: string    // Backend sends status code
  has_error?: boolean     // Or pre-computed has_error
  model_name?: string
  provider_name?: string
  total_cost?: number
}

// Frontend normalized format
interface TraceListItem {
  trace_id: string
  name: string
  service_name?: string
  duration_ms: number
  start_time: string
  has_error?: boolean
  model_name?: string
  total_cost?: number
}

// Data can be wrapped object or flat array from backend
interface TraceListData {
  traces: TraceListItem[]
  total?: number
}

// Transform raw backend data to normalized format
function normalizeTraceItem(raw: RawTraceItem): TraceListItem {
  // Convert duration: nanoseconds to milliseconds if needed
  let duration_ms = raw.duration_ms ?? 0
  if (raw.duration_nano !== undefined && raw.duration_nano > 0) {
    duration_ms = raw.duration_nano / 1_000_000
  }

  // Derive has_error from status_code if not already set
  const has_error = raw.has_error ?? (raw.status_code === 'ERROR' || raw.status_code === '2')

  return {
    trace_id: raw.trace_id,
    name: raw.name,
    service_name: raw.service_name,
    duration_ms,
    start_time: raw.start_time,
    has_error,
    model_name: raw.model_name,
    total_cost: raw.total_cost,
  }
}

// Normalize data from either flat array or wrapped format
function normalizeTraceData(data: unknown): TraceListData | null {
  if (!data) return null

  // Check if it's a wrapped format with traces array
  if (typeof data === 'object' && 'traces' in (data as object)) {
    const wrapped = data as { traces: RawTraceItem[]; total?: number }
    return {
      traces: wrapped.traces.map(normalizeTraceItem),
      total: wrapped.total,
    }
  }

  // It's a flat array from the backend
  if (Array.isArray(data)) {
    return {
      traces: (data as RawTraceItem[]).map(normalizeTraceItem),
      total: data.length,
    }
  }

  return null
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
}

function formatCost(cost: number): string {
  if (cost < 0.01) return `$${cost.toFixed(4)}`
  if (cost < 1) return `$${cost.toFixed(3)}`
  return `$${cost.toFixed(2)}`
}

export function TraceListWidget({
  widget,
  data: rawData,
  isLoading,
  error,
  projectSlug,
}: TraceListWidgetProps) {
  const maxItems = (widget.config?.maxItems as number) || 5
  const showDuration = widget.config?.showDuration !== false
  const showCost = widget.config?.showCost !== false
  const showModel = widget.config?.showModel !== false
  const compact = widget.config?.compact === true

  // Normalize the data from backend format
  const data = normalizeTraceData(rawData)

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">
            <Skeleton className="h-4 w-32" />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-2">
                <Skeleton className="h-4 w-4" />
                <Skeleton className="h-4 flex-1" />
                <Skeleton className="h-4 w-16" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-destructive">{error}</p>
        </CardContent>
      </Card>
    )
  }

  if (!data || data.traces.length === 0) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No traces found</p>
        </CardContent>
      </Card>
    )
  }

  const displayedTraces = data.traces.slice(0, maxItems)

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        {widget.description && (
          <CardDescription className="text-xs">{widget.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent>
        <div className={cn('space-y-1', compact && 'space-y-0.5')}>
          {displayedTraces.map((trace) => {
            const TraceRow = (
              <div
                className={cn(
                  'flex items-center gap-2 p-2 rounded-md hover:bg-muted/50 transition-colors group',
                  compact && 'p-1.5'
                )}
              >
                {/* Status indicator */}
                <div className="shrink-0">
                  {trace.has_error ? (
                    <AlertCircle className="h-3.5 w-3.5 text-destructive" />
                  ) : (
                    <CheckCircle className="h-3.5 w-3.5 text-green-500" />
                  )}
                </div>

                {/* Trace info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span
                      className={cn(
                        'text-xs font-medium truncate',
                        compact && 'text-[11px]'
                      )}
                    >
                      {trace.name}
                    </span>
                    {showModel && trace.model_name && (
                      <Badge variant="outline" className="text-[10px] h-4 px-1">
                        {trace.model_name}
                      </Badge>
                    )}
                  </div>
                  <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                    {trace.service_name && <span>{trace.service_name}</span>}
                    <span>
                      {formatDistanceToNow(new Date(trace.start_time), { addSuffix: true })}
                    </span>
                  </div>
                </div>

                {/* Metrics */}
                <div className="flex items-center gap-3 shrink-0">
                  {showDuration && (
                    <span className="text-[11px] text-muted-foreground tabular-nums">
                      {formatDuration(trace.duration_ms)}
                    </span>
                  )}
                  {showCost && trace.total_cost != null && trace.total_cost > 0 && (
                    <span className="text-[11px] text-muted-foreground tabular-nums">
                      {formatCost(trace.total_cost)}
                    </span>
                  )}
                  {projectSlug && (
                    <ExternalLink className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                  )}
                </div>
              </div>
            )

            if (projectSlug) {
              return (
                <Link
                  key={trace.trace_id}
                  href={`/projects/${projectSlug}/traces/${trace.trace_id}`}
                  className="block"
                >
                  {TraceRow}
                </Link>
              )
            }

            return <div key={trace.trace_id}>{TraceRow}</div>
          })}
        </div>

        {data.total != null && data.total > maxItems && (
          <p className="text-xs text-muted-foreground text-center mt-2">
            Showing {maxItems} of {data.total} traces
          </p>
        )}
      </CardContent>
    </Card>
  )
}

export type { TraceListData, TraceListItem }
