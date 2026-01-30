'use client'

import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import Editor, { type OnMount } from '@monaco-editor/react'
import type * as MonacoEditor from 'monaco-editor'
import { useTheme } from 'next-themes'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import {
  ChevronDown,
  Plus,
  Trash2,
  Sparkles,
  Loader2,
  AlertCircle,
  Wrench,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Tool, ToolChoice } from '../types'
import { createEmptyTool, validateTool } from '../types/tools'

interface ToolsEditorProps {
  tools: Tool[]
  toolChoice: ToolChoice | null
  onToolsChange: (tools: Tool[]) => void
  onToolChoiceChange: (choice: ToolChoice | null) => void
  provider?: string // To show provider compatibility warnings
  className?: string
  disabled?: boolean
}

const TOOL_CHOICE_OPTIONS = [
  { value: 'auto', label: 'Auto', description: 'Model decides whether to use tools' },
  { value: 'none', label: 'None', description: 'Model will not use any tools' },
  { value: 'required', label: 'Required', description: 'Model must use at least one tool' },
] as const

const EXAMPLE_TOOL: Tool = {
  id: 'example',
  type: 'function',
  function: {
    name: 'get_weather',
    description: 'Get the current weather for a location',
    parameters: {
      type: 'object',
      properties: {
        location: {
          type: 'string',
          description: 'The city and state, e.g. San Francisco, CA',
        },
        unit: {
          type: 'string',
          enum: ['celsius', 'fahrenheit'],
          description: 'Temperature unit',
        },
      },
      required: ['location'],
    },
  },
}

