'use client'

import { useState } from 'react'
import { Settings, Save, RefreshCw, AlertCircle } from 'lucide-react'
import { useOrganization } from '@/context/org-context'
import { getOrgSlug, getProjectSlug } from '@/lib/utils/slug-utils'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { TabsContent } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import type { ProjectStatus, ProjectEnvironment } from '@/types/organization'

export default function ProjectGeneralSettingsPage() {
  const { currentProject, currentOrganization } = useOrganization()
  
  const [isLoading, setIsLoading] = useState(false)
  const [projectName, setProjectName] = useState(currentProject?.name || '')
  const [projectDescription, setProjectDescription] = useState(currentProject?.description || '')
  const computedProjectSlug = currentProject ? getProjectSlug(currentProject) : ''
  const [projectStatus, setProjectStatus] = useState<ProjectStatus>(currentProject?.status || 'active')
  const [environment, setEnvironment] = useState<ProjectEnvironment>(currentProject?.environment || 'production')
  const [isPublic, setIsPublic] = useState(currentProject?.settings?.public || false)
  const [webhookUrl, setWebhookUrl] = useState(currentProject?.settings?.webhook_url || '')
  const [retryAttempts, setRetryAttempts] = useState(currentProject?.settings?.retry_attempts?.toString() || '3')
  const [timeoutMs, setTimeoutMs] = useState(currentProject?.settings?.timeout_ms?.toString() || '30000')

  if (!currentProject || !currentOrganization) {
    return null
  }

  const handleSaveSettings = async () => {
    setIsLoading(true)
    
    try {
      // TODO: Implement API call to update project settings
      await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
      
      toast.success('Project settings updated successfully')
    } catch (error) {
      console.error('Failed to update project settings:', error)
      toast.error('Failed to update settings. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const regenerateSlug = () => {
    const newSlug = projectName.toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .trim()
    
    setProjectSlug(newSlug)
    toast.success('Slug regenerated from project name')
  }

  const getStatusColor = (status: ProjectStatus) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'inactive':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'archived':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  return (
    <TabsContent value="general" className="space-y-6">
      {/* Project Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Project Information
          </CardTitle>
          <CardDescription>
            Basic information and configuration for this project
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="projectName">Project Name *</Label>
              <Input
                id="projectName"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                placeholder="Enter project name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="projectSlug">
                Project Slug *
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="ml-2 h-6 px-2"
                  onClick={regenerateSlug}
                >
                  <RefreshCw className="h-3 w-3" />
                </Button>
              </Label>
              <Input
                id="projectSlug"
                value={projectSlug}
                onChange={(e) => setProjectSlug(e.target.value)}
                placeholder="project-slug"
              />
              <p className="text-xs text-muted-foreground">
                URL: /organizations/{getOrgSlug(currentOrganization)}/projects/{computedProjectSlug}
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="projectDescription">Description</Label>
            <Textarea
              id="projectDescription"
              value={projectDescription}
              onChange={(e) => setProjectDescription(e.target.value)}
              placeholder="Describe what this project is for..."
              rows={3}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="projectStatus">Status</Label>
              <Select value={projectStatus} onValueChange={(value: ProjectStatus) => setProjectStatus(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                  <SelectItem value="archived">Archived</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="environment">Environment</Label>
              <Select value={environment} onValueChange={(value: ProjectEnvironment) => setEnvironment(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="development">Development</SelectItem>
                  <SelectItem value="staging">Staging</SelectItem>
                  <SelectItem value="production">Production</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Public Project</Label>
              <p className="text-sm text-muted-foreground">
                Allow other organization members to view this project
              </p>
            </div>
            <Switch
              checked={isPublic}
              onCheckedChange={setIsPublic}
            />
          </div>
        </CardContent>
      </Card>

      {/* API Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>API Configuration</CardTitle>
          <CardDescription>
            Configure API behavior and reliability settings
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="webhookUrl">Webhook URL</Label>
            <Input
              id="webhookUrl"
              value={webhookUrl}
              onChange={(e) => setWebhookUrl(e.target.value)}
              placeholder="https://your-app.com/webhooks/brokle"
              type="url"
            />
            <p className="text-xs text-muted-foreground">
              Receive real-time notifications about project events
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="retryAttempts">Retry Attempts</Label>
              <Input
                id="retryAttempts"
                value={retryAttempts}
                onChange={(e) => setRetryAttempts(e.target.value)}
                placeholder="3"
                type="number"
                min="0"
                max="10"
              />
              <p className="text-xs text-muted-foreground">
                Number of retry attempts for failed requests
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="timeoutMs">Request Timeout (ms)</Label>
              <Input
                id="timeoutMs"
                value={timeoutMs}
                onChange={(e) => setTimeoutMs(e.target.value)}
                placeholder="30000"
                type="number"
                min="1000"
                max="300000"
              />
              <p className="text-xs text-muted-foreground">
                Maximum time to wait for API responses
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Project Status Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Project Overview</CardTitle>
          <CardDescription>
            Current project status and key information
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-6">
            <div>
              <div className="text-sm font-medium text-muted-foreground">Current Status</div>
              <Badge className={getStatusColor(currentProject.status)}>
                {currentProject.status.charAt(0).toUpperCase() + currentProject.status.slice(1)}
              </Badge>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Environment</div>
              <Badge variant="outline">
                {currentProject.environment.charAt(0).toUpperCase() + currentProject.environment.slice(1)}
              </Badge>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Created</div>
              <div className="text-sm">{new Date(currentProject.created_at).toLocaleDateString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Last Updated</div>
              <div className="text-sm">{new Date(currentProject.updated_at).toLocaleDateString()}</div>
            </div>
          </div>

          <Separator />

          <div>
            <div className="text-sm font-medium text-muted-foreground mb-2">Project ID</div>
            <div className="flex items-center gap-2">
              <code className="text-xs bg-muted px-2 py-1 rounded">{currentProject.id}</code>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  navigator.clipboard.writeText(currentProject.id)
                  toast.success('Project ID copied to clipboard')
                }}
              >
                Copy
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Warning for Status Changes */}
      {projectStatus !== currentProject.status && projectStatus === 'archived' && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Archiving this project will make it read-only and stop all API requests. 
            You can reactivate it later if needed.
          </AlertDescription>
        </Alert>
      )}

      {/* Save Changes */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">Save Changes</p>
              <p className="text-xs text-muted-foreground">
                Update your project settings and configuration
              </p>
            </div>
            <Button onClick={handleSaveSettings} disabled={isLoading}>
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
          </div>
        </CardContent>
      </Card>
    </TabsContent>
  )
}