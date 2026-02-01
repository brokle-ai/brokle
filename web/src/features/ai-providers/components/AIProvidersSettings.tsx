'use client'

import { useState, useMemo } from 'react'
import { AlertCircle, AlertTriangle, Plug } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
} from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import {
  useAIProvidersQuery,
  useDeleteProviderMutation,
} from '../hooks/use-ai-providers'
import type { AIProviderCredential } from '../types'
import { PROVIDER_INFO } from '../types'
import { ProviderDialog } from './ProviderDialog'
import { ProviderIcon } from './ProviderIcon'
import { createAIProvidersColumns } from './ai-providers-columns'

interface AIProvidersSettingsProps {
  orgId: string
  addDialogOpen?: boolean
  onAddDialogOpenChange?: (open: boolean) => void
}

export function AIProvidersSettings({
  orgId,
  addDialogOpen,
  onAddDialogOpenChange,
}: AIProvidersSettingsProps) {
  // React Query hooks
  const { data: credentials, isLoading, error, refetch } = useAIProvidersQuery(orgId)
  const deleteMutation = useDeleteProviderMutation(orgId)

  // Local state - use controlled or uncontrolled mode
  const [internalAddDialogOpen, setInternalAddDialogOpen] = useState(false)
  const isAddDialogOpen = addDialogOpen ?? internalAddDialogOpen
  const setIsAddDialogOpen = onAddDialogOpenChange ?? setInternalAddDialogOpen
  const [editingCredential, setEditingCredential] = useState<AIProviderCredential | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [selectedCredential, setSelectedCredential] = useState<AIProviderCredential | null>(null)

  const openDeleteDialog = (credential: AIProviderCredential) => {
    setSelectedCredential(credential)
    setIsDeleteDialogOpen(true)
  }

  const handleDeleteDialogClose = (open: boolean) => {
    setIsDeleteDialogOpen(open)
    if (!open) {
      setTimeout(() => setSelectedCredential(null), 200)
    }
  }

  const handleConfirmDelete = async () => {
    if (!selectedCredential) return

    try {
      await deleteMutation.mutateAsync({
        credentialId: selectedCredential.id,
        displayName: selectedCredential.name,
      })
      setIsDeleteDialogOpen(false)
    } catch (error) {
      console.error('Failed to delete provider:', error)
    }
  }

  const handleEditClick = (credential: AIProviderCredential) => {
    setEditingCredential(credential)
  }

  const getAdapterDisplayName = (adapter: AIProviderCredential['adapter']): string => {
    return PROVIDER_INFO[adapter]?.name ?? adapter
  }

  // Create columns
  const columns = useMemo(
    () =>
      createAIProvidersColumns({
        onEdit: handleEditClick,
        onDelete: openDeleteDialog,
        isDeleting: deleteMutation.isPending,
      }),
    [deleteMutation.isPending]
  )

  // Initialize React Table
  const table = useReactTable({
    data: credentials || [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className="space-y-8">
      {/* Loading State */}
      {isLoading && (
        <DataTableSkeleton columns={5} rows={3} showToolbar={false} />
      )}

      {/* Error State */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error Loading Providers</AlertTitle>
          <AlertDescription className="space-y-2">
            <p>{error instanceof Error ? error.message : 'Failed to load providers'}</p>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Try Again
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* Providers List */}
      {!isLoading && !error && (
        <div className="space-y-4">
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
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
                      <DataTableEmptyState
                        title="No providers configured yet"
                        description="Add a provider to enable AI features in the playground."
                        icon={<Plug className="h-8 w-8 text-muted-foreground/50" />}
                      />
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {/* Add Provider Dialog */}
      <ProviderDialog
        orgId={orgId}
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        existingCredentials={credentials || []}
      />

      {/* Edit Provider Dialog */}
      {editingCredential && (
        <ProviderDialog
          orgId={orgId}
          open={!!editingCredential}
          onOpenChange={(open) => !open && setEditingCredential(null)}
          existingCredential={editingCredential}
          existingCredentials={credentials || []}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={handleDeleteDialogClose}>
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle className="text-red-600">Delete Provider</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this provider? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {selectedCredential && (
              <div className="rounded-lg bg-muted p-3 space-y-1">
                <div className="flex items-center gap-2">
                  <ProviderIcon provider={selectedCredential.adapter} className="h-4 w-4" />
                  <span className="font-medium">{selectedCredential.name}</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{getAdapterDisplayName(selectedCredential.adapter)}</span>
                  <span>•</span>
                  <code>{selectedCredential.key_preview}</code>
                </div>
              </div>
            )}

            <Alert className="border-red-200 bg-red-50 dark:bg-red-950/20">
              <AlertTriangle className="h-4 w-4 text-red-500" />
              <AlertDescription className="text-red-600 dark:text-red-400">
                <strong>Warning:</strong> Playground sessions using this provider will no longer work.
              </AlertDescription>
            </Alert>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete Provider'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
