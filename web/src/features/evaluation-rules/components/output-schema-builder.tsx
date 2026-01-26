'use client'

import { useCallback, useMemo } from 'react'
import { Plus, X, FileJson, Sparkles, Info, GripVertical } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { OutputField } from '../types'

// Field types available for output schema
const FIELD_TYPES = [
  {
    value: 'numeric',
    label: 'Number',
    description: 'Numeric value with optional min/max bounds',
  },
  {
    value: 'categorical',
    label: 'Category',
    description: 'Select from predefined categories',
  },
  {
    value: 'boolean',
    label: 'Boolean',
    description: 'True or false value',
  },
] as const

// Preset configurations for common evaluation patterns
interface SchemaPreset {
  id: string
  name: string
  description: string
  fields: OutputField[]
}

const SCHEMA_PRESETS: SchemaPreset[] = [
  {
    id: 'simple_score',
    name: 'Simple Score',
    description: 'Single numeric score from 0-1',
    fields: [
      {
        name: 'score',
        type: 'numeric',
        description: 'Quality score from 0 to 1',
        min_value: 0,
        max_value: 1,
      },
    ],
  },
  {
    id: 'score_reasoning',
    name: 'Score + Reasoning',
    description: 'Score with explanation text',
    fields: [
      {
        name: 'score',
        type: 'numeric',
        description: 'Quality score from 0 to 1',
        min_value: 0,
        max_value: 1,
      },
      {
        name: 'reasoning',
        type: 'categorical',
        description: 'Brief explanation of the score',
        categories: [],
      },
    ],
  },
  {
    id: 'multi_metric',
    name: 'Multi-Metric',
    description: 'Multiple evaluation dimensions',
    fields: [
      {
        name: 'relevance',
        type: 'numeric',
        description: 'How relevant is the response',
        min_value: 0,
        max_value: 1,
      },
      {
        name: 'accuracy',
        type: 'numeric',
        description: 'Factual accuracy of the response',
        min_value: 0,
        max_value: 1,
      },
      {
        name: 'completeness',
        type: 'numeric',
        description: 'How complete is the response',
        min_value: 0,
        max_value: 1,
      },
    ],
  },
  {
    id: 'pass_fail',
    name: 'Pass/Fail',
    description: 'Binary evaluation with reasoning',
    fields: [
      {
        name: 'passed',
        type: 'boolean',
        description: 'Whether the response passes evaluation',
      },
      {
        name: 'issues',
        type: 'categorical',
        description: 'Any issues found (if failed)',
        categories: [],
      },
    ],
  },
]

interface OutputSchemaBuilderProps {
  value: OutputField[]
  onChange: (fields: OutputField[]) => void
  disabled?: boolean
  maxFields?: number
}

/**
 * Output Schema Builder for LLM evaluation rules.
 *
 * Features:
 * - Define expected JSON output structure
 * - Support for numeric, categorical, and boolean field types
 * - Preset configurations for common patterns
 * - Real-time JSON preview
 * - Field reordering and management
 */
export function OutputSchemaBuilder({
  value,
  onChange,
  disabled = false,
  maxFields = 10,
}: OutputSchemaBuilderProps) {
  const canAddMore = value.length < maxFields

  // Add a new field
  const addField = useCallback(() => {
    const newField: OutputField = {
      name: '',
      type: 'numeric',
      description: '',
      min_value: 0,
      max_value: 1,
    }
    onChange([...value, newField])
  }, [value, onChange])

  // Update a field at index
  const updateField = useCallback(
    (index: number, updates: Partial<OutputField>) => {
      const newFields = [...value]
      newFields[index] = { ...newFields[index], ...updates }

      // Clear type-specific fields when type changes
      if (updates.type) {
        if (updates.type === 'numeric') {
          newFields[index] = {
            ...newFields[index],
            min_value: 0,
            max_value: 1,
            categories: undefined,
          }
        } else if (updates.type === 'categorical') {
          newFields[index] = {
            ...newFields[index],
            min_value: undefined,
            max_value: undefined,
            categories: [],
          }
        } else if (updates.type === 'boolean') {
          newFields[index] = {
            ...newFields[index],
            min_value: undefined,
            max_value: undefined,
            categories: undefined,
          }
        }
      }

      onChange(newFields)
    },
    [value, onChange]
  )

  // Remove a field at index
  const removeField = useCallback(
    (index: number) => {
      onChange(value.filter((_, i) => i !== index))
    },
    [value, onChange]
  )

  // Apply a preset
  const applyPreset = useCallback(
    (preset: SchemaPreset) => {
      onChange([...preset.fields])
    },
    [onChange]
  )

  // Generate JSON preview
  const jsonPreview = useMemo(() => {
    const example: Record<string, unknown> = {}
    value.forEach((field) => {
      if (!field.name) return
      switch (field.type) {
        case 'numeric':
          example[field.name] = field.min_value !== undefined && field.max_value !== undefined
            ? (field.min_value + field.max_value) / 2
            : 0.5
          break
        case 'categorical':
          example[field.name] = field.categories?.length ? field.categories[0] : 'example'
          break
        case 'boolean':
          example[field.name] = true
          break
      }
    })
    return JSON.stringify(example, null, 2)
  }, [value])

  return (
    <div className="space-y-4 border rounded-lg p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FileJson className="h-5 w-5 text-primary" />
          <h4 className="font-medium">Output Schema</h4>
          {value.length > 0 && (
            <Badge variant="secondary" className="text-xs">
              {value.length} {value.length === 1 ? 'field' : 'fields'}
            </Badge>
          )}
        </div>

        {/* Preset buttons */}
        <div className="flex items-center gap-1">
          {SCHEMA_PRESETS.slice(0, 3).map((preset) => (
            <Tooltip key={preset.id}>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={() => applyPreset(preset)}
                  disabled={disabled}
                >
                  <Sparkles className="mr-1 h-3 w-3" />
                  {preset.name}
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p className="text-xs">{preset.description}</p>
              </TooltipContent>
            </Tooltip>
          ))}
        </div>
      </div>

      <p className="text-sm text-muted-foreground">
        Define the expected JSON output structure from the LLM evaluator.
      </p>

      {value.length === 0 ? (
        <div className="rounded-lg border border-dashed p-4 text-center">
          <p className="text-sm text-muted-foreground mb-2">
            No output fields defined. Add fields or use a preset.
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addField}
            disabled={disabled}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Field
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {/* Header row */}
          <div className="flex items-center gap-2 text-xs text-muted-foreground px-1">
            <div className="w-6" />
            <div className="w-[140px]">Field Name</div>
            <div className="w-[100px]">Type</div>
            <div className="flex-1">Configuration</div>
            <div className="w-8" />
          </div>

          {/* Field rows */}
          {value.map((field, index) => (
            <FieldRow
              key={index}
              field={field}
              index={index}
              disabled={disabled}
              onUpdate={(updates) => updateField(index, updates)}
              onRemove={() => removeField(index)}
            />
          ))}

          {/* Add field button */}
          {canAddMore && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-full border border-dashed"
              onClick={addField}
              disabled={disabled}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Field
            </Button>
          )}
        </div>
      )}

      {/* JSON Preview */}
      {value.length > 0 && value.some((f) => f.name) && (
        <div className="space-y-2">
          <Label className="text-sm">JSON Preview</Label>
          <pre className="rounded-lg bg-muted p-3 text-xs font-mono overflow-x-auto">
            {jsonPreview}
          </pre>
        </div>
      )}
    </div>
  )
}

