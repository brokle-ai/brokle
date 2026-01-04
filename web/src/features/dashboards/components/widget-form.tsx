'use client'

import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { QueryBuilder } from './query-builder'
import type { Widget, WidgetType, WidgetQuery } from '../types'

interface WidgetFormProps {
  widget?: Widget
  onSubmit: (widget: Omit<Widget, 'id'> | Widget) => void
  onCancel: () => void
  isLoading?: boolean
  // Pre-selected type from palette
  defaultType?: WidgetType
}

const WIDGET_TYPES: Array<{ value: WidgetType; label: string; description: string }> = [
  { value: 'stat', label: 'Stat', description: 'Single value metric' },
  { value: 'time_series', label: 'Time Series', description: 'Line chart over time' },
  { value: 'table', label: 'Table', description: 'Tabular data display' },
  { value: 'bar', label: 'Bar Chart', description: 'Categorical comparisons' },
  { value: 'pie', label: 'Pie Chart', description: 'Proportional distribution' },
  { value: 'heatmap', label: 'Heatmap', description: 'Value density visualization' },
  { value: 'histogram', label: 'Histogram', description: 'Distribution analysis' },
  { value: 'trace_list', label: 'Trace List', description: 'List of traces' },
  { value: 'text', label: 'Text', description: 'Static text or markdown' },
]

const DEFAULT_QUERY: WidgetQuery = {
  view: 'traces',
  measures: [],
  dimensions: [],
  filters: [],
}

export function WidgetForm({
  widget,
  onSubmit,
  onCancel,
  isLoading,
  defaultType,
}: WidgetFormProps) {
  const [title, setTitle] = useState(widget?.title ?? '')
  const [description, setDescription] = useState(widget?.description ?? '')
  const [type, setType] = useState<WidgetType>(widget?.type ?? defaultType ?? 'stat')
  const [query, setQuery] = useState<WidgetQuery>(widget?.query ?? DEFAULT_QUERY)

  const isEditing = Boolean(widget?.id)
  const isTextWidget = type === 'text'

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const widgetData = {
      ...(widget?.id ? { id: widget.id } : {}),
      type,
      title: title.trim(),
      description: description.trim() || undefined,
      query,
      config: widget?.config ?? {},
    }

    onSubmit(widgetData as Widget)
  }

  const handleQueryChange = useCallback((newQuery: WidgetQuery) => {
    setQuery(newQuery)
  }, [])

  const isValid = title.trim().length > 0 && (isTextWidget || query.measures.length > 0)

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Info */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="title">Title *</Label>
          <Input
            id="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Enter widget title..."
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional description..."
            rows={2}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="type">Widget Type *</Label>
          <Select
            value={type}
            onValueChange={(val) => setType(val as WidgetType)}
          >
            <SelectTrigger id="type">
              <SelectValue placeholder="Select widget type..." />
            </SelectTrigger>
            <SelectContent>
              {WIDGET_TYPES.map((wt) => (
                <SelectItem key={wt.value} value={wt.value}>
                  <div className="flex flex-col">
                    <span>{wt.label}</span>
                    <span className="text-xs text-muted-foreground">
                      {wt.description}
                    </span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <Separator />

      {/* Query Builder (not for text widgets) */}
      {!isTextWidget ? (
        <div className="space-y-2">
          <Label className="text-base font-medium">Query Configuration</Label>
          <p className="text-sm text-muted-foreground mb-4">
            Configure the data source and metrics for this widget.
          </p>
          <QueryBuilder
            query={query}
            onQueryChange={handleQueryChange}
            disabled={isLoading}
          />
        </div>
      ) : (
        <div className="rounded-lg border bg-muted/50 p-4">
          <p className="text-sm text-muted-foreground">
            Text widgets display static content. Configure the content after creating the widget.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading || !isValid}>
          {isLoading ? 'Saving...' : isEditing ? 'Save Changes' : 'Create Widget'}
        </Button>
      </div>
    </form>
  )
}
