import { AlertCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import type { TopError } from '../api/types'

interface TopErrorsTableProps {
  data: TopError[]
  className?: string
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

function formatCount(count: number): string {
  if (count < 1000) return count.toString()
  if (count < 10_000) return `${(count / 1000).toFixed(1)}k`
  return `${Math.round(count / 1000)}k`
}

function truncate(s: string, max = 60): string {
  return s.length <= max ? s : `${s.slice(0, max - 1)}…`
}

export function TopErrorsTable({ data, className }: TopErrorsTableProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">Top errors</CardTitle>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="flex h-[200px] flex-col items-center justify-center gap-2 text-sm text-muted-foreground">
            <AlertCircle className="h-8 w-8 text-emerald-500" />
            <span>No errors detected.</span>
          </div>
        ) : (
          <div className="space-y-1">
            {data.map((err, i) => (
              <div
                key={`${err.message}-${i}`}
                className="flex items-center justify-between gap-3 rounded-md px-2 py-2 hover:bg-muted/50"
              >
                <div className="flex min-w-0 flex-1 items-center gap-2">
                  <AlertCircle className="h-4 w-4 shrink-0 text-destructive" />
                  <span
                    className="truncate text-sm font-medium text-destructive"
                    title={err.message}
                  >
                    {truncate(err.message)}
                  </span>
                </div>
                <div className="flex shrink-0 items-center gap-3 text-xs text-muted-foreground">
                  <Badge variant="secondary">{formatCount(err.count)}×</Badge>
                  <span className="w-16 text-right">
                    {formatRelative(err.last_seen)}
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
