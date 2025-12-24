'use client'

import { useState } from 'react'
import { Trash2, Loader2, AlertCircle, AlertTriangle, Plug, Pencil } from 'lucide-react'
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
  useAIProvidersQuery,
  useDeleteProviderMutation,
} from '../hooks/use-ai-providers'
import type { AIProviderCredential, AIProvider } from '../types'
import { PROVIDER_INFO } from '../types'
import { ProviderDialog } from './ProviderDialog'
import { ProviderIcon } from './ProviderIcon'

interface AIProvidersSettingsProps {
  projectId: string
  addDialogOpen?: boolean
  onAddDialogOpenChange?: (open: boolean) => void
}

export function AIProvidersSettings({
  projectId,
  addDialogOpen,
  onAddDialogOpenChange,
}: AIProvidersSettingsProps) {
  // React Query hooks
  const { data: credentials, isLoading, error, refetch } = useAIProvidersQuery(projectId)
  const deleteMutation = useDeleteProviderMutation(projectId)

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
        displayName: getProviderDisplayName(selectedCredential),
      })
      setIsDeleteDialogOpen(false)
    } catch (error) {
      console.error('Failed to delete provider:', error)
    }
  }

  const handleEditClick = (credential: AIProviderCredential) => {
    setEditingCredential(credential)
  }

  const getProviderDisplayName = (credential: AIProviderCredential): string => {
    // Name is now always set for all configurations
    return credential.name
  }

  const getAdapterDisplayName = (adapter: AIProvider): string => {
    return PROVIDER_INFO[adapter]?.name ?? adapter
  }

  return (
    <div className="space-y-8">
      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="flex flex-col items-center gap-2">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading providers...</p>
          </div>
        </div>
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
                <TableRow>
                  <TableHead>Provider</TableHead>
                  <TableHead>API Key</TableHead>
                  <TableHead>Base URL</TableHead>
                  <TableHead>Added</TableHead>
                  <TableHead className="w-[100px] text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {!credentials || credentials.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                      <div className="flex flex-col items-center gap-2">
                        <Plug className="h-8 w-8 text-muted-foreground/50" />
                        <p>No providers configured yet.</p>
                        <p className="text-xs">Add a provider to enable AI features in the playground.</p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  credentials.map((credential) => (
                    <TableRow key={credential.id}>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <ProviderIcon provider={credential.adapter} className="h-5 w-5" />
                          <div>
                            <div className="font-medium">{getProviderDisplayName(credential)}</div>
                            <div className="text-xs text-muted-foreground">
                              {getAdapterDisplayName(credential.adapter)}
                            </div>
                          </div>
                        </div>
                      </TableCell>

                      <TableCell>
                        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                          {credential.key_preview}
                        </code>
                      </TableCell>

                      <TableCell>
                        <div className="text-sm">
                          {credential.base_url ? (
                            <code className="text-xs bg-muted px-2 py-1 rounded font-mono truncate max-w-[200px] block">
                              {credential.base_url}
                            </code>
                          ) : (
                            <span className="text-muted-foreground">Default</span>
                          )}
                        </div>
                      </TableCell>

                      <TableCell>
                        <div className="text-sm">
                          {new Date(credential.created_at).toLocaleDateString()}
                        </div>
                      </TableCell>

                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleEditClick(credential)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openDeleteDialog(credential)}
                            disabled={deleteMutation.isPending}
                            className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                          >
                            {deleteMutation.isPending ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <Trash2 className="h-4 w-4" />
                            )}
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {/* Add Provider Dialog */}
      <ProviderDialog
        projectId={projectId}
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        existingCredentials={credentials || []}
      />

      {/* Edit Provider Dialog */}
      {editingCredential && (
        <ProviderDialog
          projectId={projectId}
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
                  <span className="font-medium">{getProviderDisplayName(selectedCredential)}</span>
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
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Provider'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
