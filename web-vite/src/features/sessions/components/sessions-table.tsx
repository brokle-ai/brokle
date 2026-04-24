import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { SessionListItem } from '../api/types'

interface SessionsTableProps {
  rows: SessionListItem[]
}

// Duration arrives in nanoseconds (summed OTLP span durations). Format
// into the biggest sensible unit — column stays narrow without hiding
// outliers.
function formatDurationNs(ns: number | undefined): string {
  if (ns === undefined || ns === null || ns === 0) return '—'
  const ms = ns / 1_000_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  const s = ms / 1_000
  if (s < 60) return `${s.toFixed(2)}s`
  const m = s / 60
  if (m < 60) return `${m.toFixed(1)}m`
  const h = m / 60
  return `${h.toFixed(1)}h`
}

function formatCost(cost: number | undefined): string {
  if (cost === undefined || cost === null) return '—'
  if (cost === 0) return '$0.00'
  if (cost < 0.01) return `$${cost.toFixed(6)}`
  return `$${cost.toFixed(4)}`
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined || tokens === null || tokens === 0) return '—'
  return tokens.toLocaleString()
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function SessionsTable({ rows }: SessionsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No sessions yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Sessions appear here once traces carry a <code>session_id</code>{' '}
          attribute.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Session ID</TableHead>
            <TableHead>Traces</TableHead>
            <TableHead>Errors</TableHead>
            <TableHead>Users</TableHead>
            <TableHead>Duration</TableHead>
            <TableHead>Tokens</TableHead>
            <TableHead>Cost</TableHead>
            <TableHead>First seen</TableHead>
            <TableHead>Last seen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((s) => {
            const isNoSession = s.session_id === 'no-session'
            return (
              <TableRow key={s.session_id}>
                <TableCell className="font-mono text-xs">
                  {isNoSession ? (
                    <span className="italic text-muted-foreground">
                      No session
                    </span>
                  ) : (
                    <span>{s.session_id.slice(0, 24)}</span>
                  )}
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className="font-mono">
                    {s.trace_count.toLocaleString()}
                  </Badge>
                </TableCell>
                <TableCell>
                  {s.error_count > 0 ? (
                    <Badge variant="destructive">{s.error_count}</Badge>
                  ) : (
                    <span className="text-muted-foreground">0</span>
                  )}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {s.user_ids.length}
                </TableCell>
                <TableCell>{formatDurationNs(s.total_duration)}</TableCell>
                <TableCell>{formatTokens(s.total_tokens)}</TableCell>
                <TableCell>{formatCost(s.total_cost)}</TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(s.first_trace)}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(s.last_trace)}
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
