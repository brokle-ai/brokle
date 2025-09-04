'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { AlertTriangle, Trash2, Shield, Download, Users, FolderOpen, CreditCard } from 'lucide-react'
import { useOrganization } from '@/context/organization-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import type { OrganizationParams } from '@/types/organization'

export default function OrganizationDangerPage() {
  const params = useParams() as OrganizationParams
  const router = useRouter()
  const { 
    currentOrganization,
    projects,
    isLoading,
    hasAccess,
    getUserRole
  } = useOrganization()
  
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [acknowledgedRisks, setAcknowledgedRisks] = useState<string[]>([])
  const [isDeleting, setIsDeleting] = useState(false)
  const [isExporting, setIsExporting] = useState(false)

  useEffect(() => {
    if (isLoading) return

    if (!hasAccess(params.orgSlug)) {
      router.push('/')
      return
    }

    // Only organization owners can access danger zone
    const userRole = getUserRole(params.orgSlug)
    if (userRole !== 'owner') {
      router.push(`/${params.orgSlug}/settings`)
      return
    }
  }, [params.orgSlug, isLoading, hasAccess, getUserRole, router])

  if (isLoading) {
    return (
      <>
        <Header>
          <Skeleton className="h-8 w-64" />
        </Header>
        <Main className="space-y-6">
          <Skeleton className="h-6 w-96" />
          <div className="space-y-4">
            <Skeleton className="h-32" />
            <Skeleton className="h-48" />
          </div>
        </Main>
      </>
    )
  }

  if (!currentOrganization) {
    return (
      <>
        <Header>
          <h1 className="text-2xl font-bold text-foreground">Access Denied</h1>
        </Header>
        <Main>
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Access Denied</h2>
            <p className="text-muted-foreground mb-4">
              Only organization owners can access the danger zone.
            </p>
            <button 
              onClick={() => router.push(`/${params.orgSlug}`)}
              className="text-primary hover:underline"
            >
              Go back to organization
            </button>
          </div>
        </Main>
      </>
    )
  }

  const organizationProjects = projects.filter(p => p.organizationId === currentOrganization.id)
  const activeProjects = organizationProjects.filter(p => p.status === 'active')
  const totalMembers = currentOrganization.members.length

  const handleDeleteOrganization = async () => {
    if (deleteConfirmation !== currentOrganization.name) {
      toast.error('Organization name does not match')
      return
    }

    if (acknowledgedRisks.length < 6) {
      toast.error('Please acknowledge all risks before proceeding')
      return
    }

    setIsDeleting(true)
    
    try {
      // TODO: Implement API call to delete organization
      await new Promise(resolve => setTimeout(resolve, 3000))
      
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

  const handleExportData = async () => {
    setIsExporting(true)
    
    try {
      // TODO: Implement API call to export organization data
      await new Promise(resolve => setTimeout(resolve, 3000))
      
      // Simulate comprehensive export
      const exportData = {
        organization: currentOrganization,
        projects: organizationProjects,
        exported_at: new Date().toISOString(),
        data: {
          members: 'Member data would be here...',
          billing: 'Billing history would be here...',
          analytics: 'Analytics data would be here...',
          settings: 'Organization settings would be here...'
        }
      }
      
      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${currentOrganization.slug}-complete-export-${new Date().toISOString().split('T')[0]}.json`
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
    { id: 'all-projects', text: `All ${organizationProjects.length} projects and their data will be permanently deleted` },
    { id: 'member-access', text: `All ${totalMembers} members will lose access immediately` },
    { id: 'billing-data', text: 'Billing history, invoices, and subscription data will be removed' },
    { id: 'api-keys', text: 'All API keys across all projects will be immediately revoked' },
    { id: 'analytics', text: 'Analytics, metrics, and usage history will be permanently deleted' },
    { id: 'irreversible', text: 'This action cannot be undone under any circumstances' }
  ]

  return (
    <>
      <Header>
        <div className="space-y-2">
          <Breadcrumbs />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              Danger Zone
            </h1>
            <p className="text-muted-foreground">
              Irreversible and destructive actions for {currentOrganization.name}
            </p>
          </div>
        </div>
      </Header>

      <Main className="space-y-8">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            <strong>Warning:</strong> These actions will affect your entire organization and all projects. 
            Only organization owners can perform these operations.
          </AlertDescription>
        </Alert>

        {/* Organization Overview */}
        <Card>
          <CardHeader>
            <CardTitle>Organization Impact Assessment</CardTitle>
            <CardDescription>
              Review what will be affected before proceeding with destructive actions
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-4 md:grid-cols-3">
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Projects</p>
                      <p className="text-2xl font-bold">{organizationProjects.length}</p>
                      <p className="text-xs text-muted-foreground">{activeProjects.length} active</p>
                    </div>
                    <FolderOpen className="h-8 w-8 text-blue-500" />
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Members</p>
                      <p className="text-2xl font-bold">{totalMembers}</p>
                      <p className="text-xs text-muted-foreground">across all roles</p>
                    </div>
                    <Users className="h-8 w-8 text-green-500" />
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Plan</p>
                      <p className="text-2xl font-bold capitalize">{currentOrganization.plan}</p>
                      <p className="text-xs text-muted-foreground">subscription active</p>
                    </div>
                    <CreditCard className="h-8 w-8 text-purple-500" />
                  </div>
                </CardContent>
              </Card>
            </div>

            {activeProjects.length > 0 && (
              <div>
                <h4 className="font-medium mb-3">Active Projects</h4>
                <div className="grid gap-2 md:grid-cols-2">
                  {activeProjects.map((project) => (
                    <div key={project.id} className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <div className="font-medium text-sm">{project.name}</div>
                        <div className="text-xs text-muted-foreground">{project.environment}</div>
                      </div>
                      <Badge className="bg-green-100 text-green-800">
                        {project.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Export Data */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5 text-blue-500" />
              Export Organization Data
            </CardTitle>
            <CardDescription>
              Download a complete backup of all organization and project data
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="text-sm text-muted-foreground">
              Complete export includes:
              <ul className="list-disc list-inside mt-2 space-y-1">
                <li>Organization settings and configuration</li>
                <li>All project data, settings, and configurations</li>
                <li>Member information and roles</li>
                <li>Analytics and usage data across all projects</li>
                <li>Billing history and subscription details</li>
                <li>API key metadata (keys themselves are not exported)</li>
              </ul>
            </div>
            
            <Button onClick={handleExportData} disabled={isExporting} className="w-full">
              {isExporting ? (
                <>
                  <Download className="mr-2 h-4 w-4 animate-bounce" />
                  Exporting All Data...
                </>
              ) : (
                <>
                  <Download className="mr-2 h-4 w-4" />
                  Export Complete Organization Data
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        <Separator />

        {/* Delete Organization */}
        <Card className="border-red-200">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-600">
              <Trash2 className="h-5 w-5" />
              Delete Organization
            </CardTitle>
            <CardDescription className="text-red-600">
              Permanently delete this organization and everything it contains. This action cannot be undone.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="text-sm text-red-600">
              <strong>This will permanently delete:</strong>
              <ul className="list-disc list-inside mt-2 space-y-1">
                <li>{organizationProjects.length} projects and all their data</li>
                <li>Access for {totalMembers} organization members</li>
                <li>All analytics, metrics, and usage history</li>
                <li>All API keys across all projects</li>
                <li>Billing history and subscription data</li>
                <li>Organization settings and configurations</li>
              </ul>
            </div>

            <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
              <DialogTrigger asChild>
                <Button variant="destructive" className="w-full">
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete Organization Permanently
                </Button>
              </DialogTrigger>
              
              <DialogContent className="sm:max-w-[600px]">
                <DialogHeader>
                  <DialogTitle className="text-red-600">Delete Organization</DialogTitle>
                  <DialogDescription>
                    This action will permanently delete "{currentOrganization.name}", all {organizationProjects.length} projects,
                    and remove access for {totalMembers} members. This cannot be undone.
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
                    <Label>Acknowledge the following consequences:</Label>
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
                      <strong>Final Warning:</strong> This will delete everything associated with this organization 
                      including all projects, member access, and billing data. Consider exporting your data first.
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
                    disabled={isDeleting || deleteConfirmation !== currentOrganization.name || acknowledgedRisks.length < 6}
                  >
                    {isDeleting ? 'Deleting Organization...' : 'Delete Forever'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </CardContent>
        </Card>
      </Main>
    </>
  )
}