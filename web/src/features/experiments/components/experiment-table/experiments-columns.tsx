'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { formatDistanceToNow } from 'date-fns'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal, Copy, Pencil, Trash2, FlaskConical, RefreshCw, ExternalLink } from 'lucide-react'
import { DataTableColumnHeader } from '@/components/data-table'
import { ExperimentStatusBadge } from '../experiment-status-badge'
import type { Experiment } from '../../types'

interface ExperimentsColumnsOptions {
  onView?: (experiment: Experiment) => void
  onEdit?: (experiment: Experiment) => void
  onDelete?: (experiment: Experiment) => void
  onRerun?: (experiment: Experiment) => void
}

export function createExperimentsColumns(options: ExperimentsColumnsOptions = {}): ColumnDef<Experiment>[] {
  const { onView, onEdit, onDelete, onRerun } = options

  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-[2px]"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
      cell: ({ row }) => {
        const experiment = row.original
        return (
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <FlaskConical className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">{experiment.name}</span>
            </div>
            {experiment.description && (
              <span className="text-xs text-muted-foreground line-clamp-1 ml-6">
                {experiment.description}
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const status = row.getValue('status') as Experiment['status']
        return <ExperimentStatusBadge status={status} />
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id))
      },
    },
    {
      accessorKey: 'dataset_id',
      header: 'Dataset',
      cell: ({ row }) => {
        const datasetId = row.getValue('dataset_id') as string | null
        if (!datasetId) {
          return <span className="text-xs text-muted-foreground">-</span>
        }
        return (
          <span className="text-xs font-mono text-muted-foreground truncate max-w-[100px] block">
            {datasetId.slice(0, 8)}...
          </span>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: 'started_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Started" />
      ),
      cell: ({ row }) => {
        const date = row.getValue('started_at') as string | null
        if (!date) {
          return <span className="text-sm text-muted-foreground">-</span>
        }
        return (
          <span className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(date), { addSuffix: true })}
          </span>
        )
      },
    },
    {
      accessorKey: 'completed_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Completed" />
      ),
      cell: ({ row }) => {
        const date = row.getValue('completed_at') as string | null
        if (!date) {
          return <span className="text-sm text-muted-foreground">-</span>
        }
        return (
          <span className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(date), { addSuffix: true })}
          </span>
        )
      },
    },
    {
      accessorKey: 'created_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Created" />
      ),
      cell: ({ row }) => {
        const date = row.getValue('created_at') as string
        return (
          <span className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(date), { addSuffix: true })}
          </span>
        )
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        const experiment = row.original

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
              >
                <MoreHorizontal className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[160px]">
              <DropdownMenuItem
                onClick={() => navigator.clipboard.writeText(experiment.id)}
              >
                <Copy className="mr-2 h-4 w-4" />
                Copy ID
              </DropdownMenuItem>
              {onView && (
                <DropdownMenuItem onClick={() => onView(experiment)}>
                  <ExternalLink className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
              )}
              {onRerun && experiment.status === 'completed' && (
                <DropdownMenuItem onClick={() => onRerun(experiment)}>
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Re-run
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              {onEdit && (
                <DropdownMenuItem onClick={() => onEdit(experiment)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {onDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(experiment)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]
}
