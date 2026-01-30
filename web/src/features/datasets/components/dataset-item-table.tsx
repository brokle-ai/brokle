'use client'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Trash2, Loader2 } from 'lucide-react'
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { useDatasetItemsQuery, useDeleteDatasetItemMutation } from '../hooks/use-datasets'
import type { DatasetItem } from '../types'

interface DatasetItemTableProps {
  projectId: string
  datasetId: string
}

const columnHelper = createColumnHelper<DatasetItem>()

function JsonCell({ value }: { value: Record<string, unknown> | undefined }) {
  if (!value) return <span className="text-muted-foreground">-</span>
  const str = JSON.stringify(value)
  const truncated = str.length > 100 ? str.slice(0, 100) + '...' : str
  return (
    <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
      {truncated}
    </code>
  )
}

export function DatasetItemTable({ projectId, datasetId }: DatasetItemTableProps) {
  const [deleteItem, setDeleteItem] = useState<DatasetItem | null>(null)
  const { data, isLoading } = useDatasetItemsQuery(projectId, datasetId)
  const deleteMutation = useDeleteDatasetItemMutation(projectId, datasetId)

  const columns = [
    columnHelper.accessor('input', {
      header: 'Input',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('expected', {
      header: 'Expected',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('metadata', {
      header: 'Metadata',
      cell: (info) => <JsonCell value={info.getValue()} />,
    }),
    columnHelper.accessor('created_at', {
      header: 'Created',
      cell: (info) => (
        <span className="text-sm text-muted-foreground">
          {formatDistanceToNow(new Date(info.getValue()), { addSuffix: true })}
        </span>
      ),
    }),
    columnHelper.display({
      id: 'actions',
      header: () => <span className="sr-only">Actions</span>,
      cell: (info) => (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setDeleteItem(info.row.original)}
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: data?.data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const handleDelete = async () => {
    if (deleteItem) {
      await deleteMutation.mutateAsync(deleteItem.id)
      setDeleteItem(null)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    )
  }

  if (!data?.data.length) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center border rounded-md">
        <p className="text-lg font-medium">No items yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Add items to this dataset to start evaluations
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

      {data && data.pagination.total > data.data.length && (
        <p className="text-sm text-muted-foreground text-center mt-4">
          Showing {data.data.length} of {data.pagination.total} items
        </p>
      )}

      <AlertDialog open={!!deleteItem} onOpenChange={(open) => !open && setDeleteItem(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Item</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this dataset item? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
