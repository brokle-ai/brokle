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
import {
  DataTablePagination,
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import type { Score, ScoreType, ScoreSource } from '../types'
import type { Pagination } from '@/lib/api/core/types'
import { ScoresToolbar } from './scores-toolbar'
import { createScoresColumns } from './scores-columns'
import type { ScoreSortField } from '../hooks/use-scores-table-state'

interface ScoresTableProps {
  data: Score[]
  pagination: Pagination
  projectSlug: string
  loading?: boolean
  error?: string

  // URL state management
  search: string | null
  dataType: ScoreType | null
  source: ScoreSource | null
  sortBy: ScoreSortField
  sortOrder: 'asc' | 'desc'

  // State setters
  onSearchChange: (search: string) => void
  onDataTypeChange: (dataType: ScoreType | null) => void
  onSourceChange: (source: ScoreSource | null) => void
  onPageChange: (page: number, pageSize?: number) => void
  onSortChange: (
    sortBy: ScoreSortField | null,
    sortOrder: 'asc' | 'desc' | null
  ) => void
  onReset: () => void
  hasActiveFilters: boolean

  // Optional handlers
  onDeleteScore?: (scoreId: string) => void
}

/**
 * Scores Table Component
 *
 * Features:
 * - Integrated toolbar with search and filters
 * - Sortable columns with URL state
 * - Score tags with deterministic colors
 * - Pagination controls
 * - Loading and error states
 */
export function ScoresTable({
  data,
  pagination,
  projectSlug,
  loading = false,
  error,
  search,
  dataType,
  source,
  sortBy,
  sortOrder,
  onSearchChange,
  onDataTypeChange,
  onSourceChange,
  onPageChange,
  onSortChange,
  onReset,
  hasActiveFilters,
  onDeleteScore,
}: ScoresTableProps) {
  // Create columns with sorting support
  const columns = useMemo(
    () =>
      createScoresColumns({
        projectSlug,
        sortBy,
        sortOrder,
        onSortChange,
        onDeleteScore,
      }),
    [projectSlug, sortBy, sortOrder, onSortChange, onDeleteScore]
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
    return <DataTableSkeleton columns={6} rows={5} toolbarSlots={3} />
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-muted-foreground">Failed to load scores</p>
          <p className="text-destructive text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <ScoresToolbar
        search={search}
        dataType={dataType}
        source={source}
        onSearchChange={onSearchChange}
        onDataTypeChange={onDataTypeChange}
        onSourceChange={onSourceChange}
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
                    title="No scores found"
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
        pageSizes={[10, 20, 50, 100]}
        totalLabel={`${pagination.total} total score${pagination.total !== 1 ? 's' : ''}`}
      />
    </div>
  )
}
