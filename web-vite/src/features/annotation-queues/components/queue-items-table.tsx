import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { QueueItem } from '../api/types'

interface QueueItemsTableProps {
  rows: QueueItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: string }) {
  if (status === 'completed') return <Badge variant="secondary">Completed</Badge>
  if (status === 'in_progress') return <Badge variant="secondary">In progress</Badge>
  if (status === 'skipped') return <Badge variant="outline">Skipped</Badge>
  return <Badge variant="outline">Pending</Badge>
}

export function QueueItemsTable({ rows }: QueueItemsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No items in this queue</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Add traces or spans to this queue to queue them for human
          annotation.
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
            <TableHead>Type</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Assignee</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((item) => (
            <TableRow key={item.id}>
              <TableCell className="font-mono text-xs">
                {item.object_id.slice(0, 12)}…
              </TableCell>
              <TableCell className="text-muted-foreground">
                {item.object_type}
              </TableCell>
              <TableCell>
                <StatusBadge status={item.status} />
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {item.annotator_user_id
                  ? `${item.annotator_user_id.slice(0, 8)}…`
                  : '—'}
              </TableCell>
              <TableCell>{item.priority}</TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(item.created_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
