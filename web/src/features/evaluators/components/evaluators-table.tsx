'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import type { ColumnDef } from '@tanstack/react-table'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  MoreHorizontal,
  Pencil,
  Copy,
  FileText,
  Trash2,
  Bot,
  BarChart3,
  Type,
} from 'lucide-react'
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
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  ChevronLeftIcon,
  ChevronRightIcon,
} from '@radix-ui/react-icons'
import { cn } from '@/lib/utils'
import type { Evaluator, ScorerType, EvaluatorStatus } from '../types'
import type { Pagination } from '@/lib/api/core/types'
import { EvaluatorsToolbar } from './evaluators-toolbar'
import type { EvaluatorSortField } from '../hooks/use-evaluators-table-state'

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
  onSortChange: (sortBy: EvaluatorSortField | null, sortOrder: 'asc' | 'desc' | null) => void
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
  const columns = useMemo<ColumnDef<Evaluator>[]>(
    () => [
      {
        accessorKey: 'name',
        header: () => (
          <SortableHeader
            label="Name"
            field="name"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <Link
            href={`/projects/${projectSlug}/evaluators/${row.original.id}`}
            className="font-medium text-foreground hover:text-primary hover:underline"
          >
            {row.original.name}
          </Link>
        ),
      },
      {
        accessorKey: 'scorer_type',
        header: 'Type',
        cell: ({ row }) => (
          <ScorerTypeBadge type={row.original.scorer_type} />
        ),
      },
      {
        accessorKey: 'status',
        header: () => (
          <SortableHeader
            label="Status"
            field="status"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <StatusToggle
            evaluator={row.original}
            onToggle={onStatusToggle}
          />
        ),
      },
      {
        accessorKey: 'sampling_rate',
        header: () => (
          <SortableHeader
            label="Sample Rate"
            field="sampling_rate"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {Math.round(row.original.sampling_rate * 100)}%
          </span>
        ),
      },
      {
        accessorKey: 'updated_at',
        header: () => (
          <SortableHeader
            label="Last Updated"
            field="updated_at"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {formatDistanceToNow(new Date(row.original.updated_at), { addSuffix: true })}
          </span>
        ),
      },
      {
        accessorKey: 'created_at',
        header: () => (
          <SortableHeader
            label="Created"
            field="created_at"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSortChange}
          />
        ),
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {formatDistanceToNow(new Date(row.original.created_at), { addSuffix: true })}
          </span>
        ),
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <EvaluatorActions
            evaluator={row.original}
            onEdit={onEdit}
            onDuplicate={onDuplicate}
            onViewLogs={onViewLogs}
            onDelete={onDelete}
          />
        ),
      },
    ],
    [projectSlug, sortBy, sortOrder, onSortChange, onStatusToggle, onEdit, onDuplicate, onViewLogs, onDelete]
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
    return <EvaluatorsTableSkeleton />
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
                    <p className="text-muted-foreground">No evaluators found</p>
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
      <EvaluatorsTablePagination
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
  field: EvaluatorSortField
  currentSort: EvaluatorSortField
  currentOrder: 'asc' | 'desc'
  onSort: (field: EvaluatorSortField | null, order: 'asc' | 'desc' | null) => void
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
 * Scorer type badge with icon
 */
function ScorerTypeBadge({ type }: { type: ScorerType }) {
  const config = getScorerTypeConfig(type)
  return (
    <Badge variant="outline" className={cn('text-xs gap-1', config.className)}>
      <config.icon className="h-3 w-3" />
      {config.label}
    </Badge>
  )
}

function getScorerTypeConfig(type: ScorerType) {
  switch (type) {
    case 'llm':
      return {
        label: 'LLM',
        icon: Bot,
        className: 'border-purple-200 bg-purple-50 text-purple-700 dark:border-purple-800 dark:bg-purple-950 dark:text-purple-300',
      }
    case 'builtin':
      return {
        label: 'Builtin',
        icon: BarChart3,
        className: 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300',
      }
    case 'regex':
      return {
        label: 'Regex',
        icon: Type,
        className: 'border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-300',
      }
    default:
      return {
        label: type,
        icon: BarChart3,
        className: '',
      }
  }
}

