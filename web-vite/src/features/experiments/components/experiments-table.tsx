import { Link } from '@tanstack/react-router'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { ExperimentListItem, ExperimentStatus } from '../api/types'

interface ExperimentsTableProps {
  rows: ExperimentListItem[]
  orgId: string
  projectId: string
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: ExperimentStatus }) {
  if (status === 'failed') return <Badge variant="destructive">Failed</Badge>
  if (status === 'completed') return <Badge variant="secondary">Completed</Badge>
  if (status === 'running') return <Badge variant="secondary">Running</Badge>
  if (status === 'cancelled') return <Badge variant="outline">Cancelled</Badge>
  return <Badge variant="outline">Draft</Badge>
}

function progressLabel(item: ExperimentListItem): string {
  if (item.total_items === 0) return `${item.completed_items} / 0`
  const pct = Math.round((item.completed_items / item.total_items) * 100)
  return `${item.completed_items} / ${item.total_items} (${pct}%)`
}

export function ExperimentsTable({ rows, orgId, projectId }: ExperimentsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No experiments yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Run an experiment to compare evaluator outputs against a dataset.
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
            <TableHead>Source</TableHead>
            <TableHead>Progress</TableHead>
            <TableHead>Failed</TableHead>
            <TableHead>Started</TableHead>
            <TableHead>Completed</TableHead>
            <TableHead>ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((experiment) => (
            <TableRow key={experiment.id}>
              <TableCell>
                <StatusBadge status={experiment.status} />
              </TableCell>
              <TableCell className="font-medium">
                <Link
                  to="/o/$orgId/p/$projectId/experiments/$experimentId"
                  params={{ orgId, projectId, experimentId: experiment.id }}
                  search={{ page: 1, limit: 20 }}
                  className="hover:underline"
                >
                  {experiment.name}
                </Link>
              </TableCell>
              <TableCell className="text-muted-foreground">
                {experiment.source}
              </TableCell>
              <TableCell>{progressLabel(experiment)}</TableCell>
              <TableCell
                className={
                  experiment.failed_items > 0
                    ? 'text-destructive'
                    : 'text-muted-foreground'
                }
              >
                {experiment.failed_items}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(experiment.started_at)}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(experiment.completed_at)}
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {experiment.id.slice(0, 8)}…
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
