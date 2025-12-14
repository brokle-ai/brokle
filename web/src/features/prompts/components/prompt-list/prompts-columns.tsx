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
import { MoreHorizontal, Copy, Pencil, Trash2, Play, History } from 'lucide-react'
import type { PromptListItem } from '../../types'
import { LabelList } from '../label-badge'

interface PromptsColumnsOptions {
  onEdit?: (prompt: PromptListItem) => void
  onDelete?: (prompt: PromptListItem) => void
  onPlayground?: (prompt: PromptListItem) => void
  onViewHistory?: (prompt: PromptListItem) => void
  protectedLabels?: string[]
}

export function createPromptsColumns(options: PromptsColumnsOptions = {}): ColumnDef<PromptListItem>[] {
  const { onEdit, onDelete, onPlayground, onViewHistory, protectedLabels = [] } = options

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
      header: 'Name',
      cell: ({ row }) => {
        const prompt = row.original
        return (
          <div className="flex flex-col gap-1">
            <span className="font-medium">{prompt.name}</span>
            {prompt.description && (
              <span className="text-xs text-muted-foreground line-clamp-1">
                {prompt.description}
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'type',
      header: 'Type',
      cell: ({ row }) => {
        const type = row.getValue('type') as string
        return (
          <Badge variant={type === 'chat' ? 'default' : 'secondary'}>
            {type}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'labels',
      header: 'Labels',
      cell: ({ row }) => {
        const labels = row.original.labels
        return (
          <LabelList
            labels={labels}
            protectedLabels={protectedLabels}
            maxVisible={2}
          />
        )
      },
    },
    {
      accessorKey: 'latest_version',
      header: 'Version',
      cell: ({ row }) => {
        const version = row.getValue('latest_version') as number
        return <span className="font-mono text-sm">v{version}</span>
      },
    },
    {
      accessorKey: 'tags',
      header: 'Tags',
      cell: ({ row }) => {
        const tags = row.getValue('tags') as string[]
        if (!tags || tags.length === 0) return null
        return (
          <div className="flex flex-wrap gap-1">
            {tags.slice(0, 2).map((tag) => (
              <Badge key={tag} variant="outline" className="text-xs">
                {tag}
              </Badge>
            ))}
            {tags.length > 2 && (
              <Badge variant="outline" className="text-xs">
                +{tags.length - 2}
              </Badge>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'updated_at',
      header: 'Updated',
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
        const prompt = row.original

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
                onClick={() => navigator.clipboard.writeText(prompt.id)}
              >
                <Copy className="mr-2 h-4 w-4" />
                Copy ID
              </DropdownMenuItem>
              {onPlayground && (
                <DropdownMenuItem onClick={() => onPlayground(prompt)}>
                  <Play className="mr-2 h-4 w-4" />
                  Playground
                </DropdownMenuItem>
              )}
              {onViewHistory && (
                <DropdownMenuItem onClick={() => onViewHistory(prompt)}>
                  <History className="mr-2 h-4 w-4" />
                  History
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              {onEdit && (
                <DropdownMenuItem onClick={() => onEdit(prompt)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {onDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(prompt)}
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
