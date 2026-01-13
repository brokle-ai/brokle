'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { AnnotationQueue, CreateQueueRequest, QueueStatus } from '../types'

const queueSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().optional(),
  instructions: z.string().optional(),
  status: z.enum(['active', 'paused', 'archived']).optional(),
  lock_timeout_seconds: z.number().min(60).max(3600).optional(),
  auto_assignment: z.boolean().optional(),
})

type QueueFormData = z.infer<typeof queueSchema>

interface QueueFormProps {
  queue?: AnnotationQueue
  onSubmit: (data: CreateQueueRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function QueueForm({
  queue,
  onSubmit,
  onCancel,
  isLoading,
}: QueueFormProps) {
  const form = useForm<QueueFormData>({
    resolver: zodResolver(queueSchema),
    defaultValues: {
      name: queue?.name ?? '',
      description: queue?.description ?? '',
      instructions: queue?.instructions ?? '',
      status: queue?.status ?? 'active',
      lock_timeout_seconds: queue?.settings?.lock_timeout_seconds ?? 300,
      auto_assignment: queue?.settings?.auto_assignment ?? false,
    },
  })

  const handleSubmit = (data: QueueFormData) => {
    onSubmit({
      name: data.name,
      description: data.description || undefined,
      instructions: data.instructions || undefined,
      settings: {
        lock_timeout_seconds: data.lock_timeout_seconds,
        auto_assignment: data.auto_assignment,
      },
    })
  }

  const isEditing = !!queue

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
                <Input placeholder="e.g., Quality Review Queue" {...field} />
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
                  placeholder="Describe the purpose of this annotation queue"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="instructions"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Instructions</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Provide guidelines for annotators (e.g., scoring criteria, what to look for)"
                  rows={4}
                  {...field}
                />
              </FormControl>
              <FormDescription>
                These instructions will be shown to annotators when reviewing items.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {isEditing && (
          <FormField
            control={form.control}
            name="status"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Status</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="paused">Paused</SelectItem>
                    <SelectItem value="archived">Archived</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        )}

        <div className="space-y-4 rounded-lg border p-4">
          <h4 className="font-medium text-sm">Queue Settings</h4>

          <FormField
            control={form.control}
            name="lock_timeout_seconds"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Lock Timeout</FormLabel>
                <Select
                  onValueChange={(value) => field.onChange(parseInt(value, 10))}
                  defaultValue={String(field.value)}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select timeout" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="60">1 minute</SelectItem>
                    <SelectItem value="180">3 minutes</SelectItem>
                    <SelectItem value="300">5 minutes (default)</SelectItem>
                    <SelectItem value="600">10 minutes</SelectItem>
                    <SelectItem value="1800">30 minutes</SelectItem>
                    <SelectItem value="3600">1 hour</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  How long an item stays locked when claimed by an annotator.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="auto_assignment"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div className="space-y-0.5">
                  <FormLabel>Auto Assignment</FormLabel>
                  <FormDescription>
                    Automatically assign items to available annotators.
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
        </div>

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Saving...' : isEditing ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
