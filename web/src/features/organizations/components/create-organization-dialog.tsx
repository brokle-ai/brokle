'use client'

import { useRouter } from 'next/navigation'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useCreateOrganizationMutation } from '../hooks/use-organization-queries'
import { buildOrgUrl } from '@/lib/utils/slug-utils'
import { toast } from 'sonner'

// Zod validation schema - name only
const createOrgSchema = z.object({
  name: z
    .string()
    .min(2, 'Organization name must be at least 2 characters')
    .max(100, 'Organization name must be less than 100 characters'),
})

type CreateOrgFormData = z.infer<typeof createOrgSchema>

interface CreateOrganizationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateOrganizationDialog({
  open,
  onOpenChange,
}: CreateOrganizationDialogProps) {
  const router = useRouter()
  const createOrgMutation = useCreateOrganizationMutation()

  const form = useForm<CreateOrgFormData>({
    resolver: zodResolver(createOrgSchema),
    defaultValues: {
      name: '',
    },
  })

  const onSubmit = async (data: CreateOrgFormData) => {
    try {
      const newOrg = await createOrgMutation.mutateAsync({
        name: data.name,
        // description is reserved for future backend use
      })

      // Navigate to new organization dashboard
      const orgUrl = buildOrgUrl(newOrg.name, newOrg.id)

      try {
        router.push(orgUrl)
      } catch (navError) {
        // Fallback: close dialog if navigation fails
        if (process.env.NODE_ENV === 'development') {
          console.error('Navigation failed:', navError)
        }
        onOpenChange(false)
        toast.error('Navigation failed. Please use the organization selector.')
      }
    } catch (error) {
      // Error handled by mutation hook (toast notification)
      if (process.env.NODE_ENV === 'development') {
        console.error('Organization creation failed:', error)
      }
    }
  }

  const handleOpenChange = (isOpen: boolean) => {
    // Prevent closing dialog during submission
    if (!isOpen && createOrgMutation.isPending) {
      return
    }

    onOpenChange(isOpen)
    if (!isOpen) {
      // Reset form when dialog closes
      form.reset()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create Organization</DialogTitle>
          <DialogDescription>
            Set up your organization to start managing projects and team
            members.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Organization Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Acme Corp"
                      {...field}
                      disabled={createOrgMutation.isPending}
                    />
                  </FormControl>
                  <FormDescription>
                    This will be visible to your team members
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={createOrgMutation.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createOrgMutation.isPending}>
                {createOrgMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Organization'
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
