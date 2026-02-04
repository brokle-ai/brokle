'use client'

import { useState, useMemo } from 'react'
import { Loader2, AlertCircle, AlertTriangle, Key, Copy } from 'lucide-react'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import {
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import { useAPIKeysQuery, useCreateAPIKeyMutation, useDeleteAPIKeyMutation } from '../hooks/use-api-key-queries'
import { createAPIKeysColumns } from './api-keys-columns'
import type { APIKey } from '../types/api-keys'

interface ProjectAPIKeysSectionProps {
  createDialogOpen?: boolean
  onCreateDialogOpenChange?: (open: boolean) => void
}

export function ProjectAPIKeysSection({
  createDialogOpen,
  onCreateDialogOpenChange
}: ProjectAPIKeysSectionProps = {}) {
  const { currentProject } = useWorkspace()

  // React Query hooks
  const { data: apiKeysData, isLoading, error, refetch } = useAPIKeysQuery(currentProject?.id)
  const createMutation = useCreateAPIKeyMutation(currentProject?.id || '')
  const deleteMutation = useDeleteAPIKeyMutation(currentProject?.id || '')

  // Local state - use controlled or uncontrolled mode
  const [internalDialogOpen, setInternalDialogOpen] = useState(false)
  const isDialogOpen = createDialogOpen ?? internalDialogOpen
  const setIsDialogOpen = onCreateDialogOpenChange ?? setInternalDialogOpen
  const [dialogMode, setDialogMode] = useState<'create' | 'success'>('create')
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyExpiry, setNewKeyExpiry] = useState<'30days' | '90days' | 'never'>('90days')
  const [createdKey, setCreatedKey] = useState<APIKey | null>(null)

  // Delete confirmation dialog state
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null)

  const apiKeys = apiKeysData?.data || []

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('API key copied to clipboard')
  }

  const handleCreateAPIKey = async () => {
    if (!newKeyName.trim()) {
      toast.error('Please enter a name for the API key')
      return
    }

    if (newKeyName.length < 2 || newKeyName.length > 100) {
      toast.error('Name must be between 2 and 100 characters')
      return
    }

    try {
      const newKey = await createMutation.mutateAsync({
        name: newKeyName.trim(),
        expiry_option: newKeyExpiry
      })

      setCreatedKey(newKey)
      setDialogMode('success')
      setNewKeyName('')
      setNewKeyExpiry('90days')
    } catch (error) {
      console.error('Failed to create API key:', error)
    }
  }

  const handleDialogClose = (open: boolean) => {
    setIsDialogOpen(open)

    if (!open) {
      setTimeout(() => {
        setDialogMode('create')
        setCreatedKey(null)
        setNewKeyName('')
        setNewKeyExpiry('90days')
      }, 200)
    }
  }

  const openDeleteDialog = (apiKey: APIKey) => {
    setSelectedKey(apiKey)
    setIsDeleteDialogOpen(true)
  }

  const handleDeleteDialogClose = (open: boolean) => {
    setIsDeleteDialogOpen(open)
    if (!open) {
      setTimeout(() => {
        setSelectedKey(null)
      }, 200)
    }
  }

  const handleConfirmDelete = async () => {
    if (!selectedKey) return

    try {
      await deleteMutation.mutateAsync({
        keyId: selectedKey.id,
        keyName: selectedKey.name
      })
      setIsDeleteDialogOpen(false)
    } catch (error) {
      console.error('Failed to delete API key:', error)
    }
  }

  // Create columns
  const columns = useMemo(
    () =>
      createAPIKeysColumns({
        onCopy: copyToClipboard,
        onDelete: openDeleteDialog,
        isDeleting: deleteMutation.isPending,
      }),
    [deleteMutation.isPending]
  )

  // Initialize React Table
  const table = useReactTable({
    data: apiKeys,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (!currentProject) {
    return null
  }

  // Loading state
  if (isLoading) {
    return <DataTableSkeleton columns={6} rows={3} showToolbar={false} />
  }

  // Error state
  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error Loading API Keys</AlertTitle>
        <AlertDescription className="space-y-2">
          <p>{error instanceof Error ? error.message : 'Failed to load API keys'}</p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Try Again
          </Button>
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="space-y-8">
      {/* Create Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={handleDialogClose}>
        <DialogContent className="sm:max-w-[500px]">
          {dialogMode === 'create' ? (
            <>
              <DialogHeader>
                <DialogTitle>Create New API Key</DialogTitle>
                <DialogDescription>
                  Generate a new API key for this project. You'll only see the full key once.
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="keyName">Key Name *</Label>
                  <Input
                    id="keyName"
                    value={newKeyName}
                    onChange={(e) => setNewKeyName(e.target.value)}
                    placeholder="e.g., Production API Key"
                    maxLength={100}
                  />
                  <p className="text-xs text-muted-foreground">
                    2-100 characters
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="keyExpiry">Expiration</Label>
                  <Select
                    value={newKeyExpiry}
                    onValueChange={(value) => setNewKeyExpiry(value as '30days' | '90days' | 'never')}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="30days">30 days</SelectItem>
                      <SelectItem value="90days">90 days</SelectItem>
                      <SelectItem value="never">Never expires</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsDialogOpen(false)}
                  disabled={createMutation.isPending}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateAPIKey}
                  disabled={createMutation.isPending}
                >
                  {createMutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    'Create Key'
                  )}
                </Button>
              </DialogFooter>
            </>
          ) : (
            <>
              <DialogHeader>
                <DialogTitle>API Key Created Successfully!</DialogTitle>
                <DialogDescription>
                  Make sure to copy your API key now. You won't be able to see it again!
                </DialogDescription>
              </DialogHeader>

              {createdKey && (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label>Key Name</Label>
                    <div className="text-sm font-medium">{createdKey.name}</div>
                  </div>

                  <div className="space-y-2">
                    <Label>API Key</Label>
                    <div className="flex items-center gap-2">
                      <code className="flex-1 text-xs bg-muted px-3 py-2 rounded font-mono break-all">
                        {createdKey.key}
                      </code>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          if (createdKey.key) {
                            copyToClipboard(createdKey.key)
                          }
                        }}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>

                  <Alert>
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Important</AlertTitle>
                    <AlertDescription>
                      This is the only time you'll see the full API key. Store it securely - we only store a hashed version.
                    </AlertDescription>
                  </Alert>
                </div>
              )}

              <DialogFooter>
                <Button onClick={() => setIsDialogOpen(false)}>
                  I've Saved My Key
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>

      {/* API Keys Table */}
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
                    title="No API keys yet"
                    description="Create your first key to get started."
                    icon={<Key className="h-8 w-8 text-muted-foreground/50" />}
                  />
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={handleDeleteDialogClose}>
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle className="text-red-600">Delete API Key</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this API key? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {selectedKey && (
              <div className="rounded-lg bg-muted p-3 space-y-1">
                <div className="text-sm font-medium">{selectedKey.name}</div>
                <code className="text-xs text-muted-foreground">{selectedKey.key_preview}</code>
              </div>
            )}

            <Alert className="border-red-200 bg-red-50 dark:bg-red-950/20">
              <AlertTriangle className="h-4 w-4 text-red-500" />
              <AlertDescription className="text-red-600 dark:text-red-400">
                <strong>Warning:</strong> Applications using this key will immediately lose access.
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
                'Delete API Key'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
