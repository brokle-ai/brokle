'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Building2, Loader2 } from 'lucide-react'
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

interface CreateOrganizationModalProps {
  trigger?: React.ReactNode
  onSuccess?: (orgSlug: string) => void
}

export function CreateOrganizationModal({ trigger, onSuccess }: CreateOrganizationModalProps) {
  const router = useRouter()
  const { createOrganization, organizations } = useOrganization()
  
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    billing_email: '',
    plan: 'free' as const,
  })
  const [slugTouched, setSlugTouched] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const existingSlugs = organizations.map(org => org.slug)

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
      newErrors.name = 'Organization name is required'
    }

    if (!formData.slug.trim()) {
      newErrors.slug = 'Organization slug is required'
    } else if (!isValidSlug(formData.slug)) {
      newErrors.slug = 'Slug can only contain lowercase letters, numbers, and hyphens'
    } else if (existingSlugs.includes(formData.slug)) {
      newErrors.slug = 'This slug is already taken'
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

    if (!validateForm()) {
      return
    }

    setIsLoading(true)

    try {
      const newOrg = await createOrganization(formData)
      
      toast.success(`Organization "${newOrg.name}" created successfully!`)
      
      // Reset form
      setFormData({
        name: '',
        slug: '',
        billing_email: '',
        plan: 'free',
      })
      setSlugTouched(false)
      setErrors({})
      setIsOpen(false)

      // Navigate to new organization or call callback
      if (onSuccess) {
        onSuccess(newOrg.slug)
      } else {
        router.push(`/${newOrg.slug}`)
      }
    } catch (error) {
      console.error('Failed to create organization:', error)
      toast.error(
        error instanceof Error ? error.message : 'Failed to create organization'
      )
    } finally {
      setIsLoading(false)
    }
  }

  const defaultTrigger = (
    <Button>
      <Plus className="mr-2 h-4 w-4" />
      Create Organization
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
            <Building2 className="h-5 w-5" />
            Create Organization
          </DialogTitle>
          <DialogDescription>
            Set up a new organization to manage your AI projects and team members.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Organization Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Acme Corp"
              className={errors.name ? 'border-destructive' : ''}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="slug">URL Slug *</Label>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">brokle.com/</span>
              <Input
                id="slug"
                value={formData.slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                placeholder="acme-corp"
                className={errors.slug ? 'border-destructive' : ''}
              />
            </div>
            {errors.slug && (
              <p className="text-sm text-destructive">{errors.slug}</p>
            )}
            <p className="text-xs text-muted-foreground">
              This will be used in your organization's URL. Only lowercase letters, numbers, and hyphens are allowed.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="billing_email">Billing Email *</Label>
            <Input
              id="billing_email"
              type="email"
              value={formData.billing_email}
              onChange={(e) => setFormData(prev => ({ ...prev, billing_email: e.target.value }))}
              placeholder="billing@acme.com"
              className={errors.billing_email ? 'border-destructive' : ''}
            />
            {errors.billing_email && (
              <p className="text-sm text-destructive">{errors.billing_email}</p>
            )}
            <p className="text-xs text-muted-foreground">
              This email will receive billing notifications and invoices.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="plan">Initial Plan</Label>
            <Select
              value={formData.plan}
              onValueChange={(value: 'free' | 'pro' | 'business' | 'enterprise') =>
                setFormData(prev => ({ ...prev, plan: value }))
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a plan" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="free">Free - 10K requests/month</SelectItem>
                <SelectItem value="pro">Pro - 100K requests/month ($29)</SelectItem>
                <SelectItem value="business">Business - 1M requests/month ($99)</SelectItem>
                <SelectItem value="enterprise">Enterprise - Unlimited (Custom)</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              You can upgrade or downgrade your plan anytime from the settings.
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
                Create Organization
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}