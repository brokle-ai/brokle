'use client'

import Link from 'next/link'
import type { ColumnDef } from '@tanstack/react-table'
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  Play,
  Plus,
  Pause,
  Archive,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { Button } from '@/components/ui/button'
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
import { SortableColumnHeader } from '@/components/shared/tables'
import type { QueueWithStats, QueueStatus, AnnotationQueue } from '../types'
import type { QueueSortField } from '../hooks/use-queues-table-state'

interface CreateQueuesColumnsOptions {
  projectSlug: string
  sortBy: QueueSortField
  sortOrder: 'asc' | 'desc'
  onSortChange: (
    field: QueueSortField | null,
    order: 'asc' | 'desc' | null
  ) => void
  onEdit?: (queue: AnnotationQueue) => void
  onAddItems?: (queue: AnnotationQueue) => void
  onDelete?: (queue: AnnotationQueue) => void
}

export function createQueuesColumns({
  projectSlug,
  sortBy,
  sortOrder,
  onSortChange,
  onEdit,
  onAddItems,
  onDelete,
}: CreateQueuesColumnsOptions): ColumnDef<QueueWithStats>[] {
  return [
    {
      accessorKey: 'name',
      header: () => (
        <SortableColumnHeader
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
        <SortableColumnHeader
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
          <span
            className={cn(
              'text-sm',
              pending > 0 ? 'text-foreground' : 'text-muted-foreground'
            )}
          >
            {pending}
          </span>
        )
      },
    },
    {
      accessorKey: 'created_at',
      header: () => (
        <SortableColumnHeader
          label="Created"
          field="created_at"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground whitespace-nowrap">
          {formatDistanceToNow(new Date(row.original.queue.created_at), {
            addSuffix: true,
          })}
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
  ]
}

// --- Helper Components ---

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
        className:
          'border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-300',
      }
    case 'paused':
      return {
        label: 'Paused',
        icon: Pause,
        className:
          'border-yellow-200 bg-yellow-50 text-yellow-700 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-300',
      }
    case 'archived':
      return {
        label: 'Archived',
        icon: Archive,
        className:
          'border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400',
      }
    default:
      return {
        label: status,
        icon: Play,
        className: '',
      }
  }
}

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
          <Link
            href={`/projects/${projectSlug}/annotation-queues/${queue.id}`}
          >
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
