'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { formatDistanceToNow } from 'date-fns'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal, Copy, Pencil, Trash2, Database, FlaskConical, History } from 'lucide-react'
import { DataTableColumnHeader } from '@/components/data-table'
import type { DatasetWithItemCount } from '../../types'

interface DatasetsColumnsOptions {
  onEdit?: (dataset: DatasetWithItemCount) => void
  onDelete?: (dataset: DatasetWithItemCount) => void
  onRunExperiment?: (dataset: DatasetWithItemCount) => void
  onViewVersions?: (dataset: DatasetWithItemCount) => void
}

export function createDatasetsColumns(options: DatasetsColumnsOptions = {}): ColumnDef<DatasetWithItemCount>[] {
  const { onEdit, onDelete, onRunExperiment, onViewVersions } = options

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
        const dataset = row.original
        return (
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <Database className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">{dataset.name}</span>
            </div>
            {dataset.description && (
              <span className="text-xs text-muted-foreground line-clamp-1 ml-6">
                {dataset.description}
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'item_count',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Items" />
      ),
      cell: ({ row }) => {
        const count = row.getValue('item_count') as number
        return (
          <Badge variant="secondary" className="font-mono">
            {count.toLocaleString()}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'current_version_id',
      header: 'Version',
      cell: ({ row }) => {
        const versionId = row.getValue('current_version_id') as string | undefined
        if (versionId) {
          return (
            <Badge variant="outline" className="text-xs">
              Pinned
            </Badge>
          )
        }
        return (
          <span className="text-xs text-muted-foreground">
            Latest
          </span>
        )
      },
      enableSorting: false,
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
      accessorKey: 'updated_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Updated" />
      ),
      cell: ({ row }) => {
        const date = row.getValue('updated_at') as string
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
        const dataset = row.original

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
                onClick={() => navigator.clipboard.writeText(dataset.id)}
              >
                <Copy className="mr-2 h-4 w-4" />
                Copy ID
              </DropdownMenuItem>
              {onRunExperiment && (
                <DropdownMenuItem onClick={() => onRunExperiment(dataset)}>
                  <FlaskConical className="mr-2 h-4 w-4" />
                  Run Experiment
                </DropdownMenuItem>
              )}
              {onViewVersions && (
                <DropdownMenuItem onClick={() => onViewVersions(dataset)}>
                  <History className="mr-2 h-4 w-4" />
                  Version History
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              {onEdit && (
                <DropdownMenuItem onClick={() => onEdit(dataset)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {onDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(dataset)}
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
