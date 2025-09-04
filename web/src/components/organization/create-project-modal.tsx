'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, FolderOpen, Loader2 } from 'lucide-react'
import { useOrganization } from '@/context/organization-context'
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
import { generateSlug, isValidSlug } from '@/lib/utils/slug-utils'

interface CreateProjectModalProps {
  trigger?: React.ReactNode
  onSuccess?: (projectSlug: string) => void
}

export function CreateProjectModal({ trigger, onSuccess }: CreateProjectModalProps) {
  const router = useRouter()
  const { createProject, currentOrganization, projects } = useOrganization()
  
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    description: '',
    environment: 'development' as const,
  })
  const [slugTouched, setSlugTouched] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  if (!currentOrganization) {
    return null // Should not render if no organization is selected
  }

  const existingSlugs = projects.map(project => project.slug)

  const handleNameChange = (name: string) => {
    setFormData(prev => ({
      ...prev,
      name,
      // Auto-generate slug if user hasn't manually edited it
      ...(slugTouched ? {} : { slug: generateSlug(name) })
    }))
    
    // Clear name error
    if (errors.name) {
      setErrors(prev => ({ ...prev, name: '' }))
    }
  }

  const handleSlugChange = (slug: string) => {
    setSlugTouched(true)
    setFormData(prev => ({ ...prev, slug }))
    
    // Clear slug error
    if (errors.slug) {
      setErrors(prev => ({ ...prev, slug: '' }))
    }
  }

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Project name is required'
    }

    if (!formData.slug.trim()) {
      newErrors.slug = 'Project slug is required'
    } else if (!isValidSlug(formData.slug)) {
      newErrors.slug = 'Slug can only contain lowercase letters, numbers, and hyphens'
    } else if (existingSlugs.includes(formData.slug)) {
      newErrors.slug = 'This slug is already taken in this organization'
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
        slug: '',
        description: '',
        environment: 'development',
      })
      setSlugTouched(false)
      setErrors({})
      setIsOpen(false)

      // Navigate to new project or call callback
      if (onSuccess) {
        onSuccess(newProject.slug)
      } else {
        router.push(`/${currentOrganization.slug}/${newProject.slug}`)
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
            <Label htmlFor="slug">URL Slug *</Label>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">.../{currentOrganization.slug}/</span>
              <Input
                id="slug"
                value={formData.slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                placeholder="customer-chatbot"
                className={errors.slug ? 'border-destructive' : ''}
              />
            </div>
            {errors.slug && (
              <p className="text-sm text-destructive">{errors.slug}</p>
            )}
            <p className="text-xs text-muted-foreground">
              This will be used in your project's URL. Only lowercase letters, numbers, and hyphens are allowed.
            </p>
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