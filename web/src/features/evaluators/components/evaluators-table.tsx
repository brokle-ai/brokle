'use client'

import { useMemo } from 'react'
import {
  flexRender,
  getCoreRowModel,
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
import { Button } from '@/components/ui/button'
import {
  DataTablePagination,
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import type { Evaluator, EvaluatorStatus } from '../types'
import type { Pagination } from '@/lib/api/core/types'
import { EvaluatorsToolbar } from './evaluators-toolbar'
import { createEvaluatorsColumns } from './evaluators-columns'
import type { EvaluatorSortField } from '../hooks/use-evaluators-table-state'
import type { ScorerType } from '../types'

interface EvaluatorsTableProps {
  data: Evaluator[]
  pagination: Pagination
  projectSlug: string
  loading?: boolean
  error?: string

  // URL state management
  search: string | null
  scorerType: ScorerType | null
  status: EvaluatorStatus | null
  sortBy: EvaluatorSortField
  sortOrder: 'asc' | 'desc'

  // State setters
  onSearchChange: (search: string) => void
  onScorerTypeChange: (scorerType: ScorerType | null) => void
  onStatusChange: (status: EvaluatorStatus | null) => void
  onPageChange: (page: number, pageSize?: number) => void
  onSortChange: (
    sortBy: EvaluatorSortField | null,
    sortOrder: 'asc' | 'desc' | null
  ) => void
  onReset: () => void
  hasActiveFilters: boolean

  // Actions
  onStatusToggle?: (evaluatorId: string, newStatus: EvaluatorStatus) => void
  onEdit?: (evaluator: Evaluator) => void
  onDuplicate?: (evaluator: Evaluator) => void
  onViewLogs?: (evaluator: Evaluator) => void
  onDelete?: (evaluator: Evaluator) => void
}

/**
 * Evaluators Table Component
 *
 * Features:
 * - Integrated toolbar with search and filters
 * - Sortable columns with URL state
 * - Inline status toggle (Opik pattern)
 * - Scorer type badges with icons
 * - Pagination controls
 * - Loading and error states
 */
export function EvaluatorsTable({
  data,
  pagination,
  projectSlug,
  loading = false,
  error,
  search,
  scorerType,
  status,
  sortBy,
  sortOrder,
  onSearchChange,
  onScorerTypeChange,
  onStatusChange,
  onPageChange,
  onSortChange,
  onReset,
  hasActiveFilters,
  onStatusToggle,
  onEdit,
  onDuplicate,
  onViewLogs,
  onDelete,
}: EvaluatorsTableProps) {
  // Create columns with sorting support
  const columns = useMemo(
    () =>
      createEvaluatorsColumns({
        projectSlug,
        sortBy,
        sortOrder,
        onSortChange,
        onStatusToggle,
        onEdit,
        onDuplicate,
        onViewLogs,
        onDelete,
      }),
    [
      projectSlug,
      sortBy,
      sortOrder,
      onSortChange,
      onStatusToggle,
      onEdit,
      onDuplicate,
      onViewLogs,
      onDelete,
    ]
  )

  const table = useReactTable({
    data,
    columns,
    pageCount: pagination.totalPages,
    state: {
      pagination: {
        pageIndex: pagination.page - 1,
        pageSize: pagination.limit,
      },
    },
    manualPagination: true,
    onPaginationChange: (updater) => {
      const newPagination =
        typeof updater === 'function'
          ? updater({
              pageIndex: pagination.page - 1,
              pageSize: pagination.limit,
            })
          : updater
      onPageChange(newPagination.pageIndex + 1, newPagination.pageSize)
    },
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return <DataTableSkeleton columns={7} rows={5} toolbarSlots={3} />
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-muted-foreground">Failed to load evaluators</p>
          <p className="text-destructive text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <EvaluatorsToolbar
        search={search}
        scorerType={scorerType}
        status={status}
        onSearchChange={onSearchChange}
        onScorerTypeChange={onScorerTypeChange}
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
                    title="No evaluators found"
                    hasFilters={hasActiveFilters}
                    onClearFilters={onReset}
                  />
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <DataTablePagination
        table={table}
        pageSizes={[10, 25, 50, 100]}
        totalLabel={`${pagination.total} evaluator${pagination.total !== 1 ? 's' : ''}`}
      />
    </div>
  )
}
