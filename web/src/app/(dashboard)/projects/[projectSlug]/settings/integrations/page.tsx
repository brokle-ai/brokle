'use client'

import { useState } from 'react'
import { 
  Zap, 
  Webhook, 
  Database, 
  MessageSquare, 
  Settings,
  Plus,
  CheckCircle,
  AlertCircle,
  ExternalLink
} from 'lucide-react'
import { useOrganization } from '@/context/org-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { TabsContent } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'sonner'

interface Integration {
  id: string
  name: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  category: 'webhook' | 'database' | 'messaging' | 'analytics' | 'monitoring'
  status: 'connected' | 'disconnected' | 'error'
  enabled: boolean
  config: Record<string, string | string[]>
  lastSync?: string
}

const AVAILABLE_INTEGRATIONS: Integration[] = [
  {
    id: 'slack-webhook',
    name: 'Slack Notifications',
    description: 'Send alerts and updates to your Slack channels',
    icon: MessageSquare,
    category: 'messaging',
    status: 'connected',
    enabled: true,
    config: {
      webhook_url: 'https://hooks.slack.com/services/...',
      channel: '#ai-alerts',
      events: ['error', 'quota_exceeded']
    },
    lastSync: '2024-03-15T14:30:00Z'
  },
  {
    id: 'custom-webhook',
    name: 'Custom Webhook',
    description: 'Send events to your custom endpoint',
    icon: Webhook,
    category: 'webhook',
    status: 'connected',
    enabled: true,
    config: {
      url: 'https://api.example.com/webhooks/brokle',
      method: 'POST',
      headers: { 'Authorization': 'Bearer ***' },
      events: ['request_completed', 'error']
    }
  },
  {
    id: 'postgres-logs',
    name: 'PostgreSQL Logging',
    description: 'Store request logs in your PostgreSQL database',
    icon: Database,
    category: 'database',
    status: 'disconnected',
    enabled: false,
    config: {
      host: '',
      database: '',
      username: '',
      password: ''
    }
  }
]

