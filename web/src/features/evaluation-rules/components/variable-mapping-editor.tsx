'use client'

import { useCallback, useMemo } from 'react'
import { Plus, X, Variable, Sparkles, Info, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
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
import type { VariableMap } from '../types'

// Source types for variable mapping
type VariableSource = VariableMap['source']

// Source options for the dropdown
const SOURCE_OPTIONS: { value: VariableSource; label: string; description: string }[] = [
  {
    value: 'span_input',
    label: 'Span Input',
    description: 'Input data sent to the span (e.g., user prompt)',
  },
  {
    value: 'span_output',
    label: 'Span Output',
    description: 'Output data from the span (e.g., AI completion)',
  },
  {
    value: 'span_metadata',
    label: 'Span Metadata',
    description: 'Metadata attributes on the span',
  },
  {
    value: 'trace_input',
    label: 'Trace Input',
    description: 'Input data from the root trace span',
  },
]

// Smart defaults for common variable names (from Langfuse pattern)
const SMART_DEFAULTS: Record<string, { source: VariableSource; json_path: string }> = {
  // Input-related variables
  input: { source: 'span_input', json_path: '' },
  query: { source: 'span_input', json_path: '' },
  question: { source: 'span_input', json_path: '' },
  prompt: { source: 'span_input', json_path: '' },
  user_message: { source: 'span_input', json_path: '' },

  // Output-related variables
  output: { source: 'span_output', json_path: '' },
  response: { source: 'span_output', json_path: '' },
  answer: { source: 'span_output', json_path: '' },
  completion: { source: 'span_output', json_path: '' },
  assistant_message: { source: 'span_output', json_path: '' },

  // Context variables
  context: { source: 'span_metadata', json_path: 'context' },
  metadata: { source: 'span_metadata', json_path: '' },
  model: { source: 'span_metadata', json_path: 'model_name' },
  user_id: { source: 'trace_input', json_path: 'user_id' },
}

// Common JSON paths for autocomplete
const COMMON_JSON_PATHS = [
  '',
  'messages',
  'messages[0].content',
  'messages[-1].content',
  'content',
  'text',
  'model_name',
  'temperature',
  'user_id',
  'session_id',
]

// Preset configurations
interface MappingPreset {
  id: string
  name: string
  description: string
  mappings: Omit<VariableMap, 'id'>[]
}

const MAPPING_PRESETS: MappingPreset[] = [
  {
    id: 'llm_io',
    name: 'LLM Input/Output',
    description: 'Standard input/output mapping for LLM evaluations',
    mappings: [
      { variable_name: 'input', source: 'span_input', json_path: '' },
      { variable_name: 'output', source: 'span_output', json_path: '' },
    ],
  },
  {
    id: 'chat_messages',
    name: 'Chat Messages',
    description: 'Map chat message arrays for conversation analysis',
    mappings: [
      { variable_name: 'messages', source: 'span_input', json_path: 'messages' },
      { variable_name: 'response', source: 'span_output', json_path: '' },
    ],
  },
  {
    id: 'rag_context',
    name: 'RAG with Context',
    description: 'Input, output, and retrieved context for RAG evaluation',
    mappings: [
      { variable_name: 'query', source: 'span_input', json_path: '' },
      { variable_name: 'context', source: 'span_metadata', json_path: 'retrieved_context' },
      { variable_name: 'answer', source: 'span_output', json_path: '' },
    ],
  },
]

interface LocalMapping {
  id: string
  variable_name: string
  source: VariableSource
  json_path: string
}

interface VariableMappingEditorProps {
  value: VariableMap[]
  onChange: (mappings: VariableMap[]) => void
  promptTemplate?: string // Optional: to detect variables from prompt
  disabled?: boolean
  maxMappings?: number
}

/**
 * Variable Mapping Editor for evaluation rules.
 *
 * Features:
 * - Smart defaults based on variable names
 * - Preset configurations for common patterns
 * - JSON path configuration for nested data
 * - Auto-detection of variables from prompt template
 */
export function VariableMappingEditor({
  value,
  onChange,
  promptTemplate,
  disabled = false,
  maxMappings = 20,
}: VariableMappingEditorProps) {
  // Convert external to internal format with stable IDs for React keys
  // IDs are derived from content to ensure stability
  const localMappings = useMemo(
    () =>
      value.map((m, index) => ({
        id: `mapping-${index}-${m.variable_name}-${m.source}`,
        variable_name: m.variable_name,
        source: m.source,
        json_path: m.json_path,
      })),
    [value]
  )

  // Detect variables from prompt template
  const detectedVariables = useMemo(() => {
    if (!promptTemplate) return []
    // Match {{variable}}, {variable}, and $variable patterns
    const matches = promptTemplate.match(/\{\{?\s*(\w+)\s*\}?\}|\$(\w+)/g) || []
    const variables = matches.map((m) => {
      const match = m.match(/\{\{?\s*(\w+)\s*\}?\}|\$(\w+)/)
      return match?.[1] || match?.[2] || ''
    })
    return [...new Set(variables)].filter(Boolean)
  }, [promptTemplate])

  // Variables that are detected but not mapped
  const unmappedVariables = useMemo(() => {
    const mappedNames = new Set(localMappings.map((m) => m.variable_name))
    return detectedVariables.filter((v) => !mappedNames.has(v))
  }, [detectedVariables, localMappings])

  const canAddMore = localMappings.length < maxMappings

  // Emit changes to parent (converts local back to external format)
  const emitChange = useCallback(
    (mappings: LocalMapping[]) => {
      const validMappings: VariableMap[] = mappings
        .filter((m) => m.variable_name.trim())
        .map((m) => ({
          variable_name: m.variable_name.trim(),
          source: m.source,
          json_path: m.json_path,
        }))
      onChange(validMappings)
    },
    [onChange]
  )

  const addMapping = useCallback(
    (variableName?: string) => {
      const name = variableName || ''
      const defaults = SMART_DEFAULTS[name.toLowerCase()] || {
        source: 'span_input' as VariableSource,
        json_path: '',
      }

      const newMappings: LocalMapping[] = [
        ...localMappings,
        {
          id: `new-${Date.now()}`,
          variable_name: name,
          source: defaults.source,
          json_path: defaults.json_path,
        },
      ]
      emitChange(newMappings)
    },
    [localMappings, emitChange]
  )

  const updateMapping = useCallback(
    (id: string, updates: Partial<LocalMapping>) => {
      const newMappings = localMappings.map((m) =>
        m.id === id ? { ...m, ...updates } : m
      )
      emitChange(newMappings)
    },
    [localMappings, emitChange]
  )

  const removeMapping = useCallback(
    (id: string) => {
      const newMappings = localMappings.filter((m) => m.id !== id)
      emitChange(newMappings)
    },
    [localMappings, emitChange]
  )

  const applyPreset = useCallback(
    (preset: MappingPreset) => {
      const newMappings: LocalMapping[] = preset.mappings.map((m, index) => ({
        id: `preset-${index}`,
        variable_name: m.variable_name,
        source: m.source,
        json_path: m.json_path,
      }))
      emitChange(newMappings)
    },
    [emitChange]
  )

  // Apply smart default when variable name changes
  const handleVariableNameChange = useCallback(
    (id: string, name: string) => {
      const defaults = SMART_DEFAULTS[name.toLowerCase()]
      const updates: Partial<LocalMapping> = { variable_name: name }

      // Only apply defaults if source wasn't manually changed
      const currentMapping = localMappings.find((m) => m.id === id)
      if (defaults && currentMapping?.source === 'span_input' && !currentMapping?.json_path) {
        updates.source = defaults.source
        updates.json_path = defaults.json_path
      }

      updateMapping(id, updates)
    },
    [localMappings, updateMapping]
  )

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Variable className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-medium">Variable Mappings</span>
          {localMappings.length > 0 && (
            <Badge variant="secondary" className="text-xs">
              {localMappings.length} {localMappings.length === 1 ? 'variable' : 'variables'}
            </Badge>
          )}
        </div>

        {/* Preset buttons */}
        <div className="flex items-center gap-2">
          {MAPPING_PRESETS.map((preset) => (
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

      {/* Unmapped variables warning */}
      {unmappedVariables.length > 0 && (
        <div className="flex items-start gap-2 rounded-lg bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800 p-3">
          <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
          <div className="flex-1">
            <p className="text-sm text-amber-800 dark:text-amber-200">
              Detected unmapped variables in your prompt:
            </p>
            <div className="flex flex-wrap gap-1 mt-1">
              {unmappedVariables.map((v) => (
                <Button
                  key={v}
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-6 text-xs bg-white dark:bg-background"
                  onClick={() => addMapping(v)}
                  disabled={disabled}
                >
                  <Plus className="mr-1 h-3 w-3" />
                  {`{{${v}}}`}
                </Button>
              ))}
            </div>
          </div>
        </div>
      )}

      {localMappings.length === 0 ? (
        <div className="rounded-lg border border-dashed p-4 text-center">
          <p className="text-sm text-muted-foreground mb-2">
            No variable mappings configured. Default span input/output will be used.
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => addMapping()}
            disabled={disabled}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Mapping
          </Button>
        </div>
      ) : (
        <div className="space-y-2">
          {/* Header row */}
          <div className="flex items-center gap-2 text-xs text-muted-foreground px-1">
            <div className="w-[140px]">Variable</div>
            <div className="w-[140px]">Source</div>
            <div className="flex-1">JSON Path (optional)</div>
            <div className="w-8" />
          </div>

          {localMappings.map((mapping) => (
            <MappingRow
              key={mapping.id}
              mapping={mapping}
              disabled={disabled}
              onVariableNameChange={(name) => handleVariableNameChange(mapping.id, name)}
              onSourceChange={(source) => updateMapping(mapping.id, { source })}
              onPathChange={(path) => updateMapping(mapping.id, { json_path: path })}
              onRemove={() => removeMapping(mapping.id)}
            />
          ))}

          {canAddMore && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-full border border-dashed"
              onClick={() => addMapping()}
              disabled={disabled}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Mapping
            </Button>
          )}
        </div>
      )}

      <p className="text-xs text-muted-foreground">
        Use variables in your prompt template with {'{{'}<span className="font-mono">variable_name</span>{'}}'}
        syntax. Variables will be replaced with actual span data at evaluation time.
      </p>
    </div>
  )
}

// Individual mapping row component
interface MappingRowProps {
  mapping: LocalMapping
  disabled: boolean
  onVariableNameChange: (name: string) => void
  onSourceChange: (source: VariableSource) => void
  onPathChange: (path: string) => void
  onRemove: () => void
}

function MappingRow({
  mapping,
  disabled,
  onVariableNameChange,
  onSourceChange,
  onPathChange,
  onRemove,
}: MappingRowProps) {
  return (
    <div className="flex items-center gap-2 group">
      {/* Variable name */}
      <div className="w-[140px]">
        <Input
          value={mapping.variable_name}
          onChange={(e) => onVariableNameChange(e.target.value)}
          placeholder="variable_name"
          className="h-8 font-mono text-sm"
          disabled={disabled}
        />
      </div>

      {/* Source selector */}
      <div className="w-[140px]">
        <Select
          value={mapping.source}
          onValueChange={onSourceChange}
          disabled={disabled}
        >
          <SelectTrigger className="h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {SOURCE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                <div className="flex items-center gap-2">
                  <span>{opt.label}</span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent side="right">
                      <p className="text-xs max-w-[200px]">{opt.description}</p>
                    </TooltipContent>
                  </Tooltip>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* JSON path with suggestions */}
      <div className="flex-1">
        <Select
          value={mapping.json_path}
          onValueChange={onPathChange}
          disabled={disabled}
        >
          <SelectTrigger className="h-8 font-mono text-sm">
            <SelectValue placeholder="(root)" />
          </SelectTrigger>
          <SelectContent>
            {COMMON_JSON_PATHS.map((path) => (
              <SelectItem key={path || '_root_'} value={path}>
                {path || '(root - entire value)'}
              </SelectItem>
            ))}
            {/* Allow custom input */}
            <div className="px-2 py-1.5 border-t">
              <Input
                value={mapping.json_path}
                onChange={(e) => onPathChange(e.target.value)}
                placeholder="Custom path..."
                className="h-7 text-xs"
                onClick={(e) => e.stopPropagation()}
              />
            </div>
          </SelectContent>
        </Select>
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
        <span className="sr-only">Remove mapping</span>
      </Button>
    </div>
  )
}
