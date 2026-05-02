import type { ColumnDef } from '@tanstack/react-table'
import { Loader2, Pencil, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { ScoreConfig } from '../api/types'

interface CreateScoreConfigsColumnsOptions {
  onEdit: (config: ScoreConfig) => void
  onDelete: (config: ScoreConfig) => void
  isDeleting?: boolean
}

const getConstraintDisplay = (config: ScoreConfig) => {
  if (config.type === 'NUMERIC') {
    if (config.min_value !== undefined || config.max_value !== undefined) {
      return `${config.min_value ?? '−∞'} to ${config.max_value ?? '∞'}`
    }
    return 'Any number'
  }
  if (config.type === 'CATEGORICAL' && config.categories?.length) {
    return config.categories.join(', ')
  }
  if (config.type === 'BOOLEAN') {
    return 'true / false'
  }
  return '—'
}

export function createScoreConfigsColumns({
  onEdit,
  onDelete,
  isDeleting,
}: CreateScoreConfigsColumnsOptions): ColumnDef<ScoreConfig>[] {
  return [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.name}</div>
          {row.original.description && (
            <div className="max-w-[200px] truncate text-xs text-muted-foreground">
              {row.original.description}
            </div>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Data Type',
      cell: ({ row }) => <Badge variant="outline">{row.original.type}</Badge>,
    },
    {
      id: 'constraints',
      header: 'Constraints',
      cell: ({ row }) => (
        <code className="rounded bg-muted px-2 py-1 font-mono text-xs">
          {getConstraintDisplay(row.original)}
        </code>
      ),
    },
    {
      accessorKey: 'created_at',
      header: 'Created',
      cell: ({ row }) => (
        <div className="text-sm">
          {new Date(row.original.created_at).toLocaleDateString()}
        </div>
      ),
    },
    {
      id: 'actions',
      header: () => <span className="sr-only text-right">Actions</span>,
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onEdit(row.original)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onDelete(row.original)
            }}
            disabled={isDeleting}
            className="text-red-600 hover:bg-red-50 hover:text-red-700 dark:hover:bg-red-950"
          >
            {isDeleting ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Trash2 className="h-4 w-4" />
            )}
          </Button>
        </div>
      ),
    },
  ]
}
