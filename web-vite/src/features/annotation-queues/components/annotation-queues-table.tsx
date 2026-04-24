import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { QueueStatus, QueueWithStats } from '../api/types'

interface AnnotationQueuesTableProps {
  rows: QueueWithStats[]
}

function StatusBadge({ status }: { status: QueueStatus }) {
  switch (status) {
    case 'active':
      return <Badge variant="secondary">Active</Badge>
    case 'paused':
      return <Badge variant="outline">Paused</Badge>
    case 'archived':
      return <Badge variant="destructive">Archived</Badge>
  }
}

export function AnnotationQueuesTable({ rows }: AnnotationQueuesTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No annotation queues yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Queues group traces or spans for human review and scoring.
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
            <TableHead>Status</TableHead>
            <TableHead>Total</TableHead>
            <TableHead>Pending</TableHead>
            <TableHead>In progress</TableHead>
            <TableHead>Completed</TableHead>
            <TableHead>Skipped</TableHead>
            <TableHead>Score configs</TableHead>
            <TableHead>Updated</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => {
            const { queue, stats } = row
            return (
              <TableRow key={queue.id}>
                <TableCell className="font-medium">{queue.name}</TableCell>
                <TableCell>
                  <StatusBadge status={queue.status} />
                </TableCell>
                <TableCell>{stats.total_items.toLocaleString()}</TableCell>
                <TableCell>{stats.pending_items.toLocaleString()}</TableCell>
                <TableCell>{stats.in_progress_items.toLocaleString()}</TableCell>
                <TableCell>{stats.completed_items.toLocaleString()}</TableCell>
                <TableCell>{stats.skipped_items.toLocaleString()}</TableCell>
                <TableCell className="text-muted-foreground">
                  {queue.score_config_ids.length}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {new Date(queue.updated_at).toLocaleString()}
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
