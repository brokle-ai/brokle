import { useState } from 'react'
import { useSuspenseQuery } from '@tanstack/react-query'
import { AlertTriangle, Loader2, Pencil, Plug, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTableEmptyState } from '@/components/shared/tables/data-table-empty-state'
import {
  aiProvidersListQueryOptions,
  useDeleteProviderMutation,
} from '../api/queries'
import type { AIProviderCredential } from '../api/types'
import { PROVIDER_INFO } from '../api/types'
import { ProviderDialog } from './provider-dialog'
import { ProviderIcon } from './provider-icon'

interface AIProvidersTableProps {
  orgId: string
  addDialogOpen: boolean
  onAddDialogOpenChange: (open: boolean) => void
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString()
}

function adapterDisplayName(adapter: AIProviderCredential['adapter']): string {
  return PROVIDER_INFO[adapter]?.name ?? adapter
}

// List + CRUD surface. The route loader prefetches the list so
// `useSuspenseQuery` is a zero-flash read. Add/edit/delete share the
// same ProviderDialog; the delete confirmation is a separate inline
// dialog so we can surface the impact-on-playground warning.
export function AIProvidersTable({
  orgId,
  addDialogOpen,
  onAddDialogOpenChange,
}: AIProvidersTableProps) {
  const { data: credentials } = useSuspenseQuery(
    aiProvidersListQueryOptions(orgId),
  )
  const deleteMutation = useDeleteProviderMutation(orgId)

  const [editingCredential, setEditingCredential] =
    useState<AIProviderCredential | null>(null)
  const [deleteTarget, setDeleteTarget] =
    useState<AIProviderCredential | null>(null)

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteMutation.mutateAsync({
        credentialId: deleteTarget.id,
        displayName: deleteTarget.name,
      })
      setDeleteTarget(null)
    } catch {
      // Mutation onError surfaces the toast; keep dialog open so the
      // user can retry.
    }
  }

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Provider</TableHead>
              <TableHead>API key</TableHead>
              <TableHead>Base URL</TableHead>
              <TableHead>Added</TableHead>
              <TableHead className="w-[1%] text-right">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {credentials.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  <DataTableEmptyState
                    title="No providers configured yet"
                    description="Add a provider to enable AI features in the playground."
                    icon={<Plug className="h-full w-full" />}
                  />
                </TableCell>
              </TableRow>
            ) : (
              credentials.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <ProviderIcon provider={c.adapter} className="h-5 w-5" />
                      <div>
                        <div className="font-medium">{c.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {adapterDisplayName(c.adapter)}
                        </div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <code className="rounded bg-muted px-2 py-1 font-mono text-xs">
                      {c.key_preview}
                    </code>
                  </TableCell>
                  <TableCell>
                    {c.base_url ? (
                      <code className="block max-w-[200px] truncate rounded bg-muted px-2 py-1 font-mono text-xs">
                        {c.base_url}
                      </code>
                    ) : (
                      <span className="text-muted-foreground">Default</span>
                    )}
                  </TableCell>
                  <TableCell className="text-sm">
                    {formatDate(c.created_at)}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditingCredential(c)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setDeleteTarget(c)}
                        disabled={deleteMutation.isPending}
                        className="text-red-600 hover:bg-red-50 hover:text-red-700 dark:hover:bg-red-950"
                      >
                        {deleteMutation.isPending &&
                        deleteMutation.variables?.credentialId === c.id ? (
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

      <ProviderDialog
        orgId={orgId}
        open={addDialogOpen}
        onOpenChange={onAddDialogOpenChange}
        existingCredentials={credentials}
      />

      {editingCredential && (
        <ProviderDialog
          orgId={orgId}
          open
          onOpenChange={(open) => {
            if (!open) setEditingCredential(null)
          }}
          existingCredential={editingCredential}
          existingCredentials={credentials}
        />
      )}

      <Dialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
      >
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle className="text-red-600">Delete provider</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this provider? This action cannot
              be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {deleteTarget && (
              <div className="space-y-1 rounded-lg bg-muted p-3">
                <div className="flex items-center gap-2">
                  <ProviderIcon
                    provider={deleteTarget.adapter}
                    className="h-4 w-4"
                  />
                  <span className="font-medium">{deleteTarget.name}</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{adapterDisplayName(deleteTarget.adapter)}</span>
                  <span>•</span>
                  <code>{deleteTarget.key_preview}</code>
                </div>
              </div>
            )}

            <Alert className="border-red-200 bg-red-50 dark:bg-red-950/20">
              <AlertTriangle className="h-4 w-4 text-red-500" />
              <AlertDescription className="text-red-600 dark:text-red-400">
                <strong>Warning:</strong> Playground sessions using this
                provider will stop working.
              </AlertDescription>
            </Alert>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteTarget(null)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting…' : 'Delete provider'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
