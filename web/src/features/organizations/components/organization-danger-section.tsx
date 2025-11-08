'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { AlertTriangle, Trash2, Download, LogOut } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'

export function OrganizationDangerSection() {
  const router = useRouter()
  const { currentOrganization } = useWorkspace()

  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isLeaveOpen, setIsLeaveOpen] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [acknowledgedRisks, setAcknowledgedRisks] = useState<string[]>([])
  const [isDeleting, setIsDeleting] = useState(false)
  const [isLeaving, setIsLeaving] = useState(false)
  const [isExporting, setIsExporting] = useState(false)

  if (!currentOrganization) {
    return null
  }

  const handleDeleteOrganization = async () => {
    if (deleteConfirmation !== currentOrganization.name) {
      toast.error('Organization name does not match')
      return
    }

    if (acknowledgedRisks.length < 4) {
      toast.error('Please acknowledge all risks before proceeding')
      return
    }

    setIsDeleting(true)

    try {
      // TODO: Implement API call to delete organization
      await new Promise(resolve => setTimeout(resolve, 2000))

      toast.success('Organization deleted successfully')
      router.push('/')
      setIsDeleteOpen(false)
    } catch (error) {
      console.error('Failed to delete organization:', error)
      toast.error('Failed to delete organization. Please try again.')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleLeaveOrganization = async () => {
    setIsLeaving(true)

    try {
      // TODO: Implement API call to leave organization
      await new Promise(resolve => setTimeout(resolve, 1500))

      toast.success('You have left the organization')
      router.push('/')
      setIsLeaveOpen(false)
    } catch (error) {
      console.error('Failed to leave organization:', error)
      toast.error('Failed to leave organization. Please try again.')
    } finally {
      setIsLeaving(false)
    }
  }

  const handleExportData = async () => {
    setIsExporting(true)

    try {
      // TODO: Implement API call to export organization data
      await new Promise(resolve => setTimeout(resolve, 2000))

      const blob = new Blob([JSON.stringify({
        organization: currentOrganization,
        exported_at: new Date().toISOString(),
        data: {
          members: 'Member data would be here...',
          projects: 'Project data would be here...',
          settings: 'Settings would be here...'
        }
      }, null, 2)], { type: 'application/json' })

      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${currentOrganization.name}-export-${new Date().toISOString().split('T')[0]}.json`
      a.click()
      URL.revokeObjectURL(url)

      toast.success('Organization data exported successfully')
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
    { id: 'all-data', text: 'All organization data, projects, and members will be permanently deleted' },
    { id: 'projects', text: 'All projects and their API keys will be removed' },
    { id: 'billing', text: 'Billing history and subscription will be cancelled' },
    { id: 'irreversible', text: 'This action cannot be undone' }
  ]

  // Mock: Check if current user is owner (you'll need real logic here)
  const isOwner = true

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
            Export Organization Data
          </h4>
          <p className="text-sm text-muted-foreground mb-4">
            Download a complete backup of your organization data
          </p>
        </div>

        <div className="text-sm text-muted-foreground">
          Export includes:
          <ul className="list-disc list-inside mt-2 space-y-1">
            <li>Organization configuration and settings</li>
            <li>Members list and roles</li>
            <li>Projects and API keys metadata</li>
            <li>Billing and usage data</li>
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
              Export Organization Data
            </>
          )}
        </Button>
      </div>

      {/* Leave Organization (Non-owners) */}
      {!isOwner && (
        <div className="rounded-lg border border-yellow-200 p-4 space-y-4">
          <div>
            <h4 className="font-medium mb-2 flex items-center gap-2">
              <LogOut className="h-4 w-4 text-yellow-500" />
              Leave Organization
            </h4>
            <p className="text-sm text-muted-foreground mb-4">
              Remove yourself from this organization
            </p>
          </div>

          <Dialog open={isLeaveOpen} onOpenChange={setIsLeaveOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full border-yellow-200 text-yellow-700 hover:bg-yellow-50">
                <LogOut className="mr-2 h-4 w-4" />
                Leave Organization
              </Button>
            </DialogTrigger>

            <DialogContent>
              <DialogHeader>
                <DialogTitle>Leave Organization</DialogTitle>
                <DialogDescription>
                  Are you sure you want to leave "{currentOrganization.name}"? You'll lose access to all projects and data.
                </DialogDescription>
              </DialogHeader>

              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  You'll need to be re-invited by an owner or admin to regain access.
                </AlertDescription>
              </Alert>

              <DialogFooter>
                <Button variant="outline" onClick={() => setIsLeaveOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleLeaveOrganization}
                  disabled={isLeaving}
                  className="bg-yellow-600 hover:bg-yellow-700"
                >
                  {isLeaving ? 'Leaving...' : 'Leave Organization'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      )}

      <Separator />

      {/* Delete Organization (Owner only) */}
      {isOwner && (
        <div className="rounded-lg border border-red-200 p-4 space-y-4">
          <div>
            <h4 className="font-medium mb-2 text-red-600 flex items-center gap-2">
              <Trash2 className="h-4 w-4" />
              Delete Organization Permanently
            </h4>
            <p className="text-sm text-red-600 mb-4">
              Permanently delete this organization and all associated data. This action cannot be undone.
            </p>
          </div>

          <div className="text-sm text-red-600">
            <strong>This will permanently delete:</strong>
            <ul className="list-disc list-inside mt-2 space-y-1">
              <li>Organization and all configuration</li>
              <li>All projects and their API keys</li>
              <li>All members and pending invitations</li>
              <li>Analytics, metrics, and usage history</li>
              <li>Billing and subscription data</li>
            </ul>
          </div>

          <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
            <DialogTrigger asChild>
              <Button variant="destructive" className="w-full">
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Organization Permanently
              </Button>
            </DialogTrigger>

            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle className="text-red-600">Delete Organization</DialogTitle>
                <DialogDescription>
                  This action will permanently delete "{currentOrganization.name}" and all associated data.
                  This cannot be undone.
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-6">
                <div className="space-y-2">
                  <Label>Type the organization name to confirm deletion</Label>
                  <Input
                    value={deleteConfirmation}
                    onChange={(e) => setDeleteConfirmation(e.target.value)}
                    placeholder={currentOrganization.name}
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
                    <strong>Final Warning:</strong> Once deleted, this organization and all its data will be gone forever.
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
                  onClick={handleDeleteOrganization}
                  disabled={isDeleting || deleteConfirmation !== currentOrganization.name || acknowledgedRisks.length < 4}
                >
                  {isDeleting ? 'Deleting...' : 'Delete Forever'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      )}
    </div>
  )
}
