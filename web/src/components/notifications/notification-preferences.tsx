'use client'

import { useState } from 'react'
import { 
  Bell, 
  Mail, 
  Smartphone, 
  Slack, 
  MessageSquare,
  AlertTriangle,
  DollarSign,
  Activity,
  Users,
  Shield,
  Settings,
  Save,
  Volume2,
  VolumeX
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Slider } from '@/components/ui/slider'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface NotificationChannel {
  id: string
  name: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  enabled: boolean
  config?: Record<string, any>
}

interface NotificationRule {
  id: string
  name: string
  description: string
  category: 'usage' | 'security' | 'billing' | 'system' | 'collaboration'
  icon: React.ComponentType<{ className?: string }>
  enabled: boolean
  channels: string[]
  conditions: {
    threshold?: number
    frequency: 'immediate' | 'hourly' | 'daily' | 'weekly'
  }
  priority: 'low' | 'medium' | 'high' | 'critical'
}

const DEFAULT_CHANNELS: NotificationChannel[] = [
  {
    id: 'email',
    name: 'Email',
    description: 'Send notifications to your email address',
    icon: Mail,
    enabled: true,
    config: {
      digest: false,
      html_format: true
    }
  },
  {
    id: 'browser',
    name: 'Browser Push',
    description: 'Show desktop notifications in your browser',
    icon: Bell,
    enabled: true
  },
  {
    id: 'sms',
    name: 'SMS',
    description: 'Send critical alerts via text message',
    icon: Smartphone,
    enabled: false,
    config: {
      phone_number: ''
    }
  },
  {
    id: 'slack',
    name: 'Slack',
    description: 'Post notifications to your Slack channel',
    icon: Slack,
    enabled: false,
    config: {
      webhook_url: '',
      channel: '#alerts'
    }
  },
  {
    id: 'webhook',
    name: 'Custom Webhook',
    description: 'Send notifications to your custom endpoint',
    icon: MessageSquare,
    enabled: false,
    config: {
      url: '',
      method: 'POST',
      headers: {}
    }
  }
]

const DEFAULT_RULES: NotificationRule[] = [
  {
    id: 'usage_quota_80',
    name: 'Usage Quota Warning',
    description: 'Alert when usage reaches 80% of monthly quota',
    category: 'usage',
    icon: Activity,
    enabled: true,
    channels: ['email', 'browser'],
    conditions: {
      threshold: 80,
      frequency: 'immediate'
    },
    priority: 'medium'
  },
  {
    id: 'usage_quota_95',
    name: 'Usage Quota Critical',
    description: 'Alert when usage reaches 95% of monthly quota',
    category: 'usage',
    icon: AlertTriangle,
    enabled: true,
    channels: ['email', 'browser', 'sms'],
    conditions: {
      threshold: 95,
      frequency: 'immediate'
    },
    priority: 'critical'
  },
  {
    id: 'cost_budget_exceeded',
    name: 'Cost Budget Exceeded',
    description: 'Alert when monthly costs exceed budget',
    category: 'billing',
    icon: DollarSign,
    enabled: true,
    channels: ['email', 'browser'],
    conditions: {
      frequency: 'immediate'
    },
    priority: 'high'
  },
  {
    id: 'api_key_expiring',
    name: 'API Key Expiring',
    description: 'Alert when API keys are about to expire (7 days)',
    category: 'security',
    icon: Shield,
    enabled: true,
    channels: ['email'],
    conditions: {
      threshold: 7,
      frequency: 'daily'
    },
    priority: 'medium'
  },
  {
    id: 'new_member_joined',
    name: 'New Team Member',
    description: 'Notify when someone joins the organization',
    category: 'collaboration',
    icon: Users,
    enabled: true,
    channels: ['email', 'browser'],
    conditions: {
      frequency: 'immediate'
    },
    priority: 'low'
  },
  {
    id: 'high_error_rate',
    name: 'High Error Rate',
    description: 'Alert when error rate exceeds 5%',
    category: 'system',
    icon: AlertTriangle,
    enabled: true,
    channels: ['email', 'browser', 'slack'],
    conditions: {
      threshold: 5,
      frequency: 'immediate'
    },
    priority: 'high'
  }
]

