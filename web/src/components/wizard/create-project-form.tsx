'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Loader2 } from 'lucide-react'
import { createProject } from '@/lib/api/services/organizations'
import { toast } from 'sonner'

interface CreateProjectFormProps {
  organizationId: string
  onSuccess: (projectId: string) => void
}

export function CreateProjectForm({ organizationId, onSuccess }: CreateProjectFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

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

    if (!validateForm()) return

    setIsLoading(true)

    try {
      const newProject = await createProject(organizationId, formData)

      toast.success(`Project "${newProject.name}" created!`)
      onSuccess(newProject.id)
    } catch (error) {
      console.error('Failed to create project:', error)
      toast.error('Failed to create project')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="project-name">Project Name *</Label>
          <Input
            id="project-name"
            placeholder="My AI Application"
            value={formData.name}
            onChange={(e) => {
              setFormData(prev => ({ ...prev, name: e.target.value }))
              if (errors.name) setErrors(prev => ({ ...prev, name: '' }))
            }}
            disabled={isLoading}
          />
          {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Textarea
            id="description"
            placeholder="Describe your project..."
            value={formData.description}
            onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
            disabled={isLoading}
            rows={3}
          />
        </div>
      </div>

      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Creating...
          </>
        ) : (
          'Create Project'
        )}
      </Button>
    </form>
  )
}
