import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { DatasetItem } from '../api/types'

interface DatasetItemsTableProps {
  rows: DatasetItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

// Stringify an arbitrary JSON value for the compact cell view. We keep
// it tight (no indentation) so the column width remains predictable;
// the `title` attribute holds the pretty form for hover inspection.
function stringifyCompact(value: unknown): string {
  if (value === null || value === undefined) return '—'
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

function stringifyPretty(value: unknown): string {
  if (value === null || value === undefined) return ''
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

export function DatasetItemsTable({ rows }: DatasetItemsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No items yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Items added to this dataset will appear here.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Input</TableHead>
            <TableHead>Expected</TableHead>
            <TableHead>Source</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((item) => (
            <TableRow key={item.id}>
              <TableCell
                className="font-mono text-xs max-w-md truncate"
                title={stringifyPretty(item.input)}
              >
                {stringifyCompact(item.input)}
              </TableCell>
              <TableCell
                className="font-mono text-xs max-w-md truncate text-muted-foreground"
                title={stringifyPretty(item.expected)}
              >
                {item.expected === undefined
                  ? '—'
                  : stringifyCompact(item.expected)}
              </TableCell>
              <TableCell>
                <Badge variant="outline">{item.source}</Badge>
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(item.created_at)}
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {item.id.slice(0, 8)}…
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
