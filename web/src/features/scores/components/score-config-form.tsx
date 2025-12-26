'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, X } from 'lucide-react'
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
import type { ScoreConfig, CreateScoreConfigRequest } from '../types'

const scoreConfigSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  description: z.string().optional(),
  data_type: z.enum(['NUMERIC', 'CATEGORICAL', 'BOOLEAN']),
  min_value: z.number().optional(),
  max_value: z.number().optional(),
  categories: z.array(z.string()).optional(),
}).superRefine((data, ctx) => {
  // Require at least one category for CATEGORICAL type
  if (data.data_type === 'CATEGORICAL') {
    if (!data.categories || data.categories.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'At least one category is required',
        path: ['categories'],
      })
    }
  }
})

type ScoreConfigFormData = z.infer<typeof scoreConfigSchema>

interface ScoreConfigFormProps {
  config?: ScoreConfig
  onSubmit: (data: CreateScoreConfigRequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function ScoreConfigForm({
  config,
  onSubmit,
  onCancel,
  isLoading,
}: ScoreConfigFormProps) {
  const isEditMode = !!config

  const form = useForm<ScoreConfigFormData>({
    resolver: zodResolver(scoreConfigSchema),
    defaultValues: {
      name: config?.name ?? '',
      description: config?.description ?? '',
      data_type: config?.data_type ?? 'NUMERIC',
      min_value: config?.min_value,
      max_value: config?.max_value,
      categories: config?.categories ?? [],
    },
  })

  const dataType = form.watch('data_type')
  const categories = form.watch('categories') || []
  const [categoryInput, setCategoryInput] = useState('')

  const addCategory = () => {
    const trimmed = categoryInput.trim()
    if (!trimmed) return

    if (categories.includes(trimmed)) {
      form.setError('categories', { message: 'Category already exists' })
      return
    }

    form.setValue('categories', [...categories, trimmed])
    setCategoryInput('')
    form.clearErrors('categories')
  }

  const removeCategory = (index: number) => {
    const updated = categories.filter((_, i) => i !== index)
    form.setValue('categories', updated)
  }

  const handleCategoryKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      addCategory()
    }
  }

  const handleSubmit = (data: ScoreConfigFormData) => {
    onSubmit({
      name: data.name,
      description: data.description || undefined,
      data_type: data.data_type,
      min_value: data.data_type === 'NUMERIC' ? data.min_value : undefined,
      max_value: data.data_type === 'NUMERIC' ? data.max_value : undefined,
      categories: data.data_type === 'CATEGORICAL' ? data.categories : undefined,
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
                <Input placeholder="e.g., relevance, accuracy" {...field} />
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
                  placeholder="Describe what this score measures"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="data_type"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Data Type</FormLabel>
              <Select
                onValueChange={(value) => {
                  field.onChange(value)
                  // Clear categories when switching away from CATEGORICAL
                  if (value !== 'CATEGORICAL') {
                    form.setValue('categories', [])
                  }
                  setCategoryInput('')
                  form.clearErrors('categories')
                }}
                defaultValue={field.value}
                disabled={isEditMode}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select data type" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="NUMERIC">
                    Numeric (0-1 or custom range)
                  </SelectItem>
                  <SelectItem value="CATEGORICAL">
                    Categorical (predefined values)
                  </SelectItem>
                  <SelectItem value="BOOLEAN">Boolean (true/false)</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                {isEditMode
                  ? 'Data type cannot be changed after creation'
                  : 'The type of value this score will hold'}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {dataType === 'NUMERIC' && (
          <div className="grid grid-cols-2 gap-4">
            <FormField
              control={form.control}
              name="min_value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Min Value</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      step="0.01"
                      placeholder="0"
                      {...field}
                      value={field.value ?? ''}
                      onChange={(e) =>
                        field.onChange(
                          e.target.value ? parseFloat(e.target.value) : undefined
                        )
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="max_value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Max Value</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      step="0.01"
                      placeholder="1"
                      {...field}
                      value={field.value ?? ''}
                      onChange={(e) =>
                        field.onChange(
                          e.target.value ? parseFloat(e.target.value) : undefined
                        )
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )}

        {dataType === 'CATEGORICAL' && (
          <div className="space-y-3">
            <FormLabel>Categories *</FormLabel>

            <div className="flex gap-2">
              <Input
                placeholder="Enter category name"
                value={categoryInput}
                onChange={(e) => setCategoryInput(e.target.value)}
                onKeyDown={handleCategoryKeyDown}
              />
              <Button
                type="button"
                variant="outline"
                onClick={addCategory}
                disabled={!categoryInput.trim()}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>

            {categories.length > 0 ? (
              <div className="space-y-2">
                {categories.map((category, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between rounded-md border px-3 py-2"
                  >
                    <span className="text-sm">{category}</span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeCategory(index)}
                      className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                No categories added yet. Add at least one category.
              </p>
            )}

            {form.formState.errors.categories && (
              <p className="text-sm text-destructive">
                {form.formState.errors.categories.message}
              </p>
            )}
          </div>
        )}

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Saving...' : config ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
