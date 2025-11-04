'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Loader2 } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { toast } from 'sonner'

interface CreateOrgFormProps {
  onSuccess: (orgId: string) => void
}

export function CreateOrgForm({ onSuccess }: CreateOrgFormProps) {
  const { createOrganization } = useWorkspace()
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    billing_email: '',
    plan: 'free' as const,
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Organization name is required'
    }

    if (!formData.billing_email.trim()) {
      newErrors.billing_email = 'Billing email is required'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.billing_email)) {
      newErrors.billing_email = 'Please enter a valid email address'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    setIsLoading(true)

    try {
      const newOrg = await createOrganization(formData)
      toast.success(`Organization "${newOrg.name}" created!`)
      onSuccess(newOrg.id)
    } catch (error) {
      console.error('Failed to create organization:', error)
      toast.error('Failed to create organization')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="name">Organization Name *</Label>
          <Input
            id="name"
            placeholder="Acme Inc"
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
          <Label htmlFor="billing_email">Billing Email *</Label>
          <Input
            id="billing_email"
            type="email"
            placeholder="billing@acme.com"
            value={formData.billing_email}
            onChange={(e) => {
              setFormData(prev => ({ ...prev, billing_email: e.target.value }))
              if (errors.billing_email) setErrors(prev => ({ ...prev, billing_email: '' }))
            }}
            disabled={isLoading}
          />
          {errors.billing_email && (
            <p className="text-sm text-destructive">{errors.billing_email}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="plan">Plan</Label>
          <Select
            value={formData.plan}
            onValueChange={(value: 'free' | 'pro' | 'enterprise') =>
              setFormData(prev => ({ ...prev, plan: value }))
            }
            disabled={isLoading}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="free">Free</SelectItem>
              <SelectItem value="pro">Pro</SelectItem>
              <SelectItem value="enterprise">Enterprise</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Creating...
          </>
        ) : (
          'Create Organization'
        )}
      </Button>
    </form>
  )
}
