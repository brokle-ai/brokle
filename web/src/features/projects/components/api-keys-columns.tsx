'use client'

import { Copy, Trash2, Loader2 } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { APIKey } from '../types/api-keys'

interface CreateAPIKeysColumnsOptions {
  onCopy: (key: string) => void
  onDelete: (apiKey: APIKey) => void
  isDeleting?: boolean
}

const getStatusColor = (status: APIKey['status']) => {
  switch (status) {
    case 'active':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
    case 'expired':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
  }
}

export function createAPIKeysColumns({
  onCopy,
  onDelete,
  isDeleting,
}: CreateAPIKeysColumnsOptions): ColumnDef<APIKey>[] {
  return [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.name}</div>
          <div className="text-sm text-muted-foreground">
            Created {new Date(row.original.created_at).toLocaleDateString()}
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'key_preview',
      header: 'API Key',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
            {row.original.key_preview}
          </code>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={(e) => {
              e.stopPropagation()
              onCopy(row.original.key_preview)
            }}
          >
            <Copy className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
    {
      accessorKey: 'last_used',
      header: 'Last Used',
      cell: ({ row }) => (
        <div className="text-sm">
          {row.original.last_used ? (
            <>
              <div>{new Date(row.original.last_used).toLocaleDateString()}</div>
              <div className="text-muted-foreground">
                {new Date(row.original.last_used).toLocaleTimeString()}
              </div>
            </>
          ) : (
            <span className="text-muted-foreground">Never</span>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'expires_at',
      header: 'Expires',
      cell: ({ row }) => (
        <div className="text-sm">
          {row.original.expires_at ? (
            new Date(row.original.expires_at).toLocaleDateString()
          ) : (
            <span className="text-muted-foreground">Never</span>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => (
        <Badge className={getStatusColor(row.original.status)}>
          {row.original.status}
        </Badge>
      ),
    },
    {
      id: 'actions',
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => (
        <div className="text-right">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onDelete(row.original)
            }}
            disabled={isDeleting}
            className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
            title="Delete API key"
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
