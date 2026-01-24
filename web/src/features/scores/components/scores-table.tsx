'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import type { ColumnDef } from '@tanstack/react-table'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { ExternalLink, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  ChevronLeftIcon,
  ChevronRightIcon,
} from '@radix-ui/react-icons'
import { cn } from '@/lib/utils'
import type { Score, ScoreDataType, ScoreSource } from '../types'
import type { Pagination } from '@/lib/api/core/types'
import { ScoreTag } from './score-tag'
import { ScoresToolbar } from './scores-toolbar'
import type { ScoreSortField } from '../hooks/use-scores-table-state'
import { getDataTypeIndicator, getSourceIndicator, getScoreTagClasses } from '../lib/score-colors'

interface ScoresTableProps {
  data: Score[]
  pagination: Pagination
  projectSlug: string
  loading?: boolean
  error?: string

  // URL state management
  search: string | null
  dataType: ScoreDataType | null
  source: ScoreSource | null
  sortBy: ScoreSortField
  sortOrder: 'asc' | 'desc'

  // State setters
  onSearchChange: (search: string) => void
  onDataTypeChange: (dataType: ScoreDataType | null) => void
  onSourceChange: (source: ScoreSource | null) => void
  onPageChange: (page: number, pageSize?: number) => void
  onSortChange: (sortBy: ScoreSortField | null, sortOrder: 'asc' | 'desc' | null) => void
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
  const columns = useMemo<ColumnDef<Score>[]>(
    () => [
      {
        accessorKey: 'name',
        header: ({ column }) => (
          <SortableHeader
            label="Score"
            field="name"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <ScoreTag
            score={row.original}
            onDelete={onDeleteScore}
            compact={false}
          />
        ),
      },
      {
        accessorKey: 'data_type',
        header: ({ column }) => (
          <SortableHeader
            label="Type"
            field="data_type"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => {
          const { symbol, label } = getDataTypeIndicator(row.original.data_type)
          return (
            <div className="flex items-center gap-1.5">
              <span className="text-muted-foreground font-mono">{symbol}</span>
              <span className="text-sm text-muted-foreground">{label}</span>
            </div>
          )
        },
      },
      {
        accessorKey: 'source',
        header: ({ column }) => (
          <SortableHeader
            label="Source"
            field="source"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => {
          const { label, className } = getSourceIndicator(row.original.source)
          return (
            <Badge variant="outline" className={cn('text-xs', className)}>
              {label}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'trace_id',
        header: 'Trace',
        cell: ({ row }) => {
          const traceId = row.original.trace_id
          if (!traceId) {
            return <span className="text-muted-foreground">-</span>
          }
          return (
            <Link
              href={`/projects/${projectSlug}/traces/${traceId}`}
              className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 hover:underline dark:text-blue-400"
              onClick={(e) => e.stopPropagation()}
            >
              <span className="font-mono truncate max-w-[100px]">
                {traceId.slice(0, 8)}...
              </span>
              <ExternalLink className="h-3 w-3 flex-shrink-0" />
            </Link>
          )
        },
      },
      {
        accessorKey: 'span_id',
        header: 'Span',
        cell: ({ row }) =>
          row.original.span_id ? (
            <span className="font-mono text-sm text-muted-foreground truncate max-w-[80px]">
              {row.original.span_id.slice(0, 8)}...
            </span>
          ) : (
            <span className="text-muted-foreground">-</span>
          ),
      },
      {
        accessorKey: 'timestamp',
        header: ({ column }) => (
          <SortableHeader
            label="Created"
            field="timestamp"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {formatDistanceToNow(new Date(row.original.timestamp), { addSuffix: true })}
          </span>
        ),
      },
    ],
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
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return <ScoresTableSkeleton />
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
                      : flexRender(header.column.columnDef.header, header.getContext())}
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
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  <div className="flex flex-col items-center gap-2">
                    <p className="text-muted-foreground">No scores found</p>
                    {hasActiveFilters && (
                      <Button variant="ghost" size="sm" onClick={onReset}>
                        Clear filters
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <ScoresTablePagination
        pagination={pagination}
        onPageChange={onPageChange}
      />
    </div>
  )
}

/**
 * Sortable column header component
 */
function SortableHeader({
  label,
  field,
  currentSort,
  currentOrder,
  onSort,
}: {
  label: string
  field: ScoreSortField
  currentSort: ScoreSortField
  currentOrder: 'asc' | 'desc'
  onSort: (field: ScoreSortField | null, order: 'asc' | 'desc' | null) => void
}) {
  const isActive = currentSort === field

  const handleClick = () => {
    if (!isActive) {
      // First click: sort desc
      onSort(field, 'desc')
    } else if (currentOrder === 'desc') {
      // Second click: sort asc
      onSort(field, 'asc')
    } else {
      // Third click: clear sort
      onSort(null, null)
    }
  }

  return (
    <Button
      variant="ghost"
      size="sm"
      className="-ml-3 h-8 data-[state=open]:bg-accent"
      onClick={handleClick}
    >
      <span>{label}</span>
      {isActive ? (
        currentOrder === 'desc' ? (
          <ArrowDown className="ml-1.5 h-3.5 w-3.5" />
        ) : (
          <ArrowUp className="ml-1.5 h-3.5 w-3.5" />
        )
      ) : (
        <ArrowUpDown className="ml-1.5 h-3.5 w-3.5 opacity-50" />
      )}
    </Button>
  )
}

/**
 * Pagination controls component
 */
function ScoresTablePagination({
  pagination,
  onPageChange,
}: {
  pagination: Pagination
  onPageChange: (page: number, pageSize?: number) => void
}) {
  return (
    <div className="flex items-center justify-between px-2">
      <div className="text-sm text-muted-foreground">
        {pagination.total} total score{pagination.total !== 1 ? 's' : ''}
      </div>
      <div className="flex items-center space-x-6">
        <div className="flex items-center space-x-2">
          <p className="text-sm font-medium">Rows per page</p>
          <Select
            value={String(pagination.limit)}
            onValueChange={(value) => onPageChange(1, Number(value))}
          >
            <SelectTrigger className="h-8 w-[70px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[10, 20, 50, 100].map((size) => (
                <SelectItem key={size} value={String(size)}>
                  {size}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex w-[100px] items-center justify-center text-sm font-medium">
          Page {pagination.page} of {pagination.totalPages}
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(pagination.page - 1)}
            disabled={!pagination.hasPrev}
            aria-label="Go to previous page"
          >
            <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(pagination.page + 1)}
            disabled={!pagination.hasNext}
            aria-label="Go to next page"
          >
            <ChevronRightIcon className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </div>
    </div>
  )
}

/**
 * Loading skeleton for the table
 */
function ScoresTableSkeleton() {
  return (
    <div className="space-y-4">
      {/* Toolbar skeleton */}
      <div className="flex items-center gap-3">
        <Skeleton className="h-9 w-[200px]" />
        <Skeleton className="h-9 w-[140px]" />
        <Skeleton className="h-9 w-[120px]" />
      </div>

      {/* Table skeleton */}
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {Array(6).fill(0).map((_, index) => (
                <TableHead key={index}>
                  <Skeleton className="h-6 w-20" />
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array(5).fill(0).map((_, rowIndex) => (
              <TableRow key={rowIndex}>
                {Array(6).fill(0).map((_, colIndex) => (
                  <TableCell key={colIndex}>
                    <Skeleton className="h-6 w-16" />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Pagination skeleton */}
      <div className="flex items-center justify-between px-2">
        <Skeleton className="h-5 w-24" />
        <div className="flex items-center gap-4">
          <Skeleton className="h-8 w-[130px]" />
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-8 w-20" />
        </div>
      </div>
    </div>
  )
}
