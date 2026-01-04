'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { RecentTrace } from '../types'

interface RecentTracesTableProps {
  data: RecentTrace[]
  isLoading?: boolean
  error?: string | null
  className?: string
  projectSlug?: string
}

function formatLatency(ms: number): string {
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

function truncateName(name: string, maxLength: number = 40): string {
  if (name.length <= maxLength) return name
  return name.substring(0, maxLength - 3) + '...'
}

export function RecentTracesTable({
  data,
  isLoading,
  error,
  className,
  projectSlug,
}: RecentTracesTableProps) {
  const router = useRouter()

  const handleViewAll = () => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/traces`)
    }
  }

  const handleTraceClick = (traceId: string) => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/traces/${traceId}`)
    }
  }

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium">Recent Traces</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="h-5 w-16" />
                </div>
                <div className="flex items-center gap-3">
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-4 w-12" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={cn('border-destructive', className)}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium text-destructive">
            Recent Traces
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">{error}</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Recent Traces</CardTitle>
        <Button
          variant="ghost"
          size="sm"
          className="gap-1 text-xs"
          onClick={handleViewAll}
        >
          View All
          <ArrowRight className="h-3 w-3" />
        </Button>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="h-[200px] flex items-center justify-center text-muted-foreground">
            No traces yet. Start sending traces to see activity.
          </div>
        ) : (
          <div className="space-y-2">
            {data.map((trace) => (
              <div
                key={trace.trace_id}
                className="flex items-center justify-between py-2 px-2 rounded-md hover:bg-muted/50 cursor-pointer transition-colors"
                onClick={() => handleTraceClick(trace.trace_id)}
                role="button"
                tabIndex={0}
              >
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <span className="text-sm font-medium truncate" title={trace.name}>
                    {truncateName(trace.name)}
                  </span>
                  <Badge
                    variant={trace.status === 'error' ? 'destructive' : 'secondary'}
                    className="shrink-0"
                  >
                    {trace.status}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 shrink-0 text-sm text-muted-foreground">
                  <span>{formatLatency(trace.latency_ms)}</span>
                  <span className="w-16 text-right">
                    {formatRelativeTime(trace.timestamp)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