export default function ProjectIntegrationsSettingsPage() {
  const { currentProject } = useOrganization()
  const [integrations, setIntegrations] = useState<Integration[]>(AVAILABLE_INTEGRATIONS)
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [editingIntegration, setEditingIntegration] = useState<Integration | null>(null)
  const [newIntegrationType, setNewIntegrationType] = useState<string>('')

  if (!currentProject) {
    return null
  }

  const toggleIntegration = (integrationId: string) => {
    setIntegrations(integrations.map(integration =>
      integration.id === integrationId
        ? { ...integration, enabled: !integration.enabled }
        : integration
    ))
    toast.success('Integration updated')
  }

  const testIntegration = async (_integrationId: string) => {
    // TODO: Implement actual test
    await new Promise(resolve => setTimeout(resolve, 1000))
    toast.success('Integration test successful')
  }

  const getStatusColor = (status: Integration['status']) => {
    switch (status) {
      case 'connected':
        return 'bg-green-100 text-green-800'
      case 'disconnected':
        return 'bg-gray-100 text-gray-800'
      case 'error':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusIcon = (status: Integration['status']) => {
    switch (status) {
      case 'connected':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'disconnected':
        return <AlertCircle className="h-4 w-4 text-gray-500" />
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-500" />
      default:
        return <AlertCircle className="h-4 w-4 text-gray-500" />
    }
  }

  return (
    <TabsContent value="integrations" className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Integrations</h3>
          <p className="text-sm text-muted-foreground">
            Connect external services to enhance your project capabilities
          </p>
        </div>
        
        <Dialog open={isAddOpen} onOpenChange={setIsAddOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add Integration
            </Button>
          </DialogTrigger>
          
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add New Integration</DialogTitle>
              <DialogDescription>
                Choose an integration type to connect with your project
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Integration Type</Label>
                <Select value={newIntegrationType} onValueChange={setNewIntegrationType}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select integration type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="webhook">Custom Webhook</SelectItem>
                    <SelectItem value="database">Database</SelectItem>
                    <SelectItem value="slack">Slack</SelectItem>
                    <SelectItem value="discord">Discord</SelectItem>
                    <SelectItem value="email">Email</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsAddOpen(false)}>
                Cancel
              </Button>
              <Button disabled={!newIntegrationType}>
                Continue Setup
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Active Integrations */}
      <div className="space-y-4">
        <div className="text-sm font-medium">Active Integrations ({integrations.filter(i => i.enabled).length})</div>
        
        {integrations.map((integration) => (
          <Card key={integration.id}>
            <CardHeader className="pb-4">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-muted rounded-lg">
                    <integration.icon className="h-5 w-5" />
                  </div>
                  <div>
                    <CardTitle className="text-lg flex items-center gap-2">
                      {integration.name}
                      <div className="flex items-center gap-1">
                        {getStatusIcon(integration.status)}
                        <Badge className={getStatusColor(integration.status)}>
                          {integration.status}
                        </Badge>
                      </div>
                    </CardTitle>
                    <CardDescription>{integration.description}</CardDescription>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <Switch
                    checked={integration.enabled}
                    onCheckedChange={() => toggleIntegration(integration.id)}
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setEditingIntegration(integration)}
                  >
                    <Settings className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            
            {integration.enabled && (
              <CardContent className="space-y-4">
                {/* Configuration Preview */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium">Configuration</Label>
                  {integration.id === 'slack-webhook' && (
                    <div className="text-sm space-y-1">
                      <div>Channel: <code className="text-xs bg-muted px-1 rounded">{integration.config.channel}</code></div>
                      <div>Events: {integration.config.events.join(', ')}</div>
                    </div>
                  )}
                  {integration.id === 'custom-webhook' && (
                    <div className="text-sm space-y-1">
                      <div>URL: <code className="text-xs bg-muted px-1 rounded">{integration.config.url}</code></div>
                      <div>Method: {integration.config.method}</div>
                      <div>Events: {integration.config.events.join(', ')}</div>
                    </div>
                  )}
                  {integration.id === 'postgres-logs' && (
                    <div className="text-sm text-muted-foreground">
                      Not configured - click Settings to set up database connection
                    </div>
                  )}
                </div>

                {/* Last Sync */}
                {integration.lastSync && (
                  <div className="text-xs text-muted-foreground">
                    Last sync: {new Date(integration.lastSync).toLocaleString()}
                  </div>
                )}

                {/* Actions */}
                <div className="flex items-center gap-2 pt-2 border-t">
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => testIntegration(integration.id)}
                    disabled={integration.status === 'disconnected'}
                  >
                    <Zap className="mr-2 h-3 w-3" />
                    Test
                  </Button>
                  
                  {integration.config.url && (
                    <Button variant="ghost" size="sm" asChild>
                      <a href="#" target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="mr-2 h-3 w-3" />
                        View Docs
                      </a>
                    </Button>
                  )}
                </div>
              </CardContent>
            )}
          </Card>
        ))}
      </div>

      {/* Edit Integration Dialog */}
      {editingIntegration && (
        <Dialog open={!!editingIntegration} onOpenChange={() => setEditingIntegration(null)}>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Configure {editingIntegration.name}</DialogTitle>
              <DialogDescription>
                Update the configuration for this integration
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              {editingIntegration.id === 'slack-webhook' && (
                <>
                  <div className="space-y-2">
                    <Label>Webhook URL</Label>
                    <Input
                      placeholder="https://hooks.slack.com/services/..."
                      defaultValue={editingIntegration.config.webhook_url}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Channel</Label>
                    <Input
                      placeholder="#alerts"
                      defaultValue={editingIntegration.config.channel}
                    />
                  </div>
                </>
              )}
              
              {editingIntegration.id === 'custom-webhook' && (
                <>
                  <div className="space-y-2">
                    <Label>Webhook URL</Label>
                    <Input
                      placeholder="https://api.example.com/webhooks"
                      defaultValue={editingIntegration.config.url}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>HTTP Method</Label>
                    <Select defaultValue={editingIntegration.config.method}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="POST">POST</SelectItem>
                        <SelectItem value="PUT">PUT</SelectItem>
                        <SelectItem value="PATCH">PATCH</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Custom Headers (JSON)</Label>
                    <Textarea
                      placeholder='{"Authorization": "Bearer token"}'
                      rows={3}
                    />
                  </div>
                </>
              )}

              {editingIntegration.id === 'postgres-logs' && (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Host</Label>
                      <Input placeholder="localhost" />
                    </div>
                    <div className="space-y-2">
                      <Label>Port</Label>
                      <Input placeholder="5432" />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>Database</Label>
                    <Input placeholder="brokle_logs" />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Username</Label>
                      <Input placeholder="postgres" />
                    </div>
                    <div className="space-y-2">
                      <Label>Password</Label>
                      <Input type="password" />
                    </div>
                  </div>
                </>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setEditingIntegration(null)}>
                Cancel
              </Button>
              <Button>
                Save Configuration
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </TabsContent>
  )
}