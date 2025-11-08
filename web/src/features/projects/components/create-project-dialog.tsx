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
import { useCreateProjectMutation } from '../hooks/use-project-queries'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { toast } from 'sonner'

// Zod validation schema - name only
const createProjectSchema = z.object({
  name: z
    .string()
    .min(2, 'Project name must be at least 2 characters')
    .max(100, 'Project name must be less than 100 characters'),
})

type CreateProjectFormData = z.infer<typeof createProjectSchema>

interface CreateProjectDialogProps {
  organizationId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateProjectDialog({
  organizationId,
  open,
  onOpenChange,
}: CreateProjectDialogProps) {
  const router = useRouter()
  const createProjectMutation = useCreateProjectMutation()

  const form = useForm<CreateProjectFormData>({
    resolver: zodResolver(createProjectSchema),
    defaultValues: {
      name: '',
    },
  })

  const onSubmit = async (data: CreateProjectFormData) => {
    try {
      const newProject = await createProjectMutation.mutateAsync({
        organizationId,
        name: data.name,
      })

      // Close dialog before navigation
      onOpenChange(false)

      // Navigate to new project dashboard
      const projectUrl = buildProjectUrl(newProject.name, newProject.id)

      try {
        router.push(projectUrl)
      } catch (navError) {
        // Fallback: show error if navigation fails
        if (process.env.NODE_ENV === 'development') {
          console.error('Navigation failed:', navError)
        }
        toast.error('Navigation failed. Please use the project selector.')
      }
    } catch (error) {
      // Error handled by mutation hook (toast notification)
      if (process.env.NODE_ENV === 'development') {
        console.error('Project creation failed:', error)
      }
    }
  }

  const handleOpenChange = (isOpen: boolean) => {
    // Prevent closing dialog during submission
    if (!isOpen && createProjectMutation.isPending) {
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
          <DialogTitle>Create Project</DialogTitle>
          <DialogDescription>
            Create a new project to organize your AI applications and APIs.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Project Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="My AI Project"
                      {...field}
                      disabled={createProjectMutation.isPending}
                    />
                  </FormControl>
                  <FormDescription>
                    A descriptive name for your project
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
                disabled={createProjectMutation.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createProjectMutation.isPending}>
                {createProjectMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Project'
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
