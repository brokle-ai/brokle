'use client'

import { useState } from 'react'
import { Save, RefreshCw, Copy } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { getOrgSlug } from '@/lib/utils/slug-utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'

export function OrganizationGeneralSection() {
  const { currentOrganization } = useWorkspace()

  const [isLoading, setIsLoading] = useState(false)
  const [organizationName, setOrganizationName] = useState(currentOrganization?.name || '')
  const [billingEmail, setBillingEmail] = useState(currentOrganization?.billing_email || '')

  if (!currentOrganization) {
    return null
  }

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      // TODO: Implement API call to update organization
      await new Promise(resolve => setTimeout(resolve, 1000))

      toast.success('Organization settings updated successfully')
    } catch (error) {
      console.error('Failed to update organization settings:', error)
      toast.error('Failed to update settings. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const copyOrganizationId = () => {
    navigator.clipboard.writeText(currentOrganization.id)
    toast.success('Organization ID copied to clipboard')
  }

  const getPlanColor = (plan: string) => {
    switch (plan) {
      case 'enterprise':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
      case 'business':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'pro':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'free':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  return (
    <form onSubmit={handleSaveSettings} className="space-y-8">
      {/* Organization Information */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="organizationName">Organization Name *</Label>
          <Input
            id="organizationName"
            value={organizationName}
            onChange={(e) => setOrganizationName(e.target.value)}
            placeholder="Enter organization name"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="organizationSlug">Organization Slug</Label>
          <Input
            id="organizationSlug"
            value={getOrgSlug(currentOrganization)}
            readOnly
            disabled
            className="bg-muted cursor-not-allowed"
          />
          <p className="text-xs text-muted-foreground">
            URL: /organizations/{getOrgSlug(currentOrganization)}
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="billingEmail">Billing Email</Label>
          <Input
            id="billingEmail"
            type="email"
            value={billingEmail}
            onChange={(e) => setBillingEmail(e.target.value)}
            placeholder="billing@example.com"
          />
          <p className="text-xs text-muted-foreground">
            Invoices and billing notifications will be sent to this email
          </p>
        </div>
      </div>

      {/* Subscription Information */}
      <div className="rounded-lg border p-4 space-y-4">
        <div className="grid grid-cols-2 gap-6">
          <div>
            <div className="text-sm font-medium text-muted-foreground">Current Plan</div>
            <Badge className={getPlanColor(currentOrganization.plan)}>
              {currentOrganization.plan.charAt(0).toUpperCase() + currentOrganization.plan.slice(1)}
            </Badge>
          </div>
          <div>
            <div className="text-sm font-medium text-muted-foreground">Subscription Status</div>
            <Badge variant="outline">Active</Badge>
          </div>
          <div>
            <div className="text-sm font-medium text-muted-foreground">Created</div>
            <div className="text-sm">{new Date(currentOrganization.created_at).toLocaleDateString()}</div>
          </div>
          <div>
            <div className="text-sm font-medium text-muted-foreground">Last Updated</div>
            <div className="text-sm">{new Date(currentOrganization.updated_at).toLocaleDateString()}</div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="text-sm font-medium text-muted-foreground mb-2">Organization ID</div>
          <div className="flex items-center gap-2">
            <code className="text-xs bg-muted px-2 py-1 rounded">{currentOrganization.id}</code>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={copyOrganizationId}
            >
              <Copy className="h-3 w-3 mr-1" />
              Copy
            </Button>
          </div>
        </div>
      </div>

      {/* Submit Button */}
      <Button type="submit" disabled={isLoading}>
        {isLoading ? (
          <>
            <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
            Saving...
          </>
        ) : (
          <>
            <Save className="mr-2 h-4 w-4" />
            Save Changes
          </>
        )}
      </Button>
    </form>
  )
}
