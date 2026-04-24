import type { ReactNode } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { TraceListItem } from '../api/types'

interface TracesTableProps {
  rows: TraceListItem[]
  // Render a link for the trace name column. The parent owns routing
  // (TanStack Router `<Link>` with typed params), the table stays
  // framework-agnostic.
  renderNameLink?: (trace: TraceListItem, children: ReactNode) => ReactNode
}

// Duration arrives in nanoseconds (OTLP spec). Format into the biggest
// sensible unit so the column stays narrow without hiding outliers.
function formatDuration(ns: number | undefined): string {
  if (ns === undefined || ns === null) return '—'
  if (ns < 1_000) return `${ns}ns`
  const us = ns / 1_000
  if (us < 1_000) return `${us.toFixed(1)}µs`
  const ms = us / 1_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  const s = ms / 1_000
  return `${s.toFixed(2)}s`
}

// Cost arrives as a shopspring/decimal string (e.g. "0.000123"). Parse
// defensively — a bad string renders as "—" rather than "NaN".
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

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ trace }: { trace: TraceListItem }) {
  if (trace.has_error || trace.status_code === 2) {
    return <Badge variant="destructive">Error</Badge>
  }
  if (trace.status_code === 1) {
    return <Badge variant="secondary">OK</Badge>
  }
  return <Badge variant="outline">Unset</Badge>
}

export function TracesTable({ rows, renderNameLink }: TracesTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No traces yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Traces will appear here once your application sends telemetry to
          Brokle.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Model</TableHead>
            <TableHead>Provider</TableHead>
            <TableHead>Duration</TableHead>
            <TableHead>Tokens</TableHead>
            <TableHead>Cost</TableHead>
            <TableHead>Start time</TableHead>
            <TableHead>Trace ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((trace) => {
            const nameCell = renderNameLink
              ? renderNameLink(trace, trace.name)
              : trace.name
            return (
              <TableRow key={trace.trace_id}>
                <TableCell>
                  <StatusBadge trace={trace} />
                </TableCell>
                <TableCell className="font-medium">{nameCell}</TableCell>
                <TableCell className="text-muted-foreground">
                  {trace.model_name ?? '—'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {trace.provider_name ?? '—'}
                </TableCell>
                <TableCell>{formatDuration(trace.duration)}</TableCell>
                <TableCell>{formatTokens(trace.total_tokens)}</TableCell>
                <TableCell>{formatCost(trace.total_cost)}</TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(trace.start_time)}
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {trace.trace_id.slice(0, 12)}…
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
