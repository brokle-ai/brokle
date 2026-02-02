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
      await createOrgMutation.mutateAsync({
        name: data.name,
        // description is reserved for future backend use
      })

      // Close dialog and navigate to root
      // Root page will show "Create Your First Project" for the new org
      // (PostHog pattern: new orgs have no projects, root handles empty state)
      // Note: Success toast is shown by the mutation hook
      onOpenChange(false)
      router.push('/')
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
