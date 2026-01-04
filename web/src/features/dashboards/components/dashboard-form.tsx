'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import type { Dashboard, CreateDashboardRequest } from '../types'

const dashboardSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().optional(),
})

type DashboardFormData = z.infer<typeof dashboardSchema>

interface DashboardFormProps {
  dashboard?: Dashboard
  onSubmit: (data: CreateDashboardRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function DashboardForm({
  dashboard,
  onSubmit,
  onCancel,
  isLoading,
}: DashboardFormProps) {
  const form = useForm<DashboardFormData>({
    resolver: zodResolver(dashboardSchema),
    defaultValues: {
      name: dashboard?.name ?? '',
      description: dashboard?.description ?? '',
    },
  })

  const handleSubmit = (data: DashboardFormData) => {
    onSubmit({
      name: data.name,
      description: data.description || undefined,
    })
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="e.g., Main Dashboard" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Description</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Describe the purpose of this dashboard"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Saving...' : dashboard ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
