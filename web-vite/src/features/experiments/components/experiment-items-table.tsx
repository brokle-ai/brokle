import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { ExperimentItem } from '../api/types'

interface ExperimentItemsTableProps {
  rows: ExperimentItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function ExperimentItemsTable({ rows }: ExperimentItemsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No runs yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Experiment items will appear here as the run completes.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Trace ID</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Trial</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Item ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((item) => (
            <TableRow key={item.id}>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {item.trace_id ? `${item.trace_id.slice(0, 12)}…` : '—'}
              </TableCell>
              <TableCell>
                {item.error ? (
                  <Badge variant="destructive">Error</Badge>
                ) : item.output !== undefined && item.output !== null ? (
                  <Badge variant="secondary">Completed</Badge>
                ) : (
                  <Badge variant="outline">Pending</Badge>
                )}
              </TableCell>
              <TableCell>{item.trial_number}</TableCell>
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
