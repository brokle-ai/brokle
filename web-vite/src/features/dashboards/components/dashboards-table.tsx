import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { DashboardListItem } from '../api/types'

interface DashboardsTableProps {
  rows: DashboardListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function DashboardsTable({ rows }: DashboardsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No dashboards yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Create a dashboard to visualize traces, spans, and scores for this
          project.
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
            <TableHead>Description</TableHead>
            <TableHead>Owner</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Updated</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((d) => (
            <TableRow key={d.id}>
              <TableCell className="font-medium">{d.name}</TableCell>
              <TableCell className="text-muted-foreground">
                {d.description && d.description.length > 0 ? d.description : '—'}
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {d.created_by ? `${d.created_by.slice(0, 8)}…` : '—'}
              </TableCell>
              <TableCell>
                {d.is_locked ? (
                  <Badge variant="secondary">Locked</Badge>
                ) : (
                  <Badge variant="outline">Editable</Badge>
                )}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(d.updated_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
