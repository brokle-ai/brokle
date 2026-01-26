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
  Trash2,
  Play,
  Plus,
  Pause,
  Archive,
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
import { Progress } from '@/components/ui/progress'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import type { QueueWithStats, QueueStatus, AnnotationQueue } from '../types'
import type { QueueSortField } from '../hooks/use-queues-table-state'
import { QueuesToolbar } from './queues-toolbar'

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
  onSortChange: (sortBy: QueueSortField | null, sortOrder: 'asc' | 'desc' | null) => void
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
  const columns = useMemo<ColumnDef<QueueWithStats>[]>(
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
            href={`/projects/${projectSlug}/annotation-queues/${row.original.queue.id}`}
            className="font-medium text-foreground hover:text-primary hover:underline"
          >
            {row.original.queue.name}
          </Link>
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
          <QueueStatusBadge status={row.original.queue.status} />
        ),
      },
      {
        id: 'progress',
        header: 'Progress',
        cell: ({ row }) => {
          const { stats } = row.original
          const total = stats.total_items
          const completed = stats.completed_items
          const percentage = total > 0 ? Math.round((completed / total) * 100) : 0
          return (
            <div className="flex items-center gap-2 min-w-[120px]">
              <Progress value={percentage} className="h-2 w-20" />
              <span className="text-xs text-muted-foreground whitespace-nowrap">
                {percentage}%
              </span>
            </div>
          )
        },
      },
      {
        id: 'items',
        header: 'Items',
        cell: ({ row }) => {
          const { stats } = row.original
          return (
            <span className="text-sm text-muted-foreground whitespace-nowrap">
              {stats.completed_items}/{stats.total_items}
            </span>
          )
        },
      },
      {
        id: 'pending',
        header: 'Pending',
        cell: ({ row }) => {
          const pending = row.original.stats.pending_items
          return (
            <span className={cn(
              'text-sm',
              pending > 0 ? 'text-foreground' : 'text-muted-foreground'
            )}>
              {pending}
            </span>
          )
        },
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
            {formatDistanceToNow(new Date(row.original.queue.created_at), { addSuffix: true })}
          </span>
        ),
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <QueueActions
            queue={row.original.queue}
            projectSlug={projectSlug}
            onEdit={onEdit}
            onAddItems={onAddItems}
            onDelete={onDelete}
          />
        ),
      },
    ],
    [projectSlug, sortBy, sortOrder, onSortChange, onEdit, onAddItems, onDelete]
  )

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return <QueuesTableSkeleton />
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
                    <p className="text-muted-foreground">No annotation queues found</p>
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
  field: QueueSortField
  currentSort: QueueSortField
  currentOrder: 'asc' | 'desc'
  onSort: (field: QueueSortField | null, order: 'asc' | 'desc' | null) => void
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
 * Queue status badge
 */
function QueueStatusBadge({ status }: { status: QueueStatus }) {
  const config = getStatusConfig(status)
  return (
    <Badge variant="outline" className={cn('text-xs gap-1', config.className)}>
      <config.icon className="h-3 w-3" />
      {config.label}
    </Badge>
  )
}

function getStatusConfig(status: QueueStatus) {
  switch (status) {
    case 'active':
      return {
        label: 'Active',
        icon: Play,
        className: 'border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-300',
      }
    case 'paused':
      return {
        label: 'Paused',
        icon: Pause,
        className: 'border-yellow-200 bg-yellow-50 text-yellow-700 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-300',
      }
    case 'archived':
      return {
        label: 'Archived',
        icon: Archive,
        className: 'border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400',
      }
    default:
      return {
        label: status,
        icon: Play,
        className: '',
      }
  }
}

/**
 * Row actions dropdown
 */
function QueueActions({
  queue,
  projectSlug,
  onEdit,
  onAddItems,
  onDelete,
}: {
  queue: AnnotationQueue
  projectSlug: string
  onEdit?: (queue: AnnotationQueue) => void
  onAddItems?: (queue: AnnotationQueue) => void
  onDelete?: (queue: AnnotationQueue) => void
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
        <DropdownMenuItem asChild>
          <Link href={`/projects/${projectSlug}/annotation-queues/${queue.id}`}>
            <Play className="mr-2 h-4 w-4" />
            Start Annotating
          </Link>
        </DropdownMenuItem>
        {onAddItems && (
          <DropdownMenuItem onClick={() => onAddItems(queue)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Items
          </DropdownMenuItem>
        )}
        {onEdit && (
          <DropdownMenuItem onClick={() => onEdit(queue)}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </DropdownMenuItem>
        )}
        {onDelete && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => onDelete(queue)}
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
 * Loading skeleton for the table
 */
function QueuesTableSkeleton() {
  return (
    <div className="space-y-4">
      {/* Toolbar skeleton */}
      <div className="flex items-center gap-3">
        <Skeleton className="h-9 w-[200px]" />
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
    </div>
  )
}
