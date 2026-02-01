'use client'

import { Trash2, Loader2, Pencil } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import type { AIProviderCredential, AIProvider } from '../types'
import { PROVIDER_INFO } from '../types'
import { ProviderIcon } from './ProviderIcon'

interface CreateAIProvidersColumnsOptions {
  onEdit: (credential: AIProviderCredential) => void
  onDelete: (credential: AIProviderCredential) => void
  isDeleting?: boolean
}

const getProviderDisplayName = (credential: AIProviderCredential): string => {
  return credential.name
}

const getAdapterDisplayName = (adapter: AIProvider): string => {
  return PROVIDER_INFO[adapter]?.name ?? adapter
}

export function createAIProvidersColumns({
  onEdit,
  onDelete,
  isDeleting,
}: CreateAIProvidersColumnsOptions): ColumnDef<AIProviderCredential>[] {
  return [
    {
      accessorKey: 'name',
      header: 'Provider',
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <ProviderIcon provider={row.original.adapter} className="h-5 w-5" />
          <div>
            <div className="font-medium">{getProviderDisplayName(row.original)}</div>
            <div className="text-xs text-muted-foreground">
              {getAdapterDisplayName(row.original.adapter)}
            </div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'key_preview',
      header: 'API Key',
      cell: ({ row }) => (
        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
          {row.original.key_preview}
        </code>
      ),
    },
    {
      accessorKey: 'base_url',
      header: 'Base URL',
      cell: ({ row }) => (
        <div className="text-sm">
          {row.original.base_url ? (
            <code className="text-xs bg-muted px-2 py-1 rounded font-mono truncate max-w-[200px] block">
              {row.original.base_url}
            </code>
          ) : (
            <span className="text-muted-foreground">Default</span>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'created_at',
      header: 'Added',
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
            className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
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
