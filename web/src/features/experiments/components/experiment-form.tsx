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
import type { Experiment, CreateExperimentRequest } from '../types'

const experimentSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().optional(),
})

type ExperimentFormData = z.infer<typeof experimentSchema>

interface ExperimentFormProps {
  experiment?: Experiment
  onSubmit: (data: CreateExperimentRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function ExperimentForm({
  experiment,
  onSubmit,
  onCancel,
  isLoading,
}: ExperimentFormProps) {
  const form = useForm<ExperimentFormData>({
    resolver: zodResolver(experimentSchema),
    defaultValues: {
      name: experiment?.name ?? '',
      description: experiment?.description ?? '',
    },
  })

  const handleSubmit = (data: ExperimentFormData) => {
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
                <Input placeholder="e.g., GPT-4 vs Claude Comparison" {...field} />
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
                  placeholder="Describe the purpose of this experiment"
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
            {isLoading ? 'Saving...' : experiment ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
