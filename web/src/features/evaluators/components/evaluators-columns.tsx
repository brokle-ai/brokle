'use client'

import Link from 'next/link'
import type { ColumnDef } from '@tanstack/react-table'
import {
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
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import { SortableColumnHeader } from '@/components/shared/tables'
import type { Evaluator, ScorerType, EvaluatorStatus } from '../types'
import type { EvaluatorSortField } from '../hooks/use-evaluators-table-state'

interface CreateEvaluatorsColumnsOptions {
  projectSlug: string
  sortBy: EvaluatorSortField
  sortOrder: 'asc' | 'desc'
  onSortChange: (field: EvaluatorSortField | null, order: 'asc' | 'desc' | null) => void
  onStatusToggle?: (evaluatorId: string, newStatus: EvaluatorStatus) => void
  onEdit?: (evaluator: Evaluator) => void
  onDuplicate?: (evaluator: Evaluator) => void
  onViewLogs?: (evaluator: Evaluator) => void
  onDelete?: (evaluator: Evaluator) => void
}

export function createEvaluatorsColumns({
  projectSlug,
  sortBy,
  sortOrder,
  onSortChange,
  onStatusToggle,
  onEdit,
  onDuplicate,
  onViewLogs,
  onDelete,
}: CreateEvaluatorsColumnsOptions): ColumnDef<Evaluator>[] {
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
      cell: ({ row }) => <ScorerTypeBadge type={row.original.scorer_type} />,
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
        <StatusToggle evaluator={row.original} onToggle={onStatusToggle} />
      ),
    },
    {
      accessorKey: 'sampling_rate',
      header: () => (
        <SortableColumnHeader
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
        <SortableColumnHeader
          label="Last Updated"
          field="updated_at"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground whitespace-nowrap">
          {formatDistanceToNow(new Date(row.original.updated_at), {
            addSuffix: true,
          })}
        </span>
      ),
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
          {formatDistanceToNow(new Date(row.original.created_at), {
            addSuffix: true,
          })}
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
  ]
}

// --- Helper Components ---

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
        className:
          'border-purple-200 bg-purple-50 text-purple-700 dark:border-purple-800 dark:bg-purple-950 dark:text-purple-300',
      }
    case 'builtin':
      return {
        label: 'Builtin',
        icon: BarChart3,
        className:
          'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300',
      }
    case 'regex':
      return {
        label: 'Regex',
        icon: Type,
        className:
          'border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-300',
      }
    default:
      return {
        label: type,
        icon: BarChart3,
        className: '',
      }
  }
}

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
        className:
          'border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-300',
      }
    case 'inactive':
      return {
        label: 'Inactive',
        className:
          'border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400',
      }
    case 'paused':
      return {
        label: 'Paused',
        className:
          'border-yellow-200 bg-yellow-50 text-yellow-700 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-300',
      }
    default:
      return {
        label: status,
        className: '',
      }
  }
}

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
