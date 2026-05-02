import { useQuery } from '@tanstack/react-query'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  type UsagePeriod,
  type UsagePeriodRelative,
  usageByProjectQueryOptions,
} from '../api/queries'

interface UsageByProjectTableProps {
  orgId: string
  period: UsagePeriod
  onPeriodChange: (period: UsagePeriod) => void
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

const PRESET_OPTIONS: { value: UsagePeriodRelative; label: string }[] = [
  { value: 'current', label: 'Current period (last 30d)' },
  { value: 'previous', label: 'Previous period' },
  { value: '7d', label: 'Last 7 days' },
  { value: '14d', label: 'Last 14 days' },
  { value: '30d', label: 'Last 30 days' },
]

// Period picker drives the by-project usage query. Backend supports
// either a `time_range` enum preset or a `from`/`to` custom window —
// we expose both. Current/previous are billing-period stand-ins until
// the backend grows real periods (see queries.ts).
export function UsageByProjectTable({
  orgId,
  period,
  onPeriodChange,
}: UsageByProjectTableProps) {
  const { data, isLoading, isError, error } = useQuery(
    usageByProjectQueryOptions(orgId, period),
  )

  const isCustom = period.kind === 'custom'

  const handlePresetChange = (value: string) => {
    if (value === 'custom') {
      // Default custom window is the last 7d so the user can tweak from
      // there without typing both bounds from scratch.
      const now = Date.now()
      const day = 24 * 60 * 60 * 1000
      onPeriodChange({
        kind: 'custom',
        from: new Date(now - 7 * day).toISOString(),
        to: new Date(now).toISOString(),
      })
    } else {
      onPeriodChange({
        kind: 'relative',
        relative: value as UsagePeriodRelative,
      })
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold">Usage by project</h2>
          <p className="text-sm text-muted-foreground">
            Breakdown across projects for the selected period.
          </p>
        </div>
        <PeriodPicker
          period={period}
          onPresetChange={handlePresetChange}
          onCustomChange={(from, to) =>
            onPeriodChange({ kind: 'custom', from, to })
          }
        />
      </div>

      {isCustom && (
        <CustomDateInputs
          from={period.from}
          to={period.to}
          onChange={(from, to) =>
            onPeriodChange({ kind: 'custom', from, to })
          }
        />
      )}

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Project</TableHead>
              <TableHead className="text-right">Spans</TableHead>
              <TableHead className="text-right">Data processed</TableHead>
              <TableHead className="text-right">Scores</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              [1, 2, 3].map((i) => (
                <TableRow key={i}>
                  <TableCell>
                    <Skeleton className="h-4 w-32" />
                  </TableCell>
                  <TableCell className="text-right">
                    <Skeleton className="ml-auto h-4 w-16" />
                  </TableCell>
                  <TableCell className="text-right">
                    <Skeleton className="ml-auto h-4 w-16" />
                  </TableCell>
                  <TableCell className="text-right">
                    <Skeleton className="ml-auto h-4 w-16" />
                  </TableCell>
                </TableRow>
              ))
            ) : isError ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-destructive">
                  {error instanceof Error ? error.message : 'Failed to load'}
                </TableCell>
              </TableRow>
            ) : !data || data.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="h-24 text-center text-muted-foreground"
                >
                  No usage data for the selected period.
                </TableCell>
              </TableRow>
            ) : (
              data.map((row) => (
                <TableRow key={row.project_id ?? 'unknown'}>
                  <TableCell className="font-medium">
                    {row.project_name ?? row.project_id ?? 'Unknown'}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {formatInt(row.total_spans)}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {formatBytes(row.total_bytes)}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {formatInt(row.total_scores)}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

function PeriodPicker({
  period,
  onPresetChange,
}: {
  period: UsagePeriod
  onPresetChange: (value: string) => void
  onCustomChange: (from: string, to: string) => void
}) {
  const value = period.kind === 'custom' ? 'custom' : period.relative
  return (
    <Select value={value} onValueChange={onPresetChange}>
      <SelectTrigger className="w-[220px]">
        <SelectValue placeholder="Select period" />
      </SelectTrigger>
      <SelectContent>
        {PRESET_OPTIONS.map((opt) => (
          <SelectItem key={opt.value} value={opt.value}>
            {opt.label}
          </SelectItem>
        ))}
        <SelectItem value="custom">Custom range…</SelectItem>
      </SelectContent>
    </Select>
  )
}

function CustomDateInputs({
  from,
  to,
  onChange,
}: {
  from: string
  to: string
  onChange: (from: string, to: string) => void
}) {
  // Use date-only inputs at the UX edge (the real OTLP query is
  // RFC 3339); convert both ways at the boundary so users can pick
  // dates without juggling timezones.
  const toDateInput = (iso: string) => {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return ''
    return d.toISOString().slice(0, 10)
  }
  const fromDateInput = (date: string, kind: 'from' | 'to') => {
    if (date.length === 0) return ''
    // Anchor "from" at start of day, "to" at end of day so a single-day
    // window actually covers the whole day.
    const iso = kind === 'from' ? `${date}T00:00:00.000Z` : `${date}T23:59:59.999Z`
    return iso
  }
  return (
    <div className="flex items-end gap-3 rounded-md border bg-muted/30 p-3">
      <div className="space-y-1">
        <Label htmlFor="usage-from" className="text-xs">
          From
        </Label>
        <Input
          id="usage-from"
          type="date"
          value={toDateInput(from)}
          onChange={(e) => onChange(fromDateInput(e.target.value, 'from'), to)}
          className="w-[170px]"
        />
      </div>
      <div className="space-y-1">
        <Label htmlFor="usage-to" className="text-xs">
          To
        </Label>
        <Input
          id="usage-to"
          type="date"
          value={toDateInput(to)}
          onChange={(e) => onChange(from, fromDateInput(e.target.value, 'to'))}
          className="w-[170px]"
        />
      </div>
    </div>
  )
}
