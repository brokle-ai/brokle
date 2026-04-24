import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { ScoreDataType, ScoreListItem, ScoreSource } from '../api/types'

interface ScoresTableProps {
  rows: ScoreListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

// Render a score's value using the field appropriate to its data type.
// NUMERIC → `value` formatted with locale digits.
// BOOLEAN → `value` 1.0/0.0 lifted to true/false.
// CATEGORICAL → `string_value`.
// Missing values render as "—" rather than 0 or "" so zero-the-number
// and empty-the-label don't collapse into "no value".
function formatValue(row: ScoreListItem): string {
  if (row.type === 'CATEGORICAL') {
    return row.string_value && row.string_value.length > 0
      ? row.string_value
      : '—'
  }
  if (row.type === 'BOOLEAN') {
    if (row.value === undefined || row.value === null) return '—'
    return row.value >= 0.5 ? 'true' : 'false'
  }
  if (row.value === undefined || row.value === null) return '—'
  // Numeric: 4 fractional digits max, but collapse trailing zeros.
  return Number.isInteger(row.value)
    ? row.value.toLocaleString()
    : row.value.toLocaleString(undefined, { maximumFractionDigits: 4 })
}

function TypeBadge({ type }: { type: ScoreDataType }) {
  switch (type) {
    case 'NUMERIC':
      return <Badge variant="secondary">Numeric</Badge>
    case 'BOOLEAN':
      return <Badge variant="outline">Boolean</Badge>
    case 'CATEGORICAL':
      return <Badge variant="outline">Categorical</Badge>
  }
}

function SourceBadge({ source }: { source: ScoreSource }) {
  // Variants picked for rough provenance intuition — automated (code) vs
  // learned (llm) vs reviewed (human). No semantic weight beyond that.
  switch (source) {
    case 'code':
      return <Badge variant="secondary">Code</Badge>
    case 'llm':
      return <Badge variant="secondary">LLM</Badge>
    case 'human':
      return <Badge variant="secondary">Human</Badge>
  }
}

export function ScoresTable({ rows }: ScoresTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No scores yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Scores attach quality metrics to traces and spans — emit them from
          code, an LLM evaluator, or a human annotator.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Source</TableHead>
            <TableHead>Value</TableHead>
            <TableHead>Trace</TableHead>
            <TableHead>Timestamp</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.id}>
              <TableCell className="font-medium">{row.name}</TableCell>
              <TableCell>
                <TypeBadge type={row.type} />
              </TableCell>
              <TableCell>
                <SourceBadge source={row.source} />
              </TableCell>
              <TableCell>{formatValue(row)}</TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {row.trace_id ? `${row.trace_id.slice(0, 12)}…` : '—'}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(row.timestamp)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
