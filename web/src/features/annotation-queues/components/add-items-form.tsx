'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { AddItemsBatchRequest, ObjectType } from '../types'

const addItemsSchema = z.object({
  objectIds: z.string().min(1, 'At least one ID is required'),
  objectType: z.enum(['trace', 'span']),
  priority: z.number().min(0).max(100).optional(),
})

type AddItemsFormData = z.infer<typeof addItemsSchema>

interface AddItemsFormProps {
  onSubmit: (data: AddItemsBatchRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function AddItemsForm({
  onSubmit,
  onCancel,
  isLoading,
}: AddItemsFormProps) {
  const [inputMode, setInputMode] = useState<'single' | 'bulk'>('single')

  const form = useForm<AddItemsFormData>({
    resolver: zodResolver(addItemsSchema),
    defaultValues: {
      objectIds: '',
      objectType: 'trace',
      priority: 0,
    },
  })

  const handleSubmit = (data: AddItemsFormData) => {
    // Parse IDs from text (comma, newline, or space separated)
    const ids = data.objectIds
      .split(/[\s,\n]+/)
      .map((id) => id.trim())
      .filter((id) => id.length > 0)

    if (ids.length === 0) {
      form.setError('objectIds', { message: 'At least one valid ID is required' })
      return
    }

    const items = ids.map((id) => ({
      object_id: id,
      object_type: data.objectType as ObjectType,
      priority: data.priority,
    }))

    onSubmit({ items })
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="objectType"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Item Type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="trace">Traces</SelectItem>
                  <SelectItem value="span">Spans</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                Choose whether to add trace IDs or span IDs.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Tabs value={inputMode} onValueChange={(v) => setInputMode(v as 'single' | 'bulk')}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="single">Single ID</TabsTrigger>
            <TabsTrigger value="bulk">Bulk Add</TabsTrigger>
          </TabsList>

          <TabsContent value="single" className="mt-4">
            <FormField
              control={form.control}
              name="objectIds"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {form.watch('objectType') === 'trace' ? 'Trace ID' : 'Span ID'}
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder={
                        form.watch('objectType') === 'trace'
                          ? 'e.g., abc123def456...'
                          : 'e.g., span789xyz...'
                      }
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </TabsContent>

          <TabsContent value="bulk" className="mt-4">
            <FormField
              control={form.control}
              name="objectIds"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {form.watch('objectType') === 'trace' ? 'Trace IDs' : 'Span IDs'}
                  </FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder={`Enter ${form.watch('objectType').toLowerCase()} IDs, one per line or comma-separated`}
                      rows={6}
                      className="font-mono text-sm"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    You can paste multiple IDs separated by commas, spaces, or newlines.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </TabsContent>
        </Tabs>

        <FormField
          control={form.control}
          name="priority"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Priority (Optional)</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  min={0}
                  max={100}
                  placeholder="0"
                  {...field}
                  onChange={(e) => field.onChange(parseInt(e.target.value, 10) || 0)}
                />
              </FormControl>
              <FormDescription>
                Higher priority items will be shown first (0-100).
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Adding...' : 'Add Items'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
