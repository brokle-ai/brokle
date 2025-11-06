'use client'

import { useState } from 'react'
import { 
  CheckSquare, 
  Square, 
  Trash2, 
  Archive, 
  Play, 
  Pause, 
  Copy, 
  Download,
  Settings,
  X
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { toast } from 'sonner'
import type { Project, ProjectStatus, ProjectEnvironment } from '../types'

interface BulkActionsBarProps {
  selectedProjects: Project[]
  onClearSelection: () => void
  onProjectsUpdated: () => void
}

export function BulkActionsBar({ 
  selectedProjects, 
  onClearSelection,
  onProjectsUpdated 
}: BulkActionsBarProps) {
  const [isActionDialogOpen, setIsActionDialogOpen] = useState(false)
  const [currentAction, setCurrentAction] = useState<string>('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [newStatus, setNewStatus] = useState<ProjectStatus>('active')
  const [newEnvironment, setNewEnvironment] = useState<ProjectEnvironment>('production')

  if (selectedProjects.length === 0) {
    return null
  }

  const selectedCount = selectedProjects.length
  const activeCount = selectedProjects.filter(p => p.status === 'active').length
  const inactiveCount = selectedProjects.filter(p => p.status === 'inactive').length
  const archivedCount = selectedProjects.filter(p => p.status === 'archived').length

  const handleBulkAction = async (action: string) => {
    setCurrentAction(action)
    
    if (action === 'delete' || action === 'archive') {
      setIsActionDialogOpen(true)
      return
    }

    await executeAction(action)
  }

  const executeAction = async (action: string) => {
    setIsProcessing(true)
    
    try {
      // TODO: Implement actual API calls
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      let message = ''
      switch (action) {
        case 'activate':
          message = `${selectedCount} projects activated`
          break
        case 'deactivate':
          message = `${selectedCount} projects deactivated`
          break
        case 'archive':
          message = `${selectedCount} projects archived`
          break
        case 'delete':
          message = `${selectedCount} projects deleted`
          break
        case 'duplicate':
          message = `${selectedCount} projects duplicated`
          break
        case 'export':
          // Simulate file download
          const blob = new Blob([JSON.stringify({
            projects: selectedProjects,
            exported_at: new Date().toISOString(),
            count: selectedCount
          }, null, 2)], { type: 'application/json' })
          
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = `bulk-projects-export-${new Date().toISOString().split('T')[0]}.json`
          a.click()
          URL.revokeObjectURL(url)
          
          message = `${selectedCount} projects exported`
          break
        case 'change-status':
          message = `${selectedCount} projects status changed to ${newStatus}`
          break
        case 'change-environment':
          message = `${selectedCount} projects environment changed to ${newEnvironment}`
          break
        default:
          message = `Bulk action completed for ${selectedCount} projects`
      }
      
      toast.success(message)
      onProjectsUpdated()
      onClearSelection()
      setIsActionDialogOpen(false)
      
    } catch (error) {
      console.error('Bulk action failed:', error)
      toast.error('Bulk action failed. Please try again.')
    } finally {
      setIsProcessing(false)
    }
  }

  const getActionContent = () => {
    switch (currentAction) {
      case 'delete':
        return {
          title: 'Delete Projects',
          description: `Are you sure you want to permanently delete ${selectedCount} projects? This action cannot be undone.`,
          destructive: true,
          content: (
            <Alert className="border-red-200">
              <AlertDescription className="text-red-600">
                This will permanently delete all selected projects and their associated data, including:
                <ul className="list-disc list-inside mt-2">
                  <li>All project configurations and settings</li>
                  <li>Analytics and usage data</li>
                  <li>API keys and access tokens</li>
                  <li>Request logs and audit trails</li>
                </ul>
              </AlertDescription>
            </Alert>
          )
        }
      case 'archive':
        return {
          title: 'Archive Projects',
          description: `Archive ${selectedCount} projects? They will be made read-only but data will be preserved.`,
          destructive: false,
          content: (
            <Alert>
              <AlertDescription>
                Archiving will:
                <ul className="list-disc list-inside mt-2">
                  <li>Stop all API requests to these projects</li>
                  <li>Revoke active API keys</li>
                  <li>Preserve all data and settings</li>
                  <li>Allow viewing and data export</li>
                </ul>
              </AlertDescription>
            </Alert>
          )
        }
      default:
        return {
          title: 'Confirm Action',
          description: `Perform bulk action on ${selectedCount} projects?`,
          destructive: false,
          content: null
        }
    }
  }

  const actionContent = getActionContent()

  return (
    <>
      <div className="sticky top-0 z-10 bg-background border-b p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <CheckSquare className="h-4 w-4 text-primary" />
              <span className="font-medium">{selectedCount} selected</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClearSelection}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
            
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {activeCount > 0 && <Badge variant="secondary">{activeCount} active</Badge>}
              {inactiveCount > 0 && <Badge variant="outline">{inactiveCount} inactive</Badge>}
              {archivedCount > 0 && <Badge variant="outline">{archivedCount} archived</Badge>}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Select onValueChange={(value) => {
              setNewStatus(value as ProjectStatus)
              handleBulkAction('change-status')
            }}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="Change Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="active">Set Active</SelectItem>
                <SelectItem value="inactive">Set Inactive</SelectItem>
                <SelectItem value="archived">Set Archived</SelectItem>
              </SelectContent>
            </Select>

            <Select onValueChange={(value) => {
              setNewEnvironment(value as ProjectEnvironment)
              handleBulkAction('change-environment')
            }}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="Change Environment" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="development">Development</SelectItem>
                <SelectItem value="staging">Staging</SelectItem>
                <SelectItem value="production">Production</SelectItem>
              </SelectContent>
            </Select>

            {activeCount > 0 && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleBulkAction('deactivate')}
              >
                <Pause className="mr-2 h-4 w-4" />
                Pause
              </Button>
            )}

            {inactiveCount > 0 && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleBulkAction('activate')}
              >
                <Play className="mr-2 h-4 w-4" />
                Activate
              </Button>
            )}

            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkAction('duplicate')}
            >
              <Copy className="mr-2 h-4 w-4" />
              Duplicate
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkAction('export')}
            >
              <Download className="mr-2 h-4 w-4" />
              Export
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkAction('archive')}
              className="border-yellow-200 text-yellow-700 hover:bg-yellow-50"
            >
              <Archive className="mr-2 h-4 w-4" />
              Archive
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkAction('delete')}
              className="border-red-200 text-red-700 hover:bg-red-50"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          </div>
        </div>
      </div>

      <Dialog open={isActionDialogOpen} onOpenChange={setIsActionDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className={actionContent.destructive ? 'text-red-600' : ''}>
              {actionContent.title}
            </DialogTitle>
            <DialogDescription>
              {actionContent.description}
            </DialogDescription>
          </DialogHeader>

          {actionContent.content && (
            <div className="py-4">
              {actionContent.content}
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsActionDialogOpen(false)}
              disabled={isProcessing}
            >
              Cancel
            </Button>
            <Button
              variant={actionContent.destructive ? "destructive" : "default"}
              onClick={() => executeAction(currentAction)}
              disabled={isProcessing}
            >
              {isProcessing ? (
                <>
                  <Settings className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                `${actionContent.title.split(' ')[0]} ${selectedCount} Projects`
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}