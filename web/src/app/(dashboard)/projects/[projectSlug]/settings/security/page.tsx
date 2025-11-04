'use client'

import { useState } from 'react'
import { 
  Shield, 
  Key, 
  Eye, 
  EyeOff, 
  AlertTriangle, 
  CheckCircle,
  Clock,
  Globe,
  Ban,
  Trash2
} from 'lucide-react'
import { useOrganization } from '@/context/org-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { TabsContent } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toast } from 'sonner'

interface SecuritySetting {
  id: string
  name: string
  description: string
  enabled: boolean
  level: 'low' | 'medium' | 'high' | 'critical'
}

interface IPRestriction {
  id: string
  ip: string
  description: string
  type: 'allow' | 'deny'
  enabled: boolean
  created_at: string
}

const SECURITY_SETTINGS: SecuritySetting[] = [
  {
    id: 'api_key_rotation',
    name: 'Automatic API Key Rotation',
    description: 'Automatically rotate API keys every 30 days',
    enabled: false,
    level: 'high'
  },
  {
    id: 'request_logging',
    name: 'Enhanced Request Logging',
    description: 'Log all API requests including headers and payloads',
    enabled: true,
    level: 'medium'
  },
  {
    id: 'rate_limiting_strict',
    name: 'Strict Rate Limiting',
    description: 'Apply stricter rate limits than plan default',
    enabled: false,
    level: 'medium'
  },
  {
    id: 'ip_restrictions',
    name: 'IP Address Restrictions',
    description: 'Restrict API access to specific IP addresses',
    enabled: false,
    level: 'high'
  },
  {
    id: 'webhook_verification',
    name: 'Webhook Signature Verification',
    description: 'Verify webhook signatures for all outbound events',
    enabled: true,
    level: 'critical'
  },
  {
    id: 'audit_alerts',
    name: 'Real-time Audit Alerts',
    description: 'Send immediate alerts for suspicious activities',
    enabled: true,
    level: 'critical'
  }
]

const MOCK_IP_RESTRICTIONS: IPRestriction[] = [
  {
    id: 'ip-001',
    ip: '192.168.1.0/24',
    description: 'Office Network',
    type: 'allow',
    enabled: true,
    created_at: '2024-03-01T10:00:00Z'
  },
  {
    id: 'ip-002',
    ip: '10.0.0.100',
    description: 'Production Server',
    type: 'allow',
    enabled: true,
    created_at: '2024-03-05T14:30:00Z'
  },
  {
    id: 'ip-003',
    ip: '203.0.113.0/24',
    description: 'Blocked Suspicious Range',
    type: 'deny',
    enabled: true,
    created_at: '2024-03-10T09:15:00Z'
  }
]

