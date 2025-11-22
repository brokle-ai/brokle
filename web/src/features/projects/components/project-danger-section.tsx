'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { AlertTriangle, Trash2, Download, Archive, RefreshCw, AlertCircle } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { getOrgSlug, getProjectSlug } from '@/lib/utils/slug-utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import { archiveProject, unarchiveProject } from '../api/projects-api'
import { useQueryClient } from '@tanstack/react-query'

export function ProjectDangerSection() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const { currentProject, currentOrganization } = useWorkspace()

  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isArchiveOpen, setIsArchiveOpen] = useState(false)
  const [isUnarchiveOpen, setIsUnarchiveOpen] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [acknowledgedRisks, setAcknowledgedRisks] = useState<string[]>([])
  const [isDeleting, setIsDeleting] = useState(false)
  const [isExporting, setIsExporting] = useState(false)

  if (!currentProject || !currentOrganization) {
    return null
  }

  const handleDeleteProject = async () => {
    if (deleteConfirmation !== currentProject.name) {
      toast.error('Project name does not match')
      return
    }

    if (acknowledgedRisks.length < 4) {
      toast.error('Please acknowledge all risks before proceeding')
      return
    }

    setIsDeleting(true)

    try {
      // TODO: Implement API call to delete project
      await new Promise(resolve => setTimeout(resolve, 2000))

      toast.success('Project deleted successfully')
      router.push(`/organizations/${getOrgSlug(currentOrganization)}`)
      setIsDeleteOpen(false)
    } catch (error) {
      console.error('Failed to delete project:', error)
      toast.error('Failed to delete project. Please try again.')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleArchiveProject = async () => {
    if (!currentProject) return

    try {
      await archiveProject(currentProject.id)
      toast.success('Project Archived!', {
        description: `${currentProject.name} has been archived successfully.`
      })
      setIsArchiveOpen(false)
      // Refresh workspace to show updated status
      queryClient.invalidateQueries({ queryKey: ['workspace'] })
    } catch (error) {
      console.error('Failed to archive project:', error)
      toast.error('Failed to Archive Project', {
        description: 'Could not archive project. Please try again.'
      })
    }
  }

  const handleUnarchiveProject = async () => {
    if (!currentProject) return

    try {
      await unarchiveProject(currentProject.id)
      toast.success('Project Unarchived!', {
        description: `${currentProject.name} is now active.`
      })
      setIsUnarchiveOpen(false)
      // Refresh workspace to show updated status
      queryClient.invalidateQueries({ queryKey: ['workspace'] })
    } catch (error) {
      console.error('Failed to unarchive project:', error)
      toast.error('Failed to Unarchive Project', {
        description: 'Could not unarchive project. Please try again.'
      })
    }
  }

  const handleExportData = async () => {
    setIsExporting(true)

    try {
      // TODO: Implement API call to export project data
      await new Promise(resolve => setTimeout(resolve, 2000))

      // Simulate file download
      const blob = new Blob([JSON.stringify({
        project: currentProject,
        exported_at: new Date().toISOString(),
        data: {
          logs: 'Request logs would be here...',
          settings: 'Project settings would be here...'
        }
      }, null, 2)], { type: 'application/json' })

      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${getProjectSlug(currentProject)}-export-${new Date().toISOString().split('T')[0]}.json`
      a.click()
      URL.revokeObjectURL(url)

      toast.success('Project data exported successfully')
    } catch (error) {
      console.error('Failed to export data:', error)
      toast.error('Failed to export data. Please try again.')
    } finally {
      setIsExporting(false)
    }
  }

  const handleRiskAcknowledgment = (riskId: string, checked: boolean) => {
    if (checked) {
      setAcknowledgedRisks([...acknowledgedRisks, riskId])
    } else {
      setAcknowledgedRisks(acknowledgedRisks.filter(id => id !== riskId))
    }
  }

  const deleteRisks = [
    { id: 'data-loss', text: 'All project data and logs will be permanently deleted' },
    { id: 'api-keys', text: 'All API keys for this project will be immediately revoked' },
    { id: 'billing-data', text: 'Billing and usage history will be removed' },
    { id: 'irreversible', text: 'This action cannot be undone' }
  ]

  return (
    <div className="space-y-6">
      <Alert>
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>
          <strong>Warning:</strong> These actions are irreversible and can cause permanent data loss.
          Please proceed with extreme caution.
        </AlertDescription>
      </Alert>

      {/* Export Data */}
      <div className="rounded-lg border p-4 space-y-4">
        <div>
          <h4 className="font-medium mb-2 flex items-center gap-2">
            <Download className="h-4 w-4 text-blue-500" />
            Export Project Data
          </h4>
          <p className="text-sm text-muted-foreground mb-4">
            Download a complete backup of your project data before making destructive changes
          </p>
        </div>

        <div className="text-sm text-muted-foreground">
          Export includes:
          <ul className="list-disc list-inside mt-2 space-y-1">
            <li>Project configuration and settings</li>
            <li>Request logs and metrics</li>
            <li>API key metadata (keys themselves are not exported)</li>
          </ul>
        </div>

        <Button onClick={handleExportData} disabled={isExporting} className="w-full">
          {isExporting ? (
            <>
              <Download className="mr-2 h-4 w-4 animate-bounce" />
              Exporting Data...
            </>
          ) : (
            <>
              <Download className="mr-2 h-4 w-4" />
              Export Project Data
            </>
          )}
        </Button>
      </div>

      {/* Show Unarchive button if project is archived */}
      {currentProject.status === 'archived' && (
        <div className="rounded-lg border border-blue-200 p-4 space-y-4 bg-blue-50 dark:bg-blue-950/20">
          <div>
            <h4 className="font-medium mb-2 text-blue-600 flex items-center gap-2">
              <RefreshCw className="h-4 w-4" />
              Unarchive Project
            </h4>
            <p className="text-sm text-blue-600 mb-4">
              Restore this archived project to active status. All features will be re-enabled.
            </p>
          </div>

          <div className="text-sm text-blue-600">
            Unarchiving will:
            <ul className="list-disc list-inside mt-2 space-y-1">
              <li>Restore the project to active status</li>
              <li>Re-enable all API requests</li>
              <li>Allow creating new API keys</li>
              <li>Enable project editing and updates</li>
              <li>Resume normal project operations</li>
            </ul>
          </div>

          <Dialog open={isUnarchiveOpen} onOpenChange={setIsUnarchiveOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full border-blue-200 text-blue-700 hover:bg-blue-50">
                <RefreshCw className="mr-2 h-4 w-4" />
                Unarchive Project
              </Button>
            </DialogTrigger>

            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle>Unarchive Project</DialogTitle>
                <DialogDescription>
                  Are you sure you want to restore "{currentProject.name}" to active status?
                </DialogDescription>
              </DialogHeader>

              <Alert className="border-blue-200 bg-blue-50">
                <AlertCircle className="h-4 w-4 text-blue-500" />
                <AlertDescription className="text-blue-600">
                  This will restore full access to the project. You'll be able to create API keys and make requests again.
                </AlertDescription>
              </Alert>

              <DialogFooter>
                <Button variant="outline" onClick={() => setIsUnarchiveOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleUnarchiveProject}>
                  Unarchive Project
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      )}

      {/* Show Archive button if project is active */}
      {currentProject.status === 'active' && (
        <div className="rounded-lg border border-yellow-200 p-4 space-y-4">
          <div>
            <h4 className="font-medium mb-2 flex items-center gap-2">
              <Archive className="h-4 w-4 text-yellow-500" />
              Archive Project
            </h4>
            <p className="text-sm text-muted-foreground mb-4">
              Archive this project to make it read-only while preserving all data
            </p>
          </div>

          <div className="text-sm text-muted-foreground">
            Archiving will:
            <ul className="list-disc list-inside mt-2 space-y-1">
              <li>Stop all API requests to this project</li>
              <li>Make the project read-only</li>
              <li>Preserve all data and analytics</li>
              <li>Allow data export and viewing</li>
              <li>Can be reversed by unarchiving</li>
            </ul>
          </div>

          <Dialog open={isArchiveOpen} onOpenChange={setIsArchiveOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full border-yellow-200 text-yellow-700 hover:bg-yellow-50">
                <Archive className="mr-2 h-4 w-4" />
                Archive Project
              </Button>
            </DialogTrigger>

            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle>Archive Project</DialogTitle>
                <DialogDescription>
                  Are you sure you want to archive "{currentProject.name}"?
                </DialogDescription>
              </DialogHeader>

              <Alert className="border-yellow-200 bg-yellow-50">
                <AlertCircle className="h-4 w-4 text-yellow-500" />
                <AlertDescription className="text-yellow-600">
                  This will make the project read-only. You can unarchive it later to restore full access.
                </AlertDescription>
              </Alert>

              <DialogFooter>
                <Button variant="outline" onClick={() => setIsArchiveOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleArchiveProject} className="bg-yellow-600 hover:bg-yellow-700">
                  Archive Project
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      )}

      <Separator />

      {/* Delete Project */}
      <div className="rounded-lg border border-red-200 p-4 space-y-4">
        <div>
          <h4 className="font-medium mb-2 text-red-600 flex items-center gap-2">
            <Trash2 className="h-4 w-4" />
            Delete Project Permanently
          </h4>
          <p className="text-sm text-red-600 mb-4">
            Permanently delete this project and all associated data. This action cannot be undone.
          </p>
        </div>

        <div className="text-sm text-red-600">
          <strong>This will permanently delete:</strong>
          <ul className="list-disc list-inside mt-2 space-y-1">
            <li>All project data and configuration</li>
            <li>All API keys and access tokens</li>
            <li>Request logs and audit trails</li>
          </ul>
        </div>

        <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
          <DialogTrigger asChild>
            <Button variant="destructive" className="w-full">
              <Trash2 className="mr-2 h-4 w-4" />
              Delete Project Permanently
            </Button>
          </DialogTrigger>

          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle className="text-red-600">Delete Project</DialogTitle>
              <DialogDescription>
                This action will permanently delete "{currentProject.name}" and all associated data.
                This cannot be undone.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6">
              <div className="space-y-2">
                <Label>Type the project name to confirm deletion</Label>
                <Input
                  value={deleteConfirmation}
                  onChange={(e) => setDeleteConfirmation(e.target.value)}
                  placeholder={currentProject.name}
                />
              </div>

              <div className="space-y-3">
                <Label>Acknowledge the following risks:</Label>
                {deleteRisks.map((risk) => (
                  <div key={risk.id} className="flex items-start space-x-2">
                    <Checkbox
                      id={risk.id}
                      checked={acknowledgedRisks.includes(risk.id)}
                      onCheckedChange={(checked) => handleRiskAcknowledgment(risk.id, checked as boolean)}
                    />
                    <Label htmlFor={risk.id} className="text-sm leading-relaxed">
                      {risk.text}
                    </Label>
                  </div>
                ))}
              </div>

              <Alert className="border-red-200">
                <AlertTriangle className="h-4 w-4 text-red-500" />
                <AlertDescription className="text-red-600">
                  <strong>Final Warning:</strong> Once deleted, this project and all its data will be gone forever.
                  Consider exporting your data first.
                </AlertDescription>
              </Alert>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDeleteOpen(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteProject}
                disabled={isDeleting || deleteConfirmation !== currentProject.name || acknowledgedRisks.length < 4}
              >
                {isDeleting ? 'Deleting...' : 'Delete Forever'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}
