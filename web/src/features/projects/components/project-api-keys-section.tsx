'use client'

import { useState } from 'react'
import { Plus, Copy, Trash2, Loader2, AlertCircle } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import { useAPIKeysQuery, useCreateAPIKeyMutation, useUpdateAPIKeyMutation, useDeleteAPIKeyMutation } from '../hooks/use-api-key-queries'
import type { APIKey } from '../types/api-keys'

export function ProjectAPIKeysSection() {
  const { currentProject } = useWorkspace()

  // React Query hooks
  const { data: apiKeysData, isLoading, error, refetch } = useAPIKeysQuery(currentProject?.id)
  const createMutation = useCreateAPIKeyMutation(currentProject?.id || '')
  const updateMutation = useUpdateAPIKeyMutation(currentProject?.id || '')
  const deleteMutation = useDeleteAPIKeyMutation(currentProject?.id || '')

  // Local state
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyExpiry, setNewKeyExpiry] = useState<'30days' | '90days' | 'never'>('90days')
  const [createdKey, setCreatedKey] = useState<APIKey | null>(null)
  const [showCreatedKeyDialog, setShowCreatedKeyDialog] = useState(false)

  if (!currentProject) {
    return null
  }

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

      // Show the created key once
      setCreatedKey(newKey)
      setShowCreatedKeyDialog(true)

      // Reset form
      setNewKeyName('')
      setNewKeyExpiry('90days')
      setIsCreateOpen(false)
    } catch (error) {
      // Error toast handled by mutation hook
      console.error('Failed to create API key:', error)
    }
  }

  const handleDeleteAPIKey = async (keyId: string, keyName: string) => {
    if (!confirm(`Delete "${keyName}"?\n\nThis will immediately revoke access and cannot be undone.`)) {
      return
    }

    try {
      await deleteMutation.mutateAsync({ keyId, keyName })
    } catch (error) {
      // Error toast handled by mutation hook
      console.error('Failed to delete API key:', error)
    }
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

  const maskKey = (keyPreview: string) => {
    // Key preview format: bk_AbCd...XyZa
    return keyPreview
  }

  return (
    <>
      {/* Loading State */}
      {isLoading && (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <div className="flex flex-col items-center gap-2">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Loading API keys...</p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Error State */}
      {error && (
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
      )}

      {/* API Keys Overview */}
      {!isLoading && !error && (
        <Card>
          <CardContent className="space-y-6 pt-6">
            <div className="flex items-center justify-between">
              <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Create API Key
                </Button>
              </DialogTrigger>

              <DialogContent className="sm:max-w-[425px]">
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
                    <Select value={newKeyExpiry} onValueChange={(value) => setNewKeyExpiry(value as '30days' | '90days' | 'never')}>
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
                    onClick={() => setIsCreateOpen(false)}
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
              </DialogContent>
            </Dialog>
          </div>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>API Key</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {apiKeys.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                      No API keys yet. Create your first key to get started.
                    </TableCell>
                  </TableRow>
                ) : (
                  apiKeys.map((apiKey) => (
                    <TableRow key={apiKey.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{apiKey.name}</div>
                          <div className="text-sm text-muted-foreground">
                            Created {new Date(apiKey.created_at).toLocaleDateString()}
                          </div>
                        </div>
                      </TableCell>

                      <TableCell>
                        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                          {maskKey(apiKey.key_preview)}
                        </code>
                      </TableCell>

                      <TableCell>
                        <div className="text-sm">
                          {apiKey.last_used ? (
                            <>
                              <div>{new Date(apiKey.last_used).toLocaleDateString()}</div>
                              <div className="text-muted-foreground">
                                {new Date(apiKey.last_used).toLocaleTimeString()}
                              </div>
                            </>
                          ) : (
                            <span className="text-muted-foreground">Never</span>
                          )}
                        </div>
                      </TableCell>

                      <TableCell>
                        <div className="text-sm">
                          {apiKey.expires_at ? (
                            new Date(apiKey.expires_at).toLocaleDateString()
                          ) : (
                            <span className="text-muted-foreground">Never</span>
                          )}
                        </div>
                      </TableCell>

                      <TableCell>
                        <Badge className={getStatusColor(apiKey.status)}>
                          {apiKey.status}
                        </Badge>
                      </TableCell>

                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={deleteMutation.isPending}
                            >
                              •••
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => handleDeleteAPIKey(apiKey.id, apiKey.name)}
                              className="text-red-600"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
      )}

      {/* Created Key Dialog - Show full key once */}
      <Dialog open={showCreatedKeyDialog} onOpenChange={setShowCreatedKeyDialog}>
        <DialogContent className="sm:max-w-[500px]">
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
            <Button onClick={() => {
              setShowCreatedKeyDialog(false)
              setCreatedKey(null)
            }}>
              I've Saved My Key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Usage Instructions */}
      {!isLoading && !error && (
        <Card>
          <CardContent className="space-y-6 pt-6">
            <div>
              <h4 className="font-medium mb-2">Authentication</h4>
              <div className="bg-muted p-4 rounded-lg">
                <code className="text-sm">
                  curl -H "X-API-Key: YOUR_API_KEY" \<br />
                  &nbsp;&nbsp;&nbsp;&nbsp; https://api.brokle.com/v1/chat/completions
                </code>
              </div>
            </div>

            <div>
              <h4 className="font-medium mb-2">Environment Variables</h4>
              <div className="bg-muted p-4 rounded-lg">
                <code className="text-sm">
                  export BROKLE_API_KEY="YOUR_API_KEY"<br />
                  export BROKLE_PROJECT_ID="{currentProject.id}"
                </code>
              </div>
            </div>

            <div>
              <h4 className="font-medium mb-2">SDK Usage</h4>
              <div className="bg-muted p-4 rounded-lg">
                <code className="text-sm">
                  from brokle import Brokle<br />
                  <br />
                  client = Brokle(<br />
                  &nbsp;&nbsp;&nbsp;&nbsp;api_key="YOUR_API_KEY",<br />
                  &nbsp;&nbsp;&nbsp;&nbsp;project_id="{currentProject.id}"<br />
                  )
                </code>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}
