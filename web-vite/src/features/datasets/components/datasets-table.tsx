import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { DatasetListItem } from '../api/types'

interface DatasetsTableProps {
  rows: DatasetListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function DatasetsTable({ rows }: DatasetsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No datasets yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Create a dataset to curate inputs and expected outputs for your
          evaluations.
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
            <TableHead>Items</TableHead>
            <TableHead>Updated</TableHead>
            <TableHead>ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((dataset) => (
            <TableRow key={dataset.id}>
              <TableCell className="font-medium">{dataset.name}</TableCell>
              <TableCell className="text-muted-foreground max-w-md truncate">
                {dataset.description ?? '—'}
              </TableCell>
              <TableCell>{dataset.item_count.toLocaleString()}</TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(dataset.updated_at)}
              </TableCell>
              <TableCell className="font-mono text-xs text-muted-foreground">
                {dataset.id.slice(0, 8)}…
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