export function NotificationPreferences() {
  const [channels, setChannels] = useState<NotificationChannel[]>(DEFAULT_CHANNELS)
  const [rules, setRules] = useState<NotificationRule[]>(DEFAULT_RULES)
  const [globalEnabled, setGlobalEnabled] = useState(true)
  const [quietHours, setQuietHours] = useState([22, 8]) // 10 PM to 8 AM
  const [isSaving, setIsSaving] = useState(false)

  const toggleChannel = (channelId: string) => {
    setChannels(channels.map(channel =>
      channel.id === channelId ? { ...channel, enabled: !channel.enabled } : channel
    ))
  }

  const toggleRule = (ruleId: string) => {
    setRules(rules.map(rule =>
      rule.id === ruleId ? { ...rule, enabled: !rule.enabled } : rule
    ))
  }

  const updateRuleChannels = (ruleId: string, channelId: string, enabled: boolean) => {
    setRules(rules.map(rule => {
      if (rule.id === ruleId) {
        const newChannels = enabled
          ? [...rule.channels, channelId]
          : rule.channels.filter(c => c !== channelId)
        return { ...rule, channels: newChannels }
      }
      return rule
    }))
  }

  const updateRuleThreshold = (ruleId: string, threshold: number) => {
    setRules(rules.map(rule =>
      rule.id === ruleId 
        ? { ...rule, conditions: { ...rule.conditions, threshold } }
        : rule
    ))
  }

  const updateRuleFrequency = (ruleId: string, frequency: NotificationRule['conditions']['frequency']) => {
    setRules(rules.map(rule =>
      rule.id === ruleId 
        ? { ...rule, conditions: { ...rule.conditions, frequency } }
        : rule
    ))
  }

  const updateChannelConfig = (channelId: string, config: Record<string, any>) => {
    setChannels(channels.map(channel =>
      channel.id === channelId 
        ? { ...channel, config: { ...channel.config, ...config } }
        : channel
    ))
  }

  const savePreferences = async () => {
    setIsSaving(true)
    
    try {
      // TODO: Implement API call to save preferences
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      toast.success('Notification preferences saved successfully')
    } catch (error) {
      console.error('Failed to save preferences:', error)
      toast.error('Failed to save preferences. Please try again.')
    } finally {
      setIsSaving(false)
    }
  }

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'usage':
        return <Activity className="h-4 w-4" />
      case 'security':
        return <Shield className="h-4 w-4" />
      case 'billing':
        return <DollarSign className="h-4 w-4" />
      case 'system':
        return <Settings className="h-4 w-4" />
      case 'collaboration':
        return <Users className="h-4 w-4" />
      default:
        return <Bell className="h-4 w-4" />
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'low':
        return 'bg-gray-100 text-gray-800'
      case 'medium':
        return 'bg-blue-100 text-blue-800'
      case 'high':
        return 'bg-orange-100 text-orange-800'
      case 'critical':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const enabledChannels = channels.filter(c => c.enabled)
  const enabledRules = rules.filter(r => r.enabled)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Notification Preferences</h2>
          <p className="text-muted-foreground">
            Configure how and when you want to receive notifications
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2">
            {globalEnabled ? (
              <Volume2 className="h-4 w-4 text-green-500" />
            ) : (
              <VolumeX className="h-4 w-4 text-red-500" />
            )}
            <Switch
              checked={globalEnabled}
              onCheckedChange={setGlobalEnabled}
            />
            <Label>{globalEnabled ? 'Enabled' : 'Disabled'}</Label>
          </div>
          
          <Button onClick={savePreferences} disabled={isSaving}>
            {isSaving ? (
              <>
                <Settings className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="rules" className="space-y-6">
        <TabsList>
          <TabsTrigger value="rules">Notification Rules</TabsTrigger>
          <TabsTrigger value="channels">Delivery Channels</TabsTrigger>
          <TabsTrigger value="general">General Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="rules" className="space-y-6">
          {/* Rules Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Rules Overview</CardTitle>
              <CardDescription>
                {enabledRules.length} of {rules.length} notification rules are active
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4" />
                  <span>{rules.filter(r => r.category === 'usage' && r.enabled).length} Usage</span>
                </div>
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4" />
                  <span>{rules.filter(r => r.category === 'security' && r.enabled).length} Security</span>
                </div>
                <div className="flex items-center gap-2">
                  <DollarSign className="h-4 w-4" />
                  <span>{rules.filter(r => r.category === 'billing' && r.enabled).length} Billing</span>
                </div>
                <div className="flex items-center gap-2">
                  <Settings className="h-4 w-4" />
                  <span>{rules.filter(r => r.category === 'system' && r.enabled).length} System</span>
                </div>
                <div className="flex items-center gap-2">
                  <Users className="h-4 w-4" />
                  <span>{rules.filter(r => r.category === 'collaboration' && r.enabled).length} Team</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Rules Configuration */}
          <div className="space-y-4">
            {rules.map((rule) => (
              <Card key={rule.id} className={cn("transition-opacity", !globalEnabled && "opacity-50")}>
                <CardHeader className="pb-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className="mt-1">
                        <Switch
                          checked={rule.enabled}
                          onCheckedChange={() => toggleRule(rule.id)}
                          disabled={!globalEnabled}
                        />
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <rule.icon className="h-5 w-5 text-muted-foreground" />
                          <CardTitle className="text-lg">{rule.name}</CardTitle>
                          <Badge className={getPriorityColor(rule.priority)}>
                            {rule.priority}
                          </Badge>
                        </div>
                        <CardDescription>{rule.description}</CardDescription>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {getCategoryIcon(rule.category)}
                      <span className="text-sm text-muted-foreground capitalize">
                        {rule.category}
                      </span>
                    </div>
                  </div>
                </CardHeader>
                
                {rule.enabled && globalEnabled && (
                  <CardContent className="space-y-4">
                    {/* Threshold Configuration */}
                    {rule.conditions.threshold !== undefined && (
                      <div className="space-y-2">
                        <Label>Threshold: {rule.conditions.threshold}%</Label>
                        <Slider
                          value={[rule.conditions.threshold]}
                          onValueChange={([value]) => updateRuleThreshold(rule.id, value)}
                          max={100}
                          min={1}
                          step={1}
                          className="w-full"
                        />
                      </div>
                    )}

                    {/* Frequency Configuration */}
                    <div className="space-y-2">
                      <Label>Frequency</Label>
                      <Select 
                        value={rule.conditions.frequency} 
                        onValueChange={(value: any) => updateRuleFrequency(rule.id, value)}
                      >
                        <SelectTrigger className="w-48">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="immediate">Immediate</SelectItem>
                          <SelectItem value="hourly">Hourly Digest</SelectItem>
                          <SelectItem value="daily">Daily Digest</SelectItem>
                          <SelectItem value="weekly">Weekly Summary</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    {/* Channel Configuration */}
                    <div className="space-y-2">
                      <Label>Delivery Channels</Label>
                      <div className="flex flex-wrap gap-2">
                        {channels.filter(c => c.enabled).map((channel) => (
                          <div key={channel.id} className="flex items-center gap-2">
                            <Switch
                              checked={rule.channels.includes(channel.id)}
                              onCheckedChange={(checked) => updateRuleChannels(rule.id, channel.id, checked)}
                              size="sm"
                            />
                            <div className="flex items-center gap-1">
                              <channel.icon className="h-3 w-3" />
                              <span className="text-sm">{channel.name}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </CardContent>
                )}
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="channels" className="space-y-6">
          {/* Channels Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Delivery Channels</CardTitle>
              <CardDescription>
                {enabledChannels.length} of {channels.length} delivery channels are configured
              </CardDescription>
            </CardHeader>
          </Card>

          {/* Channel Configuration */}
          <div className="space-y-4">
            {channels.map((channel) => (
              <Card key={channel.id}>
                <CardHeader className="pb-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Switch
                        checked={channel.enabled}
                        onCheckedChange={() => toggleChannel(channel.id)}
                      />
                      <div className="flex items-center gap-2">
                        <channel.icon className="h-5 w-5 text-muted-foreground" />
                        <CardTitle className="text-lg">{channel.name}</CardTitle>
                      </div>
                    </div>
                  </div>
                  <CardDescription>{channel.description}</CardDescription>
                </CardHeader>

                {channel.enabled && channel.config && (
                  <CardContent className="space-y-4">
                    {channel.id === 'email' && (
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                          <Switch
                            checked={channel.config.digest}
                            onCheckedChange={(checked) => updateChannelConfig(channel.id, { digest: checked })}
                          />
                          <Label>Send daily digest instead of individual emails</Label>
                        </div>
                      </div>
                    )}

                    {channel.id === 'sms' && (
                      <div className="space-y-2">
                        <Label>Phone Number</Label>
                        <Input
                          type="tel"
                          placeholder="+1 (555) 123-4567"
                          value={channel.config.phone_number || ''}
                          onChange={(e) => updateChannelConfig(channel.id, { phone_number: e.target.value })}
                          className="w-64"
                        />
                      </div>
                    )}

                    {channel.id === 'slack' && (
                      <div className="space-y-4">
                        <div className="space-y-2">
                          <Label>Webhook URL</Label>
                          <Input
                            type="url"
                            placeholder="https://hooks.slack.com/services/..."
                            value={channel.config.webhook_url || ''}
                            onChange={(e) => updateChannelConfig(channel.id, { webhook_url: e.target.value })}
                          />
                        </div>
                        <div className="space-y-2">
                          <Label>Channel</Label>
                          <Input
                            placeholder="#alerts"
                            value={channel.config.channel || ''}
                            onChange={(e) => updateChannelConfig(channel.id, { channel: e.target.value })}
                            className="w-48"
                          />
                        </div>
                      </div>
                    )}

                    {channel.id === 'webhook' && (
                      <div className="space-y-4">
                        <div className="space-y-2">
                          <Label>Webhook URL</Label>
                          <Input
                            type="url"
                            placeholder="https://your-api.com/webhooks/notifications"
                            value={channel.config.url || ''}
                            onChange={(e) => updateChannelConfig(channel.id, { url: e.target.value })}
                          />
                        </div>
                        <div className="space-y-2">
                          <Label>HTTP Method</Label>
                          <Select 
                            value={channel.config.method || 'POST'} 
                            onValueChange={(value) => updateChannelConfig(channel.id, { method: value })}
                          >
                            <SelectTrigger className="w-32">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="POST">POST</SelectItem>
                              <SelectItem value="PUT">PUT</SelectItem>
                              <SelectItem value="PATCH">PATCH</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    )}
                  </CardContent>
                )}
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="general" className="space-y-6">
          {/* Quiet Hours */}
          <Card>
            <CardHeader>
              <CardTitle>Quiet Hours</CardTitle>
              <CardDescription>
                Set hours when non-critical notifications will be suppressed
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Quiet Hours: {quietHours[0]}:00 - {quietHours[1]}:00</Label>
                <div className="flex items-center gap-4">
                  <div className="space-y-2">
                    <Label className="text-sm">Start (PM)</Label>
                    <Slider
                      value={[quietHours[0]]}
                      onValueChange={([value]) => setQuietHours([value, quietHours[1]])}
                      max={23}
                      min={18}
                      step={1}
                      className="w-32"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label className="text-sm">End (AM)</Label>
                    <Slider
                      value={[quietHours[1]]}
                      onValueChange={([value]) => setQuietHours([quietHours[0], value])}
                      max={12}
                      min={6}
                      step={1}
                      className="w-32"
                    />
                  </div>
                </div>
                <p className="text-xs text-muted-foreground">
                  Critical alerts will still be delivered during quiet hours
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Notification History */}
          <Card>
            <CardHeader>
              <CardTitle>Recent Notifications</CardTitle>
              <CardDescription>
                Your last few notifications across all channels
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {[
                  { type: 'Usage quota warning', time: '2 hours ago', channel: 'Email', priority: 'medium' },
                  { type: 'New team member joined', time: '1 day ago', channel: 'Browser', priority: 'low' },
                  { type: 'API key expired', time: '2 days ago', channel: 'Email, SMS', priority: 'high' },
                  { type: 'Cost budget exceeded', time: '3 days ago', channel: 'Email, Slack', priority: 'critical' },
                ].map((notification, index) => (
                  <div key={index} className="flex items-center justify-between p-3 border rounded-lg">
                    <div className="flex items-center gap-3">
                      <Badge className={getPriorityColor(notification.priority)}>
                        {notification.priority}
                      </Badge>
                      <div>
                        <div className="font-medium text-sm">{notification.type}</div>
                        <div className="text-xs text-muted-foreground">{notification.channel}</div>
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {notification.time}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}