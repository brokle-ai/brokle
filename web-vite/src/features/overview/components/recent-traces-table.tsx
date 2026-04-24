import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import type { RecentTrace } from '../api/types'

interface RecentTracesTableProps {
  orgId: string
  projectId: string
  data: RecentTrace[]
  className?: string
}

function formatLatency(ms: number): string {
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatRelative(timestamp: string): string {
  const d = new Date(timestamp)
  if (Number.isNaN(d.getTime())) return timestamp
  const diffSecs = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diffSecs < 60) return 'just now'
  const diffMins = Math.floor(diffSecs / 60)
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return d.toLocaleDateString()
}

function truncate(s: string, max = 44): string {
  return s.length <= max ? s : `${s.slice(0, max - 1)}…`
}

export function RecentTracesTable({
  orgId,
  projectId,
  data,
  className,
}: RecentTracesTableProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">Recent traces</CardTitle>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="flex h-[200px] items-center justify-center text-sm text-muted-foreground">
            No traces yet.
          </div>
        ) : (
          <div className="space-y-1">
            {data.slice(0, 10).map((trace) => (
              <Link
                key={trace.trace_id}
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
                className="flex items-center justify-between gap-3 rounded-md px-2 py-2 hover:bg-muted/50"
              >
                <div className="flex min-w-0 flex-1 items-center gap-2">
                  <span
                    className="truncate text-sm font-medium"
                    title={trace.name}
                  >
                    {truncate(trace.name)}
                  </span>
                  <Badge
                    variant={trace.status === 'error' ? 'destructive' : 'secondary'}
                    className="shrink-0"
                  >
                    {trace.status}
                  </Badge>
                </div>
                <div className="flex shrink-0 items-center gap-3 text-xs text-muted-foreground">
                  <span>{formatLatency(trace.latency_ms)}</span>
                  <span className="w-16 text-right">
                    {formatRelative(trace.timestamp)}
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
