'use client'

import { useRouter } from 'next/navigation'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
import { Textarea } from '@/components/ui/textarea'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card'
import { useCreateOrganizationMutation } from '@/features/organizations'
import { buildOrgUrl } from '@/lib/utils/slug-utils'

// Zod validation schema
const createOrgSchema = z.object({
  name: z
    .string()
    .min(2, 'Organization name must be at least 2 characters')
    .max(100, 'Organization name must be less than 100 characters'),
  description: z
    .string()
    .max(500, 'Description must be less than 500 characters')
    .optional(),
})

export default function CreateOrganizationPage() {
  const router = useRouter()
  const createOrgMutation = useCreateOrganizationMutation()

  const form = useForm<z.infer<typeof createOrgSchema>>({
    resolver: zodResolver(createOrgSchema),
    defaultValues: {
      name: '',
      description: '',
    },
  })

  const onSubmit = async (data: z.infer<typeof createOrgSchema>) => {
    try {
      const newOrg = await createOrgMutation.mutateAsync({
        name: data.name,
        description: data.description,
      })

      // Redirect to new organization dashboard
      const orgUrl = buildOrgUrl(newOrg.name, newOrg.id)
      router.push(orgUrl)
    } catch (error) {
      // Error handled by mutation hook (toast notification)
      console.error('Organization creation failed:', error)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Create Your Organization</CardTitle>
          <CardDescription>
            Set up your organization to start managing projects and team members.
          </CardDescription>
        </CardHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Organization Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Acme Corp" {...field} />
                    </FormControl>
                    <FormDescription>
                      This will be visible to your team members
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description (Optional)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Tell us about your organization..."
                        className="resize-none"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      A brief description of your organization
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </CardContent>
            <CardFooter>
              <Button
                type="submit"
                disabled={createOrgMutation.isPending}
                className="w-full"
              >
                {createOrgMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating Organization...
                  </>
                ) : (
                  'Create Organization'
                )}
              </Button>
            </CardFooter>
          </form>
        </Form>
      </Card>
    </div>
  )
}
