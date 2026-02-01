'use client'

import { useState, useMemo } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Trash2, Loader2, Eye } from 'lucide-react'
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
import { DataTableSkeleton } from '@/components/data-table'
import { cn } from '@/lib/utils'
import { useDatasetItemsQuery, useDeleteDatasetItemMutation } from '../hooks/use-datasets'
import { useRowHeight } from '../hooks/use-row-height'
import { AutodetectCell, RowHeightSelector, ROW_HEIGHT_VALUES } from './cells'
import { ItemPreviewSidebar } from './dataset-items/item-preview-sidebar'
import type { DatasetItem } from '../types'
import type { RowHeight } from './cells/types'

interface DatasetItemTableProps {
  projectId: string
  datasetId: string
}

const columnHelper = createColumnHelper<DatasetItem>()

export function DatasetItemTable({ projectId, datasetId }: DatasetItemTableProps) {
  const [deleteItem, setDeleteItem] = useState<DatasetItem | null>(null)
  const [previewItem, setPreviewItem] = useState<DatasetItem | null>(null)
  const { rowHeight, setRowHeight, isLoaded } = useRowHeight()
  const { data, isLoading } = useDatasetItemsQuery(projectId, datasetId)
  const deleteMutation = useDeleteDatasetItemMutation(projectId, datasetId)

  const columns = useMemo(() => [
    columnHelper.accessor('input', {
      header: 'Input',
      cell: (info) => (
        <AutodetectCell value={info.getValue()} rowHeight={rowHeight} />
      ),
    }),
    columnHelper.accessor('expected', {
      header: 'Expected Output',
      cell: (info) => (
        <AutodetectCell value={info.getValue()} rowHeight={rowHeight} />
      ),
    }),
    columnHelper.accessor('metadata', {
      header: 'Metadata',
      cell: (info) => (
        <AutodetectCell value={info.getValue()} rowHeight={rowHeight} />
      ),
    }),
    columnHelper.accessor('created_at', {
      header: 'Created',
      cell: (info) => (
        <span className="text-sm text-muted-foreground whitespace-nowrap">
          {formatDistanceToNow(new Date(info.getValue()), { addSuffix: true })}
        </span>
      ),
    }),
    columnHelper.display({
      id: 'actions',
      header: () => <span className="sr-only">Actions</span>,
      cell: (info) => (
        <div className="flex justify-end gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              setPreviewItem(info.row.original)
            }}
            className="text-muted-foreground hover:text-foreground"
          >
            <Eye className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              setDeleteItem(info.row.original)
            }}
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    }),
  ], [rowHeight])

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

  const getRowStyle = (height: RowHeight) => {
    const minHeight = ROW_HEIGHT_VALUES[height]
    return {
      minHeight: `${minHeight}px`,
    }
  }

  if (isLoading || !isLoaded) {
    return <DataTableSkeleton columns={5} rows={5} showToolbar={false} />
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

  const totalCount = data.pagination.total

  return (
    <>
      {/* Toolbar */}
      <div className="flex items-center justify-between mb-4">
        <div className="text-sm text-muted-foreground">
          {totalCount} {totalCount === 1 ? 'item' : 'items'}
        </div>
        <RowHeightSelector value={rowHeight} onChange={setRowHeight} />
      </div>

      {/* Table */}
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
              <TableRow
                key={row.id}
                className={cn(
                  'cursor-pointer hover:bg-muted/50',
                  rowHeight === 'large' && 'align-top'
                )}
                style={getRowStyle(rowHeight)}
                onClick={() => setPreviewItem(row.original)}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    className={cn(
                      'max-w-[300px]',
                      rowHeight === 'large' && 'align-top py-4'
                    )}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {data && totalCount > data.data.length && (
        <p className="text-sm text-muted-foreground text-center mt-4">
          Showing {data.data.length} of {totalCount} items
        </p>
      )}

      {/* Preview Sidebar */}
      <ItemPreviewSidebar
        item={previewItem}
        open={!!previewItem}
        onOpenChange={(open) => !open && setPreviewItem(null)}
      />

      {/* Delete Confirmation */}
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
