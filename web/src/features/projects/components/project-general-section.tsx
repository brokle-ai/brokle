'use client'

import { useState } from 'react'
import { Save, AlertCircle, Copy, Loader2 } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import { useUpdateProjectMutation } from '../hooks/use-project-queries'
import type { ProjectStatus } from '@/features/organizations'

export function ProjectGeneralSection() {
  const { currentProject, currentOrganization } = useWorkspace()
  const updateMutation = useUpdateProjectMutation()

  // Track user edits (null = not edited, use currentProject value)
  const [editedName, setEditedName] = useState<string | null>(null)
  const [editedDescription, setEditedDescription] = useState<string | null>(null)
  const [editedStatus, setEditedStatus] = useState<ProjectStatus | null>(null)

  // Derive display values: use edited value if exists, otherwise currentProject value
  const projectName = editedName ?? currentProject?.name ?? ''
  const projectDescription = editedDescription ?? currentProject?.description ?? ''
  const projectStatus = editedStatus ?? currentProject?.status ?? 'active'

  if (!currentProject || !currentOrganization) {
    return null
  }

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault()

    // Validation
    if (projectName.trim().length < 2 || projectName.trim().length > 100) {
      toast.error('Project name must be between 2 and 100 characters')
      return
    }

    if (projectDescription.length > 500) {
      toast.error('Description must be less than 500 characters')
      return
    }

    try {
      await updateMutation.mutateAsync({
        projectId: currentProject.id,
        data: {
          name: projectName.trim(),
          description: projectDescription.trim(),
          status: projectStatus as 'active' | 'paused' | 'archived'
        }
      })

      // Clear edit tracking after successful save
      setEditedName(null)
      setEditedDescription(null)
      setEditedStatus(null)
    } catch (error) {
      // Error toast handled by mutation hook
      console.error('Failed to update project:', error)
    }
  }

  const copyProjectId = () => {
    navigator.clipboard.writeText(currentProject.id)
    toast.success('Project ID copied to clipboard')
  }

  const getStatusColor = (status: ProjectStatus) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'paused':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'archived':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  return (
    <form onSubmit={handleSaveSettings} className="space-y-8">
      {/* Project Information Section */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="projectName">Project Name *</Label>
          <Input
            id="projectName"
            value={projectName}
            onChange={(e) => setEditedName(e.target.value)}
            placeholder="Enter project name"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="projectDescription">Description</Label>
          <Textarea
            id="projectDescription"
            value={projectDescription}
            onChange={(e) => setEditedDescription(e.target.value)}
            placeholder="Describe what this project is for..."
            rows={3}
          />
        </div>
      </div>

      {/* Status Section */}
      <div className="space-y-2">
        <Label htmlFor="projectStatus">Status</Label>
        <Select value={projectStatus} onValueChange={(value: ProjectStatus) => setEditedStatus(value)}>
          <SelectTrigger id="projectStatus" className="max-w-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="paused">Paused</SelectItem>
            <SelectItem value="archived">Archived</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Project Information Display */}
      <div className="rounded-lg border p-4 space-y-4">
        <div className="grid grid-cols-3 gap-6">
          <div>
            <div className="text-sm font-medium text-muted-foreground">Current Status</div>
            <Badge className={getStatusColor(currentProject.status || 'active')}>
              {currentProject.status
                ? currentProject.status.charAt(0).toUpperCase() + currentProject.status.slice(1)
                : 'Active'}
            </Badge>
          </div>
          <div>
            <div className="text-sm font-medium text-muted-foreground">Created</div>
            <div className="text-sm">{new Date(currentProject.createdAt).toLocaleDateString()}</div>
          </div>
          <div>
            <div className="text-sm font-medium text-muted-foreground">Last Updated</div>
            <div className="text-sm">{new Date(currentProject.updatedAt).toLocaleDateString()}</div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="text-sm font-medium text-muted-foreground mb-2">Project ID</div>
          <div className="flex items-center gap-2">
            <code className="text-xs bg-muted px-2 py-1 rounded">{currentProject.id}</code>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={copyProjectId}
            >
              <Copy className="h-3 w-3 mr-1" />
              Copy
            </Button>
          </div>
        </div>
      </div>

      {/* Status Change Warning */}
      {projectStatus !== currentProject.status && projectStatus === 'archived' && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Archiving this project will make it read-only and stop all API requests.
            You can reactivate it later if needed.
          </AlertDescription>
        </Alert>
      )}

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