export default function ProjectSecuritySettingsPage() {
  const { currentProject } = useOrganization()
  const [securitySettings, setSecuritySettings] = useState<SecuritySetting[]>(SECURITY_SETTINGS)
  const [ipRestrictions, setIpRestrictions] = useState<IPRestriction[]>(MOCK_IP_RESTRICTIONS)
  const [webhookSecret, setWebhookSecret] = useState('whsec_1234567890abcdef')
  const [showWebhookSecret, setShowWebhookSecret] = useState(false)
  const [newIpAddress, setNewIpAddress] = useState('')
  const [newIpDescription, setNewIpDescription] = useState('')
  const [newIpType, setNewIpType] = useState<'allow' | 'deny'>('allow')

  if (!currentProject) {
    return null
  }

  const toggleSecuritySetting = (settingId: string) => {
    setSecuritySettings(settings =>
      settings.map(setting =>
        setting.id === settingId
          ? { ...setting, enabled: !setting.enabled }
          : setting
      )
    )
    toast.success('Security setting updated')
  }

  const regenerateWebhookSecret = () => {
    const newSecret = 'whsec_' + Math.random().toString(36).substring(2, 24)
    setWebhookSecret(newSecret)
    toast.success('Webhook secret regenerated')
  }

  const addIpRestriction = () => {
    if (!newIpAddress.trim()) {
      toast.error('Please enter an IP address')
      return
    }

    const newRestriction: IPRestriction = {
      id: `ip-${Date.now()}`,
      ip: newIpAddress,
      description: newIpDescription || 'No description',
      type: newIpType,
      enabled: true,
      created_at: new Date().toISOString()
    }

    setIpRestrictions([...ipRestrictions, newRestriction])
    setNewIpAddress('')
    setNewIpDescription('')
    toast.success('IP restriction added')
  }

  const removeIpRestriction = (restrictionId: string) => {
    setIpRestrictions(restrictions => restrictions.filter(r => r.id !== restrictionId))
    toast.success('IP restriction removed')
  }

  const getLevelColor = (level: SecuritySetting['level']) => {
    switch (level) {
      case 'low':
        return 'bg-blue-100 text-blue-800'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800'
      case 'high':
        return 'bg-orange-100 text-orange-800'
      case 'critical':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getTypeColor = (type: IPRestriction['type']) => {
    return type === 'allow' 
      ? 'bg-green-100 text-green-800' 
      : 'bg-red-100 text-red-800'
  }

  const enabledHighSecuritySettings = securitySettings.filter(s => s.enabled && (s.level === 'high' || s.level === 'critical')).length
  const totalHighSecuritySettings = securitySettings.filter(s => s.level === 'high' || s.level === 'critical').length

  return (
    <TabsContent value="security" className="space-y-6">
      {/* Security Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security Overview
          </CardTitle>
          <CardDescription>
            Current security posture and recommendations for this project
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {enabledHighSecuritySettings}/{totalHighSecuritySettings}
              </div>
              <div className="text-sm text-muted-foreground">High Security Features</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {ipRestrictions.filter(r => r.enabled).length}
              </div>
              <div className="text-sm text-muted-foreground">Active IP Rules</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                A+
              </div>
              <div className="text-sm text-muted-foreground">Security Grade</div>
            </div>
          </div>

          {enabledHighSecuritySettings < totalHighSecuritySettings && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                Consider enabling additional high-security features to improve your security posture.
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Security Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Security Features</CardTitle>
          <CardDescription>
            Configure advanced security features for this project
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {securitySettings.map((setting) => (
            <div key={setting.id} className="flex items-start justify-between p-4 border rounded-lg">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <div className="font-medium">{setting.name}</div>
                  <Badge className={getLevelColor(setting.level)}>
                    {setting.level}
                  </Badge>
                </div>
                <div className="text-sm text-muted-foreground">
                  {setting.description}
                </div>
              </div>
              <Switch
                checked={setting.enabled}
                onCheckedChange={() => toggleSecuritySetting(setting.id)}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Webhook Security */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            Webhook Security
          </CardTitle>
          <CardDescription>
            Manage webhook authentication and verification
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>Webhook Signing Secret</Label>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Input
                  type={showWebhookSecret ? 'text' : 'password'}
                  value={webhookSecret}
                  readOnly
                  className="font-mono text-sm pr-10"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowWebhookSecret(!showWebhookSecret)}
                >
                  {showWebhookSecret ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <Button variant="outline" onClick={regenerateWebhookSecret}>
                Regenerate
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Use this secret to verify webhook signatures from Brokle
            </p>
          </div>

          <Alert>
            <Shield className="h-4 w-4" />
            <AlertDescription>
              Store this secret securely and use it to verify that webhook requests are from Brokle.
              Learn more in our <a href="#" className="text-primary hover:underline">webhook security documentation</a>.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      {/* IP Access Control */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            IP Access Control
          </CardTitle>
          <CardDescription>
            Restrict API access based on IP addresses and ranges
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Add New IP Restriction */}
          <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
            <div className="font-medium text-sm">Add IP Restriction</div>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="space-y-2">
                <Label>IP Address/Range</Label>
                <Input
                  placeholder="192.168.1.0/24"
                  value={newIpAddress}
                  onChange={(e) => setNewIpAddress(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Description</Label>
                <Input
                  placeholder="Office network"
                  value={newIpDescription}
                  onChange={(e) => setNewIpDescription(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Type</Label>
                <Select value={newIpType} onValueChange={(value: 'allow' | 'deny') => setNewIpType(value)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="allow">Allow</SelectItem>
                    <SelectItem value="deny">Deny</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>&nbsp;</Label>
                <Button onClick={addIpRestriction} className="w-full">
                  Add Rule
                </Button>
              </div>
            </div>
          </div>

          {/* Current Restrictions */}
          <div className="space-y-4">
            <div className="font-medium text-sm">
              Current Restrictions ({ipRestrictions.length})
            </div>
            
            {ipRestrictions.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>IP Address/Range</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-[100px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {ipRestrictions.map((restriction) => (
                    <TableRow key={restriction.id}>
                      <TableCell className="font-mono text-sm">
                        {restriction.ip}
                      </TableCell>
                      <TableCell>{restriction.description}</TableCell>
                      <TableCell>
                        <Badge className={getTypeColor(restriction.type)}>
                          {restriction.type === 'allow' ? 'Allow' : 'Deny'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <CheckCircle className="h-3 w-3 text-green-500" />
                          <span className="text-sm">Active</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(restriction.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeIpRestriction(restriction.id)}
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <Ban className="mx-auto h-8 w-8 mb-2 opacity-50" />
                <div className="text-sm">No IP restrictions configured</div>
                <div className="text-xs">API access is allowed from all IP addresses</div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Security Recommendations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-orange-500" />
            Security Recommendations
          </CardTitle>
          <CardDescription>
            Improve your project's security with these recommendations
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-3">
            {!securitySettings.find(s => s.id === 'api_key_rotation')?.enabled && (
              <Alert>
                <Shield className="h-4 w-4" />
                <AlertDescription>
                  <strong>Enable API Key Rotation:</strong> Automatically rotate your API keys every 30 days to reduce the risk of compromised keys.
                </AlertDescription>
              </Alert>
            )}
            
            {ipRestrictions.length === 0 && (
              <Alert>
                <Globe className="h-4 w-4" />
                <AlertDescription>
                  <strong>Add IP Restrictions:</strong> Limit API access to trusted IP addresses for enhanced security.
                </AlertDescription>
              </Alert>
            )}

            {!securitySettings.find(s => s.id === 'audit_alerts')?.enabled && (
              <Alert>
                <Clock className="h-4 w-4" />
                <AlertDescription>
                  <strong>Enable Audit Alerts:</strong> Get notified immediately of suspicious activities or security events.
                </AlertDescription>
              </Alert>
            )}
          </div>
        </CardContent>
      </Card>
    </TabsContent>
  )
}