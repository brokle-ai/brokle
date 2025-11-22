'use client'

import { useState } from 'react'
import { Save, Copy, Loader2 } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { useUpdateOrganizationMutation } from '../hooks/use-organization-queries'

export function OrganizationGeneralSection() {
  const { currentOrganization } = useWorkspace()
  const updateMutation = useUpdateOrganizationMutation()

  // Track user edits (null = not edited, use currentOrganization value)
  const [editedName, setEditedName] = useState<string | null>(null)

  // Derive display value
  const organizationName = editedName ?? currentOrganization?.name ?? ''

  if (!currentOrganization) {
    return null
  }

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!currentOrganization) return

    // Validation
    if (organizationName.trim().length < 2 || organizationName.trim().length > 100) {
      toast.error('Organization name must be between 2 and 100 characters')
      return
    }

    try {
      await updateMutation.mutateAsync({
        orgId: currentOrganization.id,
        data: {
          name: organizationName.trim()
        }
      })

      // Clear edit tracking after success
      setEditedName(null)
    } catch (error) {
      console.error('Failed to update organization:', error)
    }
  }

  const copyOrganizationId = () => {
    navigator.clipboard.writeText(currentOrganization.id)
    toast.success('Organization ID copied to clipboard')
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
            onChange={(e) => setEditedName(e.target.value)}
            placeholder="Enter organization name"
            required
          />
        </div>
      </div>

      {/* Organization Identification */}
      <div className="rounded-lg border p-4">
        <div className="space-y-2">
          <div className="text-sm font-medium text-muted-foreground">Organization ID</div>
          <div className="flex items-center gap-2">
            <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
              {currentOrganization.id}
            </code>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={copyOrganizationId}
            >
              <Copy className="h-3 w-3" />
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Use this ID for API integration and support requests
          </p>
        </div>
      </div>

      {/* Submit Button */}
      <Button type="submit" disabled={updateMutation.isPending}>
        {updateMutation.isPending ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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
