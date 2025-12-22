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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import type { Dataset, CreateDatasetRequest } from '../types'

const datasetSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().optional(),
})

type DatasetFormData = z.infer<typeof datasetSchema>

interface DatasetFormProps {
  dataset?: Dataset
  onSubmit: (data: CreateDatasetRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function DatasetForm({
  dataset,
  onSubmit,
  onCancel,
  isLoading,
}: DatasetFormProps) {
  const form = useForm<DatasetFormData>({
    resolver: zodResolver(datasetSchema),
    defaultValues: {
      name: dataset?.name ?? '',
      description: dataset?.description ?? '',
    },
  })

  const handleSubmit = (data: DatasetFormData) => {
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
                <Input placeholder="e.g., Customer Support Test Cases" {...field} />
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
                  placeholder="Describe the purpose of this dataset"
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
            {isLoading ? 'Saving...' : dataset ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
