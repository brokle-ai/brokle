import { useState } from 'react'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type {
  AddItemsBatchRequest,
  AddQueueItemRequest,
  ObjectType,
} from '../api/types'

// React-hook-form lets us colocate validation, the resolver hooks zod
// in for us. Splitting `objectIds` out of the API shape keeps the form
// state simple — we parse the bulk-input string at submit time and
// emit an `AddItemsBatchRequest` to the parent.
const schema = z.object({
  objectIds: z.string().min(1, 'At least one ID is required'),
  objectType: z.enum(['trace', 'span']),
  priority: z.number().min(0).max(100).optional(),
  // Toggle is form-only state — never sent on the wire. Surfaces the
  // single/bulk Tabs choice through the form so RHF can drive both
  // tabs from one `objectIds` field.
})

type AddItemsFormValues = z.infer<typeof schema>

interface AddItemsFormProps {
  onSubmit: (data: AddItemsBatchRequest) => void
  onCancel: () => void
  isLoading?: boolean
  /** Optional preselect for the object-type select. */
  defaultObjectType?: ObjectType
}

export function AddItemsForm({
  onSubmit,
  onCancel,
  isLoading,
  defaultObjectType = 'trace',
}: AddItemsFormProps) {
  const [mode, setMode] = useState<'single' | 'bulk'>('single')

  const form = useForm<AddItemsFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      objectIds: '',
      objectType: defaultObjectType,
      priority: 0,
    },
  })

  const handleSubmit = (values: AddItemsFormValues) => {
    const ids = values.objectIds
      .split(/[\s,\n]+/)
      .map((id) => id.trim())
      .filter((id) => id.length > 0)

    if (ids.length === 0) {
      form.setError('objectIds', {
        message: 'At least one valid ID is required',
      })
      return
    }

    const items: AddQueueItemRequest[] = ids.map((id) => ({
      object_id: id,
      object_type: values.objectType,
      priority: values.priority,
    }))
    onSubmit({ items })
  }

  const objectType = form.watch('objectType')

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="objectType"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Item Type</FormLabel>
              <Select onValueChange={field.onChange} value={field.value}>
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

        <Tabs
          value={mode}
          onValueChange={(v) => setMode(v as 'single' | 'bulk')}
        >
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
                    {objectType === 'trace' ? 'Trace ID' : 'Span ID'}
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder={
                        objectType === 'trace'
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
                    {objectType === 'trace' ? 'Trace IDs' : 'Span IDs'}
                  </FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder={`Enter ${objectType.toLowerCase()} IDs, one per line or comma-separated`}
                      rows={6}
                      className="font-mono text-sm"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Multiple IDs separated by commas, spaces, or newlines.
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
                  value={field.value ?? 0}
                  onChange={(e) =>
                    field.onChange(parseInt(e.target.value, 10) || 0)
                  }
                />
              </FormControl>
              <FormDescription>
                Higher priority items are surfaced first (0–100).
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