/**
 * Inline status toggle (Opik pattern)
 */
function StatusToggle({
  evaluator,
  onToggle,
}: {
  evaluator: Evaluator
  onToggle?: (evaluatorId: string, newStatus: EvaluatorStatus) => void
}) {
  const isActive = evaluator.status === 'active'

  const handleToggle = () => {
    if (onToggle) {
      onToggle(evaluator.id, isActive ? 'inactive' : 'active')
    }
  }

  return (
    <div className="flex items-center gap-2">
      <Switch
        checked={isActive}
        onCheckedChange={handleToggle}
        disabled={!onToggle || evaluator.status === 'paused'}
        aria-label={`Toggle evaluator ${evaluator.name} ${isActive ? 'off' : 'on'}`}
      />
      <StatusBadge status={evaluator.status} />
    </div>
  )
}

function StatusBadge({ status }: { status: EvaluatorStatus }) {
  const config = getStatusConfig(status)
  return (
    <Badge variant="outline" className={cn('text-xs', config.className)}>
      {config.label}
    </Badge>
  )
}

function getStatusConfig(status: EvaluatorStatus) {
  switch (status) {
    case 'active':
      return {
        label: 'Active',
        className: 'border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-300',
      }
    case 'inactive':
      return {
        label: 'Inactive',
        className: 'border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400',
      }
    case 'paused':
      return {
        label: 'Paused',
        className: 'border-yellow-200 bg-yellow-50 text-yellow-700 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-300',
      }
    default:
      return {
        label: status,
        className: '',
      }
  }
}

/**
 * Row actions dropdown
 */
function EvaluatorActions({
  evaluator,
  onEdit,
  onDuplicate,
  onViewLogs,
  onDelete,
}: {
  evaluator: Evaluator
  onEdit?: (evaluator: Evaluator) => void
  onDuplicate?: (evaluator: Evaluator) => void
  onViewLogs?: (evaluator: Evaluator) => void
  onDelete?: (evaluator: Evaluator) => void
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8">
          <MoreHorizontal className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {onEdit && (
          <DropdownMenuItem onClick={() => onEdit(evaluator)}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </DropdownMenuItem>
        )}
        {onDuplicate && (
          <DropdownMenuItem onClick={() => onDuplicate(evaluator)}>
            <Copy className="mr-2 h-4 w-4" />
            Duplicate
          </DropdownMenuItem>
        )}
        {onViewLogs && (
          <DropdownMenuItem onClick={() => onViewLogs(evaluator)}>
            <FileText className="mr-2 h-4 w-4" />
            View Logs
          </DropdownMenuItem>
        )}
        {onDelete && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => onDelete(evaluator)}
              className="text-destructive focus:text-destructive"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

/**
 * Pagination controls component
 */
function EvaluatorsTablePagination({
  pagination,
  onPageChange,
}: {
  pagination: Pagination
  onPageChange: (page: number, pageSize?: number) => void
}) {
  return (
    <div className="flex items-center justify-between px-2">
      <div className="text-sm text-muted-foreground">
        {pagination.total} evaluator{pagination.total !== 1 ? 's' : ''}
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
              {[10, 25, 50, 100].map((size) => (
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
function EvaluatorsTableSkeleton() {
  return (
    <div className="space-y-4">
      {/* Toolbar skeleton */}
      <div className="flex items-center gap-3">
        <Skeleton className="h-9 w-[200px]" />
        <Skeleton className="h-9 w-[140px]" />
        <Skeleton className="h-9 w-[130px]" />
      </div>

      {/* Table skeleton */}
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {Array(7).fill(0).map((_, index) => (
                <TableHead key={index}>
                  <Skeleton className="h-6 w-20" />
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array(5).fill(0).map((_, rowIndex) => (
              <TableRow key={rowIndex}>
                {Array(7).fill(0).map((_, colIndex) => (
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
