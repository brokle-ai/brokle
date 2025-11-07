'use client'

import { useState } from 'react'
import { Key, Plus, Copy, Eye, EyeOff, Trash2, Shield } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import { toast } from 'sonner'

interface APIKey {
  id: string
  name: string
  key: string
  permissions: string[]
  created_at: string
  last_used: string | null
  expires_at: string | null
  status: 'active' | 'inactive' | 'expired'
}

const MOCK_API_KEYS: APIKey[] = [
  {
    id: 'key-prod-001',
    name: 'Production API Key',
    key: 'bk-proj-1234567890abcdef1234567890abcdef',
    permissions: ['read', 'write'],
    created_at: '2024-01-15T10:30:00Z',
    last_used: '2024-03-10T14:20:00Z',
    expires_at: null,
    status: 'active'
  },
  {
    id: 'key-dev-002',
    name: 'Development Testing',
    key: 'bk-proj-abcdef1234567890abcdef1234567890',
    permissions: ['read'],
    created_at: '2024-02-01T09:00:00Z',
    last_used: '2024-03-09T16:45:00Z',
    expires_at: '2024-12-31T23:59:59Z',
    status: 'active'
  },
  {
    id: 'key-old-003',
    name: 'Legacy Integration',
    key: 'bk-proj-fedcba0987654321fedcba0987654321',
    permissions: ['read', 'write'],
    created_at: '2023-11-20T15:00:00Z',
    last_used: '2024-01-15T11:30:00Z',
    expires_at: '2024-01-31T23:59:59Z',
    status: 'expired'
  }
]

export function ProjectAPIKeysSection() {
  const { currentProject } = useWorkspace()

  const [apiKeys, setApiKeys] = useState<APIKey[]>(MOCK_API_KEYS)
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set())
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyPermissions, setNewKeyPermissions] = useState<string[]>(['read'])
  const [newKeyExpiry, setNewKeyExpiry] = useState<string>('never')

  if (!currentProject) {
    return null
  }

  const toggleKeyVisibility = (keyId: string) => {
    const newVisible = new Set(visibleKeys)
    if (newVisible.has(keyId)) {
      newVisible.delete(keyId)
    } else {
      newVisible.add(keyId)
    }
    setVisibleKeys(newVisible)
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('API key copied to clipboard')
  }

  const createAPIKey = async () => {
    if (!newKeyName.trim()) {
      toast.error('Please enter a name for the API key')
      return
    }

    const newKey: APIKey = {
      id: `key-${Date.now()}`,
      name: newKeyName,
      key: `bk-proj-${Math.random().toString(36).substring(2, 34)}`,
      permissions: newKeyPermissions,
      created_at: new Date().toISOString(),
      last_used: null,
      expires_at: newKeyExpiry === 'never' ? null : new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
      status: 'active'
    }

    setApiKeys([newKey, ...apiKeys])
    setNewKeyName('')
    setNewKeyPermissions(['read'])
    setNewKeyExpiry('never')
    setIsCreateOpen(false)

    toast.success('API key created successfully')
  }

  const deleteAPIKey = (keyId: string) => {
    setApiKeys(apiKeys.filter(key => key.id !== keyId))
    toast.success('API key deleted')
  }

  const revokeAPIKey = (keyId: string) => {
    setApiKeys(apiKeys.map(key =>
      key.id === keyId ? { ...key, status: 'inactive' as const } : key
    ))
    toast.success('API key revoked')
  }

  const getStatusColor = (status: APIKey['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'inactive':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      case 'expired':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const maskKey = (key: string) => {
    if (key.length <= 8) return key
    return key.substring(0, 8) + '•'.repeat(24) + key.substring(key.length - 4)
  }

  return (
    <>
      {/* API Keys Overview */}
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
                    Generate a new API key for this project with specific permissions.
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
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Permissions</Label>
                    <div className="space-y-2">
                      <label className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          checked={newKeyPermissions.includes('read')}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setNewKeyPermissions([...newKeyPermissions, 'read'])
                            } else {
                              setNewKeyPermissions(newKeyPermissions.filter(p => p !== 'read'))
                            }
                          }}
                        />
                        <span className="text-sm">Read access</span>
                      </label>
                      <label className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          checked={newKeyPermissions.includes('write')}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setNewKeyPermissions([...newKeyPermissions, 'write'])
                            } else {
                              setNewKeyPermissions(newKeyPermissions.filter(p => p !== 'write'))
                            }
                          }}
                        />
                        <span className="text-sm">Write access</span>
                      </label>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="keyExpiry">Expiration</Label>
                    <Select value={newKeyExpiry} onValueChange={setNewKeyExpiry}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="never">Never expires</SelectItem>
                        <SelectItem value="30days">30 days</SelectItem>
                        <SelectItem value="90days">90 days</SelectItem>
                        <SelectItem value="1year">1 year</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={createAPIKey}>Create Key</Button>
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
                  <TableHead>Permissions</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {apiKeys.map((apiKey) => (
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
                      <div className="flex items-center gap-2">
                        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                          {visibleKeys.has(apiKey.id) ? apiKey.key : maskKey(apiKey.key)}
                        </code>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleKeyVisibility(apiKey.id)}
                        >
                          {visibleKeys.has(apiKey.id) ? (
                            <EyeOff className="h-3 w-3" />
                          ) : (
                            <Eye className="h-3 w-3" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => copyToClipboard(apiKey.key)}
                        >
                          <Copy className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>

                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {apiKey.permissions.map((permission) => (
                          <Badge key={permission} variant="secondary" className="text-xs">
                            {permission}
                          </Badge>
                        ))}
                      </div>
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
                      <Badge className={getStatusColor(apiKey.status)}>
                        {apiKey.status}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            •••
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => copyToClipboard(apiKey.key)}>
                            <Copy className="mr-2 h-4 w-4" />
                            Copy Key
                          </DropdownMenuItem>
                          {apiKey.status === 'active' && (
                            <DropdownMenuItem onClick={() => revokeAPIKey(apiKey.id)}>
                              <Shield className="mr-2 h-4 w-4" />
                              Revoke
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => deleteAPIKey(apiKey.id)}
                            className="text-red-600"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Usage Instructions */}
      <Card>
        <CardContent className="space-y-6 pt-6">
          <div>
            <h4 className="font-medium mb-2">Authentication</h4>
            <div className="bg-muted p-4 rounded-lg">
              <code className="text-sm">
                curl -H "Authorization: Bearer YOUR_API_KEY" \<br />
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
    </>
  )
}
