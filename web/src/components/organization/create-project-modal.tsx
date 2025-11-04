'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, FolderOpen, Loader2 } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { getOrgSlug, getProjectSlug } from '@/lib/utils/slug-utils'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'sonner'

interface CreateProjectModalProps {
  trigger?: React.ReactNode
  onSuccess?: (projectSlug: string) => void
}

export function CreateProjectModal({ trigger, onSuccess }: CreateProjectModalProps) {
  const router = useRouter()
  const { createProject, currentOrganization, projects } = useWorkspace()
  
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    environment: 'development' as const,
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  if (!currentOrganization) {
    return null // Should not render if no organization is selected
  }

  const handleNameChange = (name: string) => {
    setFormData(prev => ({ ...prev, name }))
    
    // Clear name error
    if (errors.name) {
      setErrors(prev => ({ ...prev, name: '' }))
    }
  }

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Project name is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    setIsLoading(true)

    try {
      const newProject = await createProject({
        ...formData,
        organizationId: currentOrganization.id,
      })
      
      toast.success(`Project "${newProject.name}" created successfully!`)
      
      // Reset form
      setFormData({
        name: '',
        description: '',
        environment: 'development',
      })
      setErrors({})
      setIsOpen(false)

      // Navigate to new project or call callback
      if (onSuccess) {
        onSuccess(getProjectSlug(newProject))
      } else {
        router.push(`/organizations/${getOrgSlug(currentOrganization)}/projects/${getProjectSlug(newProject)}`)
      }
    } catch (error) {
      console.error('Failed to create project:', error)
      toast.error(
        error instanceof Error ? error.message : 'Failed to create project'
      )
    } finally {
      setIsLoading(false)
    }
  }

  const defaultTrigger = (
    <Button>
      <Plus className="mr-2 h-4 w-4" />
      New Project
    </Button>
  )

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {trigger || defaultTrigger}
      </DialogTrigger>
      
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5" />
            Create Project
          </DialogTitle>
          <DialogDescription>
            Create a new AI project in {currentOrganization.name}. Each project can have its own settings, API keys, and analytics.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Project Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Customer Chatbot"
              className={errors.name ? 'border-destructive' : ''}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name}</p>
            )}
          </div>


          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="AI-powered customer support chatbot for handling common inquiries..."
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              Optional description to help identify this project.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="environment">Environment</Label>
            <Select
              value={formData.environment}
              onValueChange={(value: 'development' | 'staging' | 'production') =>
                setFormData(prev => ({ ...prev, environment: value }))
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select environment" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="development">Development</SelectItem>
                <SelectItem value="staging">Staging</SelectItem>
                <SelectItem value="production">Production</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Choose the environment type for this project. This helps organize your projects and can affect default settings.
            </p>
          </div>
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => setIsOpen(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Plus className="mr-2 h-4 w-4" />
                Create Project
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}