export function ToolsEditor({
  tools,
  toolChoice,
  onToolsChange,
  onToolChoiceChange,
  provider,
  className,
  disabled = false,
}: ToolsEditorProps) {
  const { resolvedTheme } = useTheme()
  const isDarkMode = resolvedTheme === 'dark'
  const [expandedToolId, setExpandedToolId] = useState<string | null>(null)
  const [isOpen, setIsOpen] = useState(tools.length > 0)
  const editorRef = useRef<MonacoEditor.editor.IStandaloneCodeEditor | null>(null)

  // Add a new empty tool
  const handleAddTool = useCallback(() => {
    const newTool = createEmptyTool()
    onToolsChange([...tools, newTool])
    setExpandedToolId(newTool.id)
    setIsOpen(true)
  }, [tools, onToolsChange])

  // Add example tool
  const handleAddExample = useCallback(() => {
    const newTool = { ...EXAMPLE_TOOL, id: crypto.randomUUID() }
    onToolsChange([...tools, newTool])
    setExpandedToolId(newTool.id)
    setIsOpen(true)
  }, [tools, onToolsChange])

  // Delete a tool
  const handleDeleteTool = useCallback(
    (toolId: string) => {
      onToolsChange(tools.filter((t) => t.id !== toolId))
      if (expandedToolId === toolId) {
        setExpandedToolId(null)
      }
    },
    [tools, onToolsChange, expandedToolId]
  )

  // Update a tool
  const handleUpdateTool = useCallback(
    (toolId: string, updates: Partial<Tool['function']>) => {
      onToolsChange(
        tools.map((t) =>
          t.id === toolId ? { ...t, function: { ...t.function, ...updates } } : t
        )
      )
    },
    [tools, onToolsChange]
  )

  // Update tool from JSON editor
  const handleToolJSONChange = useCallback(
    (toolId: string, jsonString: string) => {
      try {
        const parsed = JSON.parse(jsonString)
        onToolsChange(
          tools.map((t) =>
            t.id === toolId
              ? {
                  ...t,
                  function: {
                    name: parsed.name || t.function.name,
                    description: parsed.description,
                    parameters: parsed.parameters,
                    strict: parsed.strict,
                  },
                }
              : t
          )
        )
      } catch {
        // Invalid JSON, don't update
      }
    },
    [tools, onToolsChange]
  )

  // Prettify JSON in editor
  const handlePrettify = useCallback(() => {
    if (editorRef.current) {
      editorRef.current.getAction('editor.action.formatDocument')?.run()
    }
  }, [])

  // Handle editor mount
  const handleEditorMount: OnMount = useCallback((editor) => {
    editorRef.current = editor
  }, [])

  // Get current tool JSON for editor
  const getToolJSON = useCallback((tool: Tool): string => {
    return JSON.stringify(tool.function, null, 2)
  }, [])

  // Validate all tools
  const toolErrors = useMemo(() => {
    const errors: Record<string, string[]> = {}
    for (const tool of tools) {
      const toolValidationErrors = validateTool(tool)
      if (toolValidationErrors.length > 0) {
        errors[tool.id] = toolValidationErrors
      }
    }
    return errors
  }, [tools])

  const hasErrors = Object.keys(toolErrors).length > 0

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen} className={className}>
      <Card className="border-dashed">
        <CollapsibleTrigger asChild>
          <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {isOpen ? (
                  <ChevronDown className="h-4 w-4" />
                ) : (
                  <ChevronRight className="h-4 w-4" />
                )}
                <Wrench className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-medium">Tools / Function Calling</CardTitle>
                {tools.length > 0 && (
                  <Badge variant="secondary">
                    {tools.length} tool{tools.length !== 1 ? 's' : ''}
                  </Badge>
                )}
                {hasErrors && (
                  <Badge variant="destructive" className="gap-1">
                    <AlertCircle className="h-3 w-3" />
                    Errors
                  </Badge>
                )}
              </div>
            </div>
          </CardHeader>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="pt-0 space-y-4">
            {/* Tool Choice selector */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Label htmlFor="tool-choice" className="text-sm whitespace-nowrap">
                  Tool Choice:
                </Label>
                <Select
                  value={toolChoice?.toString() || 'auto'}
                  onValueChange={(value) =>
                    onToolChoiceChange(value === 'auto' ? null : (value as ToolChoice))
                  }
                  disabled={disabled || tools.length === 0}
                >
                  <SelectTrigger id="tool-choice" className="w-[140px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {TOOL_CHOICE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex-1" />

              {/* Add tool buttons */}
              <Button
                variant="outline"
                size="sm"
                onClick={handleAddExample}
                disabled={disabled}
                className="gap-1"
              >
                <Sparkles className="h-3 w-3" />
                Add Example
              </Button>
              <Button
                variant="default"
                size="sm"
                onClick={handleAddTool}
                disabled={disabled}
                className="gap-1"
              >
                <Plus className="h-3 w-3" />
                Add Tool
              </Button>
            </div>

            {/* Provider warning */}
            {provider === 'anthropic' && tools.length > 0 && (
              <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-3 py-2 rounded-md">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>
                  Anthropic uses a different tools format. Tools will be converted automatically.
                </span>
              </div>
            )}

            {/* Tools list */}
            {tools.length === 0 ? (
              <div className="text-sm text-muted-foreground text-center py-8 border rounded-md border-dashed">
                No tools defined. Add a tool to enable function calling.
              </div>
            ) : (
              <div className="space-y-3">
                {tools.map((tool) => (
                  <ToolCard
                    key={tool.id}
                    tool={tool}
                    isExpanded={expandedToolId === tool.id}
                    onToggle={() =>
                      setExpandedToolId(expandedToolId === tool.id ? null : tool.id)
                    }
                    onUpdate={(updates) => handleUpdateTool(tool.id, updates)}
                    onJSONChange={(json) => handleToolJSONChange(tool.id, json)}
                    onDelete={() => handleDeleteTool(tool.id)}
                    onPrettify={handlePrettify}
                    onEditorMount={handleEditorMount}
                    errors={toolErrors[tool.id]}
                    isDarkMode={isDarkMode}
                    disabled={disabled}
                  />
                ))}
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}

interface ToolCardProps {
  tool: Tool
  isExpanded: boolean
  onToggle: () => void
  onUpdate: (updates: Partial<Tool['function']>) => void
  onJSONChange: (json: string) => void
  onDelete: () => void
  onPrettify: () => void
  onEditorMount: OnMount
  errors?: string[]
  isDarkMode: boolean
  disabled: boolean
}

function ToolCard({
  tool,
  isExpanded,
  onToggle,
  onUpdate,
  onJSONChange,
  onDelete,
  onPrettify,
  onEditorMount,
  errors,
  isDarkMode,
  disabled,
}: ToolCardProps) {
  const [jsonValue, setJsonValue] = useState(() =>
    JSON.stringify(tool.function, null, 2)
  )

  // Sync local JSON when tool changes externally (only when collapsed)
  useEffect(() => {
    if (!isExpanded) {
      const toolJson = JSON.stringify(tool.function, null, 2)
      if (toolJson !== jsonValue) {
        setJsonValue(toolJson)
      }
    }
    // Note: jsonValue intentionally excluded to avoid infinite loop
    // We only want to sync when tool.function changes externally
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tool.function, isExpanded])

  const handleEditorChange = useCallback(
    (value: string | undefined) => {
      if (value !== undefined) {
        setJsonValue(value)
        onJSONChange(value)
      }
    },
    [onJSONChange]
  )

  return (
    <Card className={cn(errors && errors.length > 0 && 'border-destructive')}>
      <div
        className="flex items-center gap-3 p-3 cursor-pointer hover:bg-muted/50 transition-colors"
        onClick={onToggle}
      >
        {isExpanded ? (
          <ChevronDown className="h-4 w-4 flex-shrink-0" />
        ) : (
          <ChevronRight className="h-4 w-4 flex-shrink-0" />
        )}

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm font-medium truncate">
              {tool.function.name || 'Unnamed function'}
            </span>
            {errors && errors.length > 0 && (
              <Badge variant="destructive" className="text-xs">
                {errors.length} error{errors.length !== 1 ? 's' : ''}
              </Badge>
            )}
          </div>
          {tool.function.description && (
            <p className="text-xs text-muted-foreground truncate">
              {tool.function.description}
            </p>
          )}
        </div>

        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 flex-shrink-0"
              onClick={(e) => e.stopPropagation()}
              disabled={disabled}
            >
              <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete Tool</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to delete "{tool.function.name || 'this tool'}"? This
                action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={onDelete}>Delete</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {isExpanded && (
        <CardContent className="pt-0 space-y-4">
          {/* Quick edit fields */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor={`tool-name-${tool.id}`}>Function Name</Label>
              <Input
                id={`tool-name-${tool.id}`}
                value={tool.function.name}
                onChange={(e) => onUpdate({ name: e.target.value })}
                placeholder="function_name"
                className="font-mono"
                disabled={disabled}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor={`tool-desc-${tool.id}`}>Description</Label>
              <Input
                id={`tool-desc-${tool.id}`}
                value={tool.function.description || ''}
                onChange={(e) => onUpdate({ description: e.target.value })}
                placeholder="What does this function do?"
                disabled={disabled}
              />
            </div>
          </div>

          {/* Validation errors */}
          {errors && errors.length > 0 && (
            <div className="text-sm text-destructive space-y-1">
              {errors.map((error, i) => (
                <div key={i} className="flex items-center gap-2">
                  <AlertCircle className="h-3 w-3 flex-shrink-0" />
                  <span>{error}</span>
                </div>
              ))}
            </div>
          )}

          {/* JSON editor */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>Function Definition (JSON)</Label>
              <Button
                variant="outline"
                size="sm"
                onClick={onPrettify}
                disabled={disabled}
                className="h-7"
              >
                Prettify
              </Button>
            </div>
            <div className="border rounded-md overflow-hidden">
              <Editor
                height={200}
                language="json"
                theme={isDarkMode ? 'vs-dark' : 'light'}
                value={jsonValue}
                onChange={handleEditorChange}
                onMount={onEditorMount}
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: 'off',
                  folding: true,
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  tabSize: 2,
                  readOnly: disabled,
                  wordWrap: 'on',
                }}
                loading={
                  <div className="flex items-center justify-center h-[200px]">
                    <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                  </div>
                }
              />
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  )
}
