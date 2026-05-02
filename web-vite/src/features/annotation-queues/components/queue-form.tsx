import { useQuery } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ChevronsUpDown, Loader2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
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
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import { scoreConfigsQueryOptions } from '@/features/scores/api/queries'
import type { AnnotationQueue, CreateQueueRequest } from '../api/types'

const schema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().optional(),
  instructions: z.string().optional(),
  status: z.enum(['active', 'paused', 'archived']).optional(),
  lock_timeout_seconds: z.number().min(60).max(3600).optional(),
  auto_assignment: z.boolean().optional(),
  score_config_ids: z.array(z.string()).optional(),
})

type QueueFormValues = z.infer<typeof schema>

interface QueueFormProps {
  projectId: string
  queue?: AnnotationQueue
  onSubmit: (data: CreateQueueRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

/**
 * Shared queue create / update form. The same shape powers
 * `settings-dialog` (update) and any future create-queue flow. Fields
 * map directly onto `CreateQueueRequest` plus the `status` toggle that
 * the backend's `UpdateQueueRequest` exposes for editing — we read it
 * from `queue?.status` and only render the field in edit mode.
 */
export function QueueForm({
  projectId,
  queue,
  onSubmit,
  onCancel,
  isLoading,
}: QueueFormProps) {
  const configsQuery = useQuery(scoreConfigsQueryOptions(projectId))
  const isEditing = !!queue

  const form = useForm<QueueFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: queue?.name ?? '',
      description: queue?.description ?? '',
      instructions: queue?.instructions ?? '',
      status: queue?.status ?? 'active',
      lock_timeout_seconds: queue?.settings?.lock_timeout_seconds ?? 300,
      auto_assignment: queue?.settings?.auto_assignment ?? false,
      score_config_ids: queue?.score_config_ids ?? [],
    },
  })

  const handleSubmit = (values: QueueFormValues) => {
    onSubmit({
      name: values.name,
      description: values.description || undefined,
      instructions: values.instructions || undefined,
      // On update: preserve the user's intent to clear (`[]`); on
      // create: omit empty so the backend's "no default attached
      // configs" branch triggers.
      score_config_ids: isEditing
        ? (values.score_config_ids ?? [])
        : values.score_config_ids && values.score_config_ids.length > 0
          ? values.score_config_ids
          : undefined,
      settings: {
        lock_timeout_seconds: values.lock_timeout_seconds,
        auto_assignment: values.auto_assignment,
      },
    })
  }

  const selectedIds = form.watch('score_config_ids') ?? []
  const configs = configsQuery.data?.data ?? []

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
                These instructions are shown to annotators when reviewing items.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="score_config_ids"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Score Configurations</FormLabel>
              <FormDescription>
                Pick which scores annotators provide when reviewing items.
              </FormDescription>
              {configsQuery.isLoading ? (
                <div className="flex items-center gap-2 py-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading score configs...
                </div>
              ) : configs.length === 0 ? (
                <p className="text-sm text-muted-foreground py-2">
                  No score configurations found. Create score configs in the
                  Scores section first.
                </p>
              ) : (
                <div className="space-y-3">
                  <Popover>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant="outline"
                          role="combobox"
                          className={cn(
                            'w-full justify-between',
                            selectedIds.length === 0 && 'text-muted-foreground',
                          )}
                        >
                          {selectedIds.length > 0
                            ? `${selectedIds.length} score config(s) selected`
                            : 'Select score configs...'}
                          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                        </Button>
                      </FormControl>
                    </PopoverTrigger>
                    <PopoverContent className="w-[400px] p-0" align="start">
                      <Command>
                        <CommandInput placeholder="Search score configs..." />
                        <CommandList>
                          <CommandEmpty>No score configs found.</CommandEmpty>
                          <CommandGroup>
                            {configs.map((c) => {
                              const isSelected = selectedIds.includes(c.id)
                              return (
                                <CommandItem
                                  key={c.id}
                                  value={c.name}
                                  onSelect={() => {
                                    field.onChange(
                                      isSelected
                                        ? selectedIds.filter((id) => id !== c.id)
                                        : [...selectedIds, c.id],
                                    )
                                  }}
                                >
                                  <Checkbox
                                    checked={isSelected}
                                    className="mr-2"
                                  />
                                  <div className="flex flex-col">
                                    <span>{c.name}</span>
                                    <span className="text-xs text-muted-foreground">
                                      {c.type}
                                      {c.description ? ` • ${c.description}` : ''}
                                    </span>
                                  </div>
                                </CommandItem>
                              )
                            })}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>

                  {selectedIds.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {selectedIds.map((id) => {
                        const c = configs.find((x) => x.id === id)
                        if (!c) return null
                        return (
                          <Badge key={id} variant="secondary" className="gap-1">
                            {c.name}
                            <button
                              type="button"
                              className="ml-1 rounded-full outline-none hover:bg-secondary-foreground/20"
                              onClick={() =>
                                field.onChange(selectedIds.filter((x) => x !== id))
                              }
                            >
                              ×
                            </button>
                          </Badge>
                        )
                      })}
                    </div>
                  )}
                </div>
              )}
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
                <Select onValueChange={field.onChange} value={field.value}>
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
                  onValueChange={(value) =>
                    field.onChange(parseInt(value, 10))
                  }
                  value={String(field.value ?? 300)}
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
                    checked={field.value ?? false}
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
