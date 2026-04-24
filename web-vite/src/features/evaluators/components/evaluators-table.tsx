import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { EvaluatorListItem } from '../api/types'

interface EvaluatorsTableProps {
  rows: EvaluatorListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: EvaluatorListItem['status'] }) {
  if (status === 'active') return <Badge variant="secondary">Active</Badge>
  if (status === 'paused') return <Badge variant="outline">Paused</Badge>
  return <Badge variant="outline">Inactive</Badge>
}

export function EvaluatorsTable({ rows }: EvaluatorsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No evaluators yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Create an evaluator to score traces or spans as they arrive.
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
            <TableHead>Trigger</TableHead>
            <TableHead>Scope</TableHead>
            <TableHead>Scorer</TableHead>
            <TableHead>Sampling</TableHead>
            <TableHead>Updated</TableHead>
            <TableHead>ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((evaluator) => (
            <TableRow key={evaluator.id}>
              <TableCell>
                <StatusBadge status={evaluator.status} />
              </TableCell>
              <TableCell className="font-medium">{evaluator.name}</TableCell>
              <TableCell className="text-muted-foreground">
                {evaluator.trigger_type}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {evaluator.target_scope}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {evaluator.scorer_type}
              </TableCell>
              <TableCell>
                {`${(evaluator.sampling_rate * 100).toFixed(0)}%`}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(evaluator.updated_at)}
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {evaluator.id.slice(0, 8)}…
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
