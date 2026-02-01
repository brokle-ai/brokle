'use client'

import { useMemo } from 'react'
import {
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DataTablePagination,
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import type { QueueWithStats, QueueStatus, AnnotationQueue } from '../types'
import type { QueueSortField } from '../hooks/use-queues-table-state'
import { QueuesToolbar } from './queues-toolbar'
import { createQueuesColumns } from './queues-columns'

interface QueuesTableProps {
  data: QueueWithStats[]
  projectSlug: string
  loading?: boolean
  error?: string

  // URL state management
  search: string | null
  status: QueueStatus | null
  sortBy: QueueSortField
  sortOrder: 'asc' | 'desc'

  // State setters
  onSearchChange: (search: string) => void
  onStatusChange: (status: QueueStatus | null) => void
  onSortChange: (
    sortBy: QueueSortField | null,
    sortOrder: 'asc' | 'desc' | null
  ) => void
  onReset: () => void
  hasActiveFilters: boolean

  // Actions
  onEdit?: (queue: AnnotationQueue) => void
  onAddItems?: (queue: AnnotationQueue) => void
  onDelete?: (queue: AnnotationQueue) => void
}

/**
 * Annotation Queues Table Component
 *
 * Features:
 * - Integrated toolbar with search and filters
 * - Sortable columns with URL state
 * - Progress bar showing completion
 * - Status badges
 * - Client-side pagination (data is pre-filtered)
 * - Loading and error states
 */
export function QueuesTable({
  data,
  projectSlug,
  loading = false,
  error,
  search,
  status,
  sortBy,
  sortOrder,
  onSearchChange,
  onStatusChange,
  onSortChange,
  onReset,
  hasActiveFilters,
  onEdit,
  onAddItems,
  onDelete,
}: QueuesTableProps) {
  // Create columns with sorting support
  const columns = useMemo(
    () =>
      createQueuesColumns({
        projectSlug,
        sortBy,
        sortOrder,
        onSortChange,
        onEdit,
        onAddItems,
        onDelete,
      }),
    [projectSlug, sortBy, sortOrder, onSortChange, onEdit, onAddItems, onDelete]
  )

  // Client-side pagination since data is already filtered
  const table = useReactTable({
    data,
    columns,
    initialState: {
      pagination: {
        pageSize: 10,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  })

  if (loading) {
    return <DataTableSkeleton columns={7} rows={5} toolbarSlots={2} />
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-muted-foreground">Failed to load annotation queues</p>
          <p className="text-destructive text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <QueuesToolbar
        search={search}
        status={status}
        onSearchChange={onSearchChange}
        onStatusChange={onStatusChange}
        onReset={onReset}
        hasActiveFilters={hasActiveFilters}
        isLoading={loading}
      />

      {/* Table */}
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} colSpan={header.colSpan}>
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
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  <DataTableEmptyState
                    title="No annotation queues found"
                    hasFilters={hasActiveFilters}
                    onClearFilters={onReset}
                  />
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination - NEW! Was missing before */}
      <DataTablePagination
        table={table}
        pageSizes={[10, 25, 50, 100]}
        totalLabel={`${data.length} queue${data.length !== 1 ? 's' : ''}`}
      />
    </div>
  )
}
