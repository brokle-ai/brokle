'use client'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { ExternalLink } from 'lucide-react'
import Link from 'next/link'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTableSkeleton } from '@/components/data-table'
import { useExperimentItemsQuery } from '../hooks/use-experiments'
import type { ExperimentItem } from '../types'

interface ExperimentItemTableProps {
  projectId: string
  projectSlug: string
  experimentId: string
}

const columnHelper = createColumnHelper<ExperimentItem>()

function JsonCell({ value }: { value: Record<string, unknown> | null | undefined }) {
  if (!value) return <span className="text-muted-foreground">-</span>
  const str = JSON.stringify(value)
  const truncated = str.length > 100 ? str.slice(0, 100) + '...' : str
  return (
    <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
      {truncated}
    </code>
  )
}

export function ExperimentItemTable({
  projectId,
  projectSlug,
  experimentId,
}: ExperimentItemTableProps) {
  const [page, setPage] = useState(0)
  const limit = 50

  const { data, isLoading } = useExperimentItemsQuery(
    projectId,
    experimentId,
    limit,
    page * limit
  )

  const columns = [
    columnHelper.accessor('trial_number', {
      header: 'Trial',
      cell: (info) => (
        <span className="font-mono text-sm">#{info.getValue()}</span>
      ),
    }),
    columnHelper.accessor('input', {
      header: 'Input',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('output', {
      header: 'Output',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('expected', {
      header: 'Expected',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('trace_id', {
      header: 'Trace',
      cell: (info) => {
        const traceId = info.getValue()
        if (!traceId) return <span className="text-muted-foreground">-</span>
        return (
          <Link
            href={`/projects/${projectSlug}/traces/${traceId}`}
            className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
          >
            <span className="font-mono">{traceId.slice(0, 8)}...</span>
            <ExternalLink className="h-3 w-3" />
          </Link>
        )
      },
    }),
    columnHelper.accessor('created_at', {
      header: 'Created',
      cell: (info) => (
        <span className="text-sm text-muted-foreground">
          {formatDistanceToNow(new Date(info.getValue()), { addSuffix: true })}
        </span>
      ),
    }),
  ]

  const table = useReactTable({
    data: data?.items ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (isLoading) {
    return <DataTableSkeleton columns={6} rows={5} showToolbar={false} />
  }

  if (!data?.items.length) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center border rounded-md">
        <p className="text-lg font-medium">No items yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Experiment items are created via the SDK using{' '}
          <code className="text-xs bg-muted px-1 py-0.5 rounded">brokle.evaluate()</code>
        </p>
      </div>
    )
  }

  return (
    <>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {data && data.total > 0 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-muted-foreground">
            Showing {page * limit + 1}-{Math.min((page + 1) * limit, data.total)} of {data.total} items
          </p>
          {data.total > limit && (
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => p + 1)}
                disabled={(page + 1) * limit >= data.total}
              >
                Next
              </Button>
            </div>
          )}
        </div>
      )}
    </>
  )
}
