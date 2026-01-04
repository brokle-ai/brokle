'use client'

import { useState } from 'react'
import { Plus, FileText, LayoutTemplate } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DashboardForm } from './dashboard-form'
import { TemplateSelector } from './template-selector'
import { useCreateDashboardMutation } from '../hooks/use-dashboards-queries'
import { useCreateFromTemplateMutation } from '../hooks/use-templates'
import type { CreateDashboardRequest, DashboardTemplate } from '../types'

interface CreateDashboardDialogProps {
  projectId: string
}

export function CreateDashboardDialog({ projectId }: CreateDashboardDialogProps) {
  const [open, setOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<'blank' | 'template'>('blank')

  const createMutation = useCreateDashboardMutation(projectId)
  const createFromTemplateMutation = useCreateFromTemplateMutation(projectId)

  const handleBlankSubmit = async (data: CreateDashboardRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  const handleTemplateSelect = async (template: DashboardTemplate, dashboardName: string) => {
    await createFromTemplateMutation.mutateAsync({
      template_id: template.id,
      name: dashboardName,
    })
    setOpen(false)
  }

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    // Reset to blank tab when dialog closes
    if (!isOpen) {
      setActiveTab('blank')
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Dashboard
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Dashboard</DialogTitle>
          <DialogDescription>
            Create a new dashboard to visualize your observability data.
          </DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'blank' | 'template')}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="blank" className="gap-2">
              <FileText className="h-4 w-4" />
              Blank Dashboard
            </TabsTrigger>
            <TabsTrigger value="template" className="gap-2">
              <LayoutTemplate className="h-4 w-4" />
              From Template
            </TabsTrigger>
          </TabsList>

          <TabsContent value="blank" className="mt-4">
            <DashboardForm
              onSubmit={handleBlankSubmit}
              onCancel={() => setOpen(false)}
              isLoading={createMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="template" className="mt-4">
            <TemplateSelector
              onSelect={handleTemplateSelect}
              onCancel={() => setOpen(false)}
              isLoading={createFromTemplateMutation.isPending}
            />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}