// Individual field row component
interface FieldRowProps {
  field: OutputField
  index: number
  disabled: boolean
  onUpdate: (updates: Partial<OutputField>) => void
  onRemove: () => void
}

function FieldRow({ field, index, disabled, onUpdate, onRemove }: FieldRowProps) {
  return (
    <div className="flex items-start gap-2 group">
      {/* Drag handle (visual only for now) */}
      <div className="w-6 pt-2 opacity-30 group-hover:opacity-50">
        <GripVertical className="h-4 w-4" />
      </div>

      {/* Field name */}
      <div className="w-[140px]">
        <Input
          value={field.name}
          onChange={(e) => onUpdate({ name: e.target.value.replace(/\s/g, '_').toLowerCase() })}
          placeholder="field_name"
          className="h-8 font-mono text-sm"
          disabled={disabled}
        />
      </div>

      {/* Type selector */}
      <div className="w-[100px]">
        <Select
          value={field.type}
          onValueChange={(type) => onUpdate({ type: type as OutputField['type'] })}
          disabled={disabled}
        >
          <SelectTrigger className="h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {FIELD_TYPES.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                <div className="flex items-center gap-2">
                  <span>{type.label}</span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent side="right">
                      <p className="text-xs max-w-[200px]">{type.description}</p>
                    </TooltipContent>
                  </Tooltip>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Type-specific configuration */}
      <div className="flex-1">
        <TypeConfig field={field} disabled={disabled} onUpdate={onUpdate} />
      </div>

      {/* Remove button */}
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className={cn(
          'h-8 w-8 shrink-0',
          'opacity-0 group-hover:opacity-100 transition-opacity'
        )}
        onClick={onRemove}
        disabled={disabled}
      >
        <X className="h-4 w-4" />
        <span className="sr-only">Remove field</span>
      </Button>
    </div>
  )
}

// Type-specific configuration component
interface TypeConfigProps {
  field: OutputField
  disabled: boolean
  onUpdate: (updates: Partial<OutputField>) => void
}

function TypeConfig({ field, disabled, onUpdate }: TypeConfigProps) {
  switch (field.type) {
    case 'numeric':
      return (
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1">
            <Label className="text-xs text-muted-foreground">Min:</Label>
            <Input
              type="number"
              value={field.min_value ?? 0}
              onChange={(e) => onUpdate({ min_value: parseFloat(e.target.value) || 0 })}
              className="h-8 w-16 text-sm"
              disabled={disabled}
            />
          </div>
          <div className="flex items-center gap-1">
            <Label className="text-xs text-muted-foreground">Max:</Label>
            <Input
              type="number"
              value={field.max_value ?? 1}
              onChange={(e) => onUpdate({ max_value: parseFloat(e.target.value) || 1 })}
              className="h-8 w-16 text-sm"
              disabled={disabled}
            />
          </div>
          <Input
            value={field.description || ''}
            onChange={(e) => onUpdate({ description: e.target.value })}
            placeholder="Description..."
            className="h-8 flex-1 text-sm"
            disabled={disabled}
          />
        </div>
      )

    case 'categorical':
      return (
        <div className="flex items-center gap-2">
          <Input
            value={field.categories?.join(', ') || ''}
            onChange={(e) =>
              onUpdate({
                categories: e.target.value
                  .split(',')
                  .map((c) => c.trim())
                  .filter(Boolean),
              })
            }
            placeholder="Categories (comma-separated) or leave empty for free-form"
            className="h-8 flex-1 text-sm"
            disabled={disabled}
          />
        </div>
      )

    case 'boolean':
      return (
        <div className="flex items-center gap-2">
          <Input
            value={field.description || ''}
            onChange={(e) => onUpdate({ description: e.target.value })}
            placeholder="Description (e.g., 'Whether the response is appropriate')"
            className="h-8 flex-1 text-sm"
            disabled={disabled}
          />
        </div>
      )

    default:
      return null
  }
}
