import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { UsageOverview } from '../api/types'

interface UsageTableProps {
  overview: UsageOverview
}

// Rows for the three billable dimensions. `limit` is the total free
// tier; `used = limit - remaining` (server-authoritative, no client
// arithmetic error).
interface Row {
  metric: string
  used: number
  limit: number
  unit: string
  format: 'int' | 'bytes'
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(
    units.length - 1,
    Math.floor(Math.log(Math.abs(bytes)) / Math.log(k)),
  )
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${units[i]}`
}

function formatInt(n: number): string {
  return n.toLocaleString()
}

function formatValue(n: number, fmt: 'int' | 'bytes'): string {
  return fmt === 'bytes' ? formatBytes(n) : formatInt(n)
}

function percent(used: number, limit: number): number {
  if (limit <= 0) return 0
  return Math.min(100, Math.max(0, (used / limit) * 100))
}

function UsageBar({ pct }: { pct: number }) {
  const tone =
    pct >= 100
      ? 'bg-destructive'
      : pct >= 80
        ? 'bg-yellow-500'
        : 'bg-primary'
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
      <div
        className={`h-full ${tone}`}
        style={{ width: `${Math.min(100, pct)}%` }}
      />
    </div>
  )
}

function formatDateRange(start: string, end: string): string {
  const fmt = (iso: string) => {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return iso
    return d.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }
  return `${fmt(start)} – ${fmt(end)}`
}

export function UsageTable({ overview }: UsageTableProps) {
  const rows: Row[] = [
    {
      metric: 'Spans',
      used: overview.spans,
      limit: overview.free_spans_total,
      unit: 'spans',
      format: 'int',
    },
    {
      metric: 'Data processed',
      used: overview.bytes,
      limit: overview.free_bytes_total,
      unit: 'bytes',
      format: 'bytes',
    },
    {
      metric: 'Scores',
      used: overview.scores,
      limit: overview.free_scores_total,
      unit: 'scores',
      format: 'int',
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-baseline justify-between">
        <p className="text-sm text-muted-foreground">
          Period: {formatDateRange(overview.period_start, overview.period_end)}
        </p>
        <p className="text-sm">
          Estimated cost:{' '}
          <span className="font-semibold">${overview.estimated_cost}</span>
        </p>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Metric</TableHead>
              <TableHead>Used</TableHead>
              <TableHead>Included</TableHead>
              <TableHead className="w-[240px]">Usage</TableHead>
              <TableHead className="text-right">% used</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => {
              const pct = percent(row.used, row.limit)
              return (
                <TableRow key={row.metric}>
                  <TableCell className="font-medium">{row.metric}</TableCell>
                  <TableCell>{formatValue(row.used, row.format)}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {row.limit > 0
                      ? formatValue(row.limit, row.format)
                      : 'Unlimited'}
                  </TableCell>
                  <TableCell>
                    <UsageBar pct={pct} />
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {row.limit > 0 ? `${pct.toFixed(1)}%` : '—'}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
