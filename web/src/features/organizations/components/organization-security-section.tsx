'use client'

import { useState } from 'react'
import { Shield, Lock, Clock, Globe, AlertTriangle } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { toast } from 'sonner'

interface SecuritySetting {
  id: string
  name: string
  description: string
  enabled: boolean
  enterpriseOnly?: boolean
}

const ORG_SECURITY_SETTINGS: SecuritySetting[] = [
  {
    id: 'enforce_2fa',
    name: 'Enforce Two-Factor Authentication',
    description: 'Require all organization members to enable 2FA',
    enabled: false,
  },
  {
    id: 'sso_enabled',
    name: 'Single Sign-On (SSO)',
    description: 'Enable SAML/OIDC authentication for organization members',
    enabled: false,
    enterpriseOnly: true,
  },
  {
    id: 'ip_whitelist',
    name: 'IP Address Whitelist',
    description: 'Restrict access to specific IP addresses',
    enabled: false,
  },
  {
    id: 'audit_logging',
    name: 'Enhanced Audit Logging',
    description: 'Log all organization-level actions and changes',
    enabled: true,
  },
]

export function OrganizationSecuritySection() {
  const { currentOrganization } = useWorkspace()
  const [securitySettings, setSecuritySettings] = useState<SecuritySetting[]>(ORG_SECURITY_SETTINGS)
  const [sessionTimeout, setSessionTimeout] = useState('24h')

  if (!currentOrganization) {
    return null
  }

  const toggleSecuritySetting = (settingId: string) => {
    const setting = securitySettings.find(s => s.id === settingId)
    if (setting?.enterpriseOnly && currentOrganization.plan !== 'enterprise') {
      toast.error('This feature is only available on the Enterprise plan')
      return
    }

    setSecuritySettings(settings =>
      settings.map(s =>
        s.id === settingId ? { ...s, enabled: !s.enabled } : s
      )
    )
    toast.success('Security setting updated')
  }

  const isEnterprise = currentOrganization.plan === 'enterprise'

  return (
    <div className="space-y-8">
      {/* Security Settings */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Security Features</h3>

        {securitySettings.map((setting) => (
          <div key={setting.id} className="flex items-start justify-between rounded-lg border p-4">
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <div className="font-medium">{setting.name}</div>
                {setting.enterpriseOnly && (
                  <Badge variant="outline" className="text-xs">
                    Enterprise
                  </Badge>
                )}
              </div>
              <div className="text-sm text-muted-foreground">
                {setting.description}
              </div>
            </div>
            <Switch
              checked={setting.enabled}
              onCheckedChange={() => toggleSecuritySetting(setting.id)}
              disabled={setting.enterpriseOnly && !isEnterprise}
            />
          </div>
        ))}
      </div>

      {/* Session Management */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Session Management</h3>

        <div className="space-y-2">
          <Label htmlFor="sessionTimeout">Session Timeout</Label>
          <Select value={sessionTimeout} onValueChange={setSessionTimeout}>
            <SelectTrigger id="sessionTimeout">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1h">1 hour</SelectItem>
              <SelectItem value="8h">8 hours</SelectItem>
              <SelectItem value="24h">24 hours (recommended)</SelectItem>
              <SelectItem value="7d">7 days</SelectItem>
              <SelectItem value="30d">30 days</SelectItem>
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            Automatically log out inactive users after this duration
          </p>
        </div>

        <Button variant="outline">
          <Clock className="mr-2 h-4 w-4" />
          Invalidate All Sessions
        </Button>
      </div>

      {/* SSO Configuration (Enterprise) */}
      {isEnterprise && (
        <div className="space-y-4">
          <h3 className="text-lg font-medium">Single Sign-On Configuration</h3>

          <div className="rounded-lg border p-4 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-muted rounded-lg">
                  <Shield className="h-5 w-5" />
                </div>
                <div>
                  <div className="font-medium">SAML 2.0</div>
                  <div className="text-sm text-muted-foreground">Configure SAML identity provider</div>
                </div>
              </div>
              <Button variant="outline" size="sm">
                Configure
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Security Recommendations */}
      {!isEnterprise && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            <strong>Upgrade to Enterprise</strong> to unlock advanced security features like SSO, advanced audit logging, and custom security policies.
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
