'use client'

import { useState } from 'react'
import { 
  Shield, 
  User, 
  Settings, 
  Key, 
  Trash2, 
  Edit, 
  Plus, 
  Eye, 
  AlertTriangle,
  CheckCircle,
  Clock,
  Filter,
  Download,
  RefreshCw,
  Search
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

interface AuditLog {
  id: string
  timestamp: string
  actor: {
    id: string
    name: string
    email: string
    type: 'user' | 'system' | 'api'
  }
  action: string
  resource: {
    type: 'project' | 'organization' | 'member' | 'api_key' | 'settings' | 'billing'
    id: string
    name: string
  }
  details: {
    ip_address: string
    user_agent?: string
    changes?: Record<string, { from: any; to: any }>
    metadata?: Record<string, any>
  }
  severity: 'info' | 'warning' | 'critical'
  category: 'authentication' | 'authorization' | 'data_change' | 'system' | 'security'
}

const MOCK_AUDIT_LOGS: AuditLog[] = [
  {
    id: 'log-001',
    timestamp: '2024-03-15T14:30:00Z',
    actor: {
      id: 'user-123',
      name: 'John Doe',
      email: 'john@acmecorp.com',
      type: 'user'
    },
    action: 'project.created',
    resource: {
      type: 'project',
      id: 'proj-456',
      name: 'AI Customer Support'
    },
    details: {
      ip_address: '192.168.1.100',
      user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      metadata: {
        template: 'chatbot',
        environment: 'production'
      }
    },
    severity: 'info',
    category: 'data_change'
  },
  {
    id: 'log-002',
    timestamp: '2024-03-15T14:25:00Z',
    actor: {
      id: 'user-456',
      name: 'Jane Smith',
      email: 'jane@acmecorp.com',
      type: 'user'
    },
    action: 'member.role_changed',
    resource: {
      type: 'member',
      id: 'member-789',
      name: 'Mike Johnson'
    },
    details: {
      ip_address: '192.168.1.101',
      user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      changes: {
        role: { from: 'developer', to: 'admin' }
      }
    },
    severity: 'warning',
    category: 'authorization'
  },
  {
    id: 'log-003',
    timestamp: '2024-03-15T14:20:00Z',
    actor: {
      id: 'system',
      name: 'System',
      email: 'system@brokle.com',
      type: 'system'
    },
    action: 'api_key.expired',
    resource: {
      type: 'api_key',
      id: 'key-old-003',
      name: 'Legacy Integration'
    },
    details: {
      ip_address: '10.0.0.1',
      metadata: {
        expiry_date: '2024-03-15T14:20:00Z',
        auto_disabled: true
      }
    },
    severity: 'critical',
    category: 'security'
  },
  {
    id: 'log-004',
    timestamp: '2024-03-15T14:15:00Z',
    actor: {
      id: 'user-123',
      name: 'John Doe',
      email: 'john@acmecorp.com',
      type: 'user'
    },
    action: 'settings.billing_updated',
    resource: {
      type: 'billing',
      id: 'billing-001',
      name: 'Payment Method'
    },
    details: {
      ip_address: '192.168.1.100',
      user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      changes: {
        payment_method: { from: '****4242', to: '****5678' }
      }
    },
    severity: 'info',
    category: 'data_change'
  },
  {
    id: 'log-005',
    timestamp: '2024-03-15T14:10:00Z',
    actor: {
      id: 'api-client-789',
      name: 'Production API Client',
      email: 'api@acmecorp.com',
      type: 'api'
    },
    action: 'project.deleted',
    resource: {
      type: 'project',
      id: 'proj-old-123',
      name: 'Deprecated Test Project'
    },
    details: {
      ip_address: '203.0.113.1',
      metadata: {
        force_delete: true,
        confirmed_by: 'user-123'
      }
    },
    severity: 'critical',
    category: 'data_change'
  }
]

interface AuditLogsViewerProps {
  organizationId?: string
  projectId?: string
  level: 'organization' | 'project'
}

export function AuditLogsViewer({ organizationId, projectId, level }: AuditLogsViewerProps) {
  const [logs, setLogs] = useState<AuditLog[]>(MOCK_AUDIT_LOGS)
  const [searchTerm, setSearchTerm] = useState('')
  const [severityFilter, setSeverityFilter] = useState<string>('all')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [actorTypeFilter, setActorTypeFilter] = useState<string>('all')
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d' | 'all'>('24h')
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const filteredLogs = logs.filter(log => {
    const matchesSearch = 
      log.action.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.actor.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.actor.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.resource.name.toLowerCase().includes(searchTerm.toLowerCase())

    const matchesSeverity = severityFilter === 'all' || log.severity === severityFilter
    const matchesCategory = categoryFilter === 'all' || log.category === categoryFilter
    const matchesActorType = actorTypeFilter === 'all' || log.actor.type === actorTypeFilter

    return matchesSearch && matchesSeverity && matchesCategory && matchesActorType
  })

  const handleRefresh = async () => {
    setIsRefreshing(true)
    // TODO: Implement actual refresh logic
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsRefreshing(false)
    toast.success('Audit logs refreshed')
  }

  const handleExport = () => {
    const exportData = {
      metadata: {
        exported_at: new Date().toISOString(),
        level,
        organization_id: organizationId,
        project_id: projectId,
        filters: {
          search: searchTerm,
          severity: severityFilter,
          category: categoryFilter,
          actor_type: actorTypeFilter,
          time_range: timeRange
        },
        total_records: filteredLogs.length
      },
      logs: filteredLogs
    }

    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-logs-${level}-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)

    toast.success('Audit logs exported successfully')
  }

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'info':
        return <CheckCircle className="h-4 w-4 text-blue-500" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      case 'critical':
        return <AlertTriangle className="h-4 w-4 text-red-500" />
      default:
        return <Clock className="h-4 w-4 text-gray-500" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'info':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'warning':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'critical':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'authentication':
        return <User className="h-4 w-4" />
      case 'authorization':
        return <Shield className="h-4 w-4" />
      case 'data_change':
        return <Edit className="h-4 w-4" />
      case 'system':
        return <Settings className="h-4 w-4" />
      case 'security':
        return <Key className="h-4 w-4" />
      default:
        return <Eye className="h-4 w-4" />
    }
  }

  const getActionIcon = (action: string) => {
    if (action.includes('created')) return <Plus className="h-3 w-3" />
    if (action.includes('deleted')) return <Trash2 className="h-3 w-3" />
    if (action.includes('updated') || action.includes('changed')) return <Edit className="h-3 w-3" />
    return <Eye className="h-3 w-3" />
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  const getActorTypeColor = (type: string) => {
    switch (type) {
      case 'user':
        return 'bg-green-100 text-green-800'
      case 'system':
        return 'bg-gray-100 text-gray-800'
      case 'api':
        return 'bg-purple-100 text-purple-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Audit Logs</h2>
          <p className="text-muted-foreground">
            Track all activities and changes in your {level}
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isRefreshing}>
            <RefreshCw className={cn("mr-2 h-4 w-4", isRefreshing && "animate-spin")} />
            Refresh
          </Button>
          
          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters & Search
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
              <Input
                placeholder="Search by action, user, or resource..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            
            <Select value={severityFilter} onValueChange={setSeverityFilter}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder="Severity" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Severity</SelectItem>
                <SelectItem value="info">Info</SelectItem>
                <SelectItem value="warning">Warning</SelectItem>
                <SelectItem value="critical">Critical</SelectItem>
              </SelectContent>
            </Select>

            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-full sm:w-36">
                <SelectValue placeholder="Category" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Categories</SelectItem>
                <SelectItem value="authentication">Authentication</SelectItem>
                <SelectItem value="authorization">Authorization</SelectItem>
                <SelectItem value="data_change">Data Changes</SelectItem>
                <SelectItem value="system">System</SelectItem>
                <SelectItem value="security">Security</SelectItem>
              </SelectContent>
            </Select>

            <Select value={actorTypeFilter} onValueChange={setActorTypeFilter}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder="Actor" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Actors</SelectItem>
                <SelectItem value="user">Users</SelectItem>
                <SelectItem value="system">System</SelectItem>
                <SelectItem value="api">API</SelectItem>
              </SelectContent>
            </Select>

            <Select value={timeRange} onValueChange={(value: any) => setTimeRange(value)}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder="Time Range" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1h">Last Hour</SelectItem>
                <SelectItem value="24h">Last 24h</SelectItem>
                <SelectItem value="7d">Last 7 days</SelectItem>
                <SelectItem value="30d">Last 30 days</SelectItem>
                <SelectItem value="all">All Time</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <span>Showing {filteredLogs.length} of {logs.length} logs</span>
            {(searchTerm || severityFilter !== 'all' || categoryFilter !== 'all' || actorTypeFilter !== 'all') && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSearchTerm('')
                  setSeverityFilter('all')
                  setCategoryFilter('all')
                  setActorTypeFilter('all')
                }}
              >
                Clear Filters
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Logs Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Category</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredLogs.map((log) => (
                <TableRow key={log.id} className="cursor-pointer hover:bg-muted/50">
                  <TableCell className="text-sm">
                    {formatTimestamp(log.timestamp)}
                  </TableCell>
                  
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getSeverityIcon(log.severity)}
                      <Badge className={cn("text-xs", getSeverityColor(log.severity))}>
                        {log.severity}
                      </Badge>
                    </div>
                  </TableCell>
                  
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getActionIcon(log.action)}
                      <span className="font-medium text-sm">{log.action}</span>
                    </div>
                  </TableCell>
                  
                  <TableCell>
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{log.actor.name}</span>
                        <Badge className={cn("text-xs", getActorTypeColor(log.actor.type))}>
                          {log.actor.type}
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground">{log.actor.email}</div>
                    </div>
                  </TableCell>
                  
                  <TableCell>
                    <div className="space-y-1">
                      <div className="font-medium text-sm">{log.resource.name}</div>
                      <div className="text-xs text-muted-foreground capitalize">
                        {log.resource.type.replace('_', ' ')}
                      </div>
                    </div>
                  </TableCell>
                  
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getCategoryIcon(log.category)}
                      <span className="text-sm capitalize">{log.category.replace('_', ' ')}</span>
                    </div>
                  </TableCell>
                  
                  <TableCell>
                    <Dialog>
                      <DialogTrigger asChild>
                        <Button variant="ghost" size="sm" onClick={() => setSelectedLog(log)}>
                          <Eye className="h-4 w-4" />
                        </Button>
                      </DialogTrigger>
                      
                      <DialogContent className="sm:max-w-[600px]">
                        <DialogHeader>
                          <DialogTitle>Audit Log Details</DialogTitle>
                          <DialogDescription>
                            Detailed information about this audit log entry
                          </DialogDescription>
                        </DialogHeader>

                        {selectedLog && (
                          <Tabs defaultValue="overview">
                            <TabsList>
                              <TabsTrigger value="overview">Overview</TabsTrigger>
                              <TabsTrigger value="details">Technical Details</TabsTrigger>
                            </TabsList>

                            <TabsContent value="overview" className="space-y-4">
                              <div className="grid grid-cols-2 gap-4">
                                <div>
                                  <Label className="text-sm font-medium">Action</Label>
                                  <div className="flex items-center gap-2 mt-1">
                                    {getActionIcon(selectedLog.action)}
                                    <span>{selectedLog.action}</span>
                                  </div>
                                </div>
                                
                                <div>
                                  <Label className="text-sm font-medium">Severity</Label>
                                  <div className="flex items-center gap-2 mt-1">
                                    {getSeverityIcon(selectedLog.severity)}
                                    <Badge className={getSeverityColor(selectedLog.severity)}>
                                      {selectedLog.severity}
                                    </Badge>
                                  </div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Actor</Label>
                                  <div className="mt-1">
                                    <div className="font-medium">{selectedLog.actor.name}</div>
                                    <div className="text-sm text-muted-foreground">{selectedLog.actor.email}</div>
                                  </div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Resource</Label>
                                  <div className="mt-1">
                                    <div className="font-medium">{selectedLog.resource.name}</div>
                                    <div className="text-sm text-muted-foreground capitalize">
                                      {selectedLog.resource.type.replace('_', ' ')}
                                    </div>
                                  </div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Timestamp</Label>
                                  <div className="mt-1 text-sm">{formatTimestamp(selectedLog.timestamp)}</div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Category</Label>
                                  <div className="flex items-center gap-2 mt-1">
                                    {getCategoryIcon(selectedLog.category)}
                                    <span className="capitalize">{selectedLog.category.replace('_', ' ')}</span>
                                  </div>
                                </div>
                              </div>

                              {selectedLog.details.changes && (
                                <div>
                                  <Label className="text-sm font-medium">Changes Made</Label>
                                  <div className="mt-2 space-y-2">
                                    {Object.entries(selectedLog.details.changes).map(([field, change]) => (
                                      <div key={field} className="p-2 border rounded text-sm">
                                        <div className="font-medium capitalize">{field.replace('_', ' ')}</div>
                                        <div className="flex items-center gap-2 mt-1">
                                          <span className="text-red-600">From: {change.from}</span>
                                          <span>→</span>
                                          <span className="text-green-600">To: {change.to}</span>
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </TabsContent>

                            <TabsContent value="details" className="space-y-4">
                              <div className="space-y-4">
                                <div>
                                  <Label className="text-sm font-medium">IP Address</Label>
                                  <div className="mt-1 text-sm font-mono">{selectedLog.details.ip_address}</div>
                                </div>

                                {selectedLog.details.user_agent && (
                                  <div>
                                    <Label className="text-sm font-medium">User Agent</Label>
                                    <div className="mt-1 text-sm font-mono text-muted-foreground">
                                      {selectedLog.details.user_agent}
                                    </div>
                                  </div>
                                )}

                                {selectedLog.details.metadata && (
                                  <div>
                                    <Label className="text-sm font-medium">Additional Metadata</Label>
                                    <div className="mt-1 p-3 bg-muted rounded-lg">
                                      <pre className="text-xs">
                                        {JSON.stringify(selectedLog.details.metadata, null, 2)}
                                      </pre>
                                    </div>
                                  </div>
                                )}

                                <div>
                                  <Label className="text-sm font-medium">Log ID</Label>
                                  <div className="mt-1 text-sm font-mono">{selectedLog.id}</div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Resource ID</Label>
                                  <div className="mt-1 text-sm font-mono">{selectedLog.resource.id}</div>
                                </div>

                                <div>
                                  <Label className="text-sm font-medium">Actor ID</Label>
                                  <div className="mt-1 text-sm font-mono">{selectedLog.actor.id}</div>
                                </div>
                              </div>
                            </TabsContent>
                          </Tabs>
                        )}
                      </DialogContent>
                    </Dialog>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {filteredLogs.length === 0 && (
            <div className="text-center py-12">
              <Shield className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium mb-2">No audit logs found</h3>
              <p className="text-muted-foreground mb-4">
                {searchTerm || severityFilter !== 'all' || categoryFilter !== 'all' || actorTypeFilter !== 'all'
                  ? 'Try adjusting your filters to find what you\'re looking for.'
                  : 'Audit logs will appear here as activities occur in your organization.'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}