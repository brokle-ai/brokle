'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
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
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  ChevronDown,
  ChevronRight,
  FileJson2,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ResponseFormat, JsonSchemaResponseFormat } from '../types'

interface ResponseFormatEditorProps {
  responseFormat: ResponseFormat | null
  onResponseFormatChange: (format: ResponseFormat | null) => void
  provider?: string // To show provider compatibility warnings
  hasTools?: boolean // Show warning if tools are also enabled
  className?: string
  disabled?: boolean
}

type FormatType = 'text' | 'json_object' | 'json_schema'

const DEFAULT_SCHEMA = {
  type: 'object',
  properties: {
    result: {
      type: 'string',
      description: 'The result of the operation',
    },
  },
  required: ['result'],
  additionalProperties: false,
}

export function ResponseFormatEditor({
  responseFormat,
  onResponseFormatChange,
  provider,
  hasTools = false,
  className,
  disabled = false,
}: ResponseFormatEditorProps) {
  const { resolvedTheme } = useTheme()
  const isDarkMode = resolvedTheme === 'dark'
  const [isOpen, setIsOpen] = useState(
    responseFormat !== null && responseFormat.type !== 'text'
  )
  const editorRef = useRef<MonacoEditor.editor.IStandaloneCodeEditor | null>(null)

  // Local state for JSON schema editing
  const [schemaName, setSchemaName] = useState(() =>
    responseFormat?.type === 'json_schema'
      ? (responseFormat as JsonSchemaResponseFormat).json_schema.name
      : 'response_schema'
  )
  const [strictMode, setStrictMode] = useState(() =>
    responseFormat?.type === 'json_schema'
      ? (responseFormat as JsonSchemaResponseFormat).json_schema.strict ?? true
      : true
  )
  const [schemaJson, setSchemaJson] = useState(() =>
    responseFormat?.type === 'json_schema'
      ? JSON.stringify(
          (responseFormat as JsonSchemaResponseFormat).json_schema.schema,
          null,
          2
        )
      : JSON.stringify(DEFAULT_SCHEMA, null, 2)
  )
  const [jsonError, setJsonError] = useState<string | null>(null)

  // Get current format type
  const formatType: FormatType = responseFormat
    ? (responseFormat.type as FormatType)
    : 'text'

  // Update parent when schema values change
  const updateJsonSchemaFormat = useCallback(() => {
    try {
      const schema = JSON.parse(schemaJson)
      setJsonError(null)
      const format: JsonSchemaResponseFormat = {
        type: 'json_schema',
        json_schema: {
          name: schemaName || 'response_schema',
          schema,
          strict: strictMode,
        },
      }
      onResponseFormatChange(format)
    } catch (e) {
      setJsonError((e as Error).message)
    }
  }, [schemaName, strictMode, schemaJson, onResponseFormatChange])

  // Handle format type change
  const handleFormatTypeChange = useCallback(
    (type: FormatType) => {
      if (type === 'text') {
        onResponseFormatChange(null) // null = no response_format, default text output
      } else if (type === 'json_object') {
        onResponseFormatChange({ type: 'json_object' })
      } else if (type === 'json_schema') {
        // Use current schema values
        try {
          const schema = JSON.parse(schemaJson)
          onResponseFormatChange({
            type: 'json_schema',
            json_schema: {
              name: schemaName || 'response_schema',
              schema,
              strict: strictMode,
            },
          })
        } catch {
          // If schema is invalid, set with default
          onResponseFormatChange({
            type: 'json_schema',
            json_schema: {
              name: schemaName || 'response_schema',
              schema: DEFAULT_SCHEMA,
              strict: strictMode,
            },
          })
        }
      }
    },
    [schemaName, strictMode, schemaJson, onResponseFormatChange]
  )

  // Handle editor mount
  const handleEditorMount: OnMount = useCallback((editor) => {
    editorRef.current = editor
  }, [])

  // Prettify JSON
  const handlePrettify = useCallback(() => {
    if (editorRef.current) {
      editorRef.current.getAction('editor.action.formatDocument')?.run()
    }
  }, [])

  // Handle schema JSON change
  const handleSchemaChange = useCallback(
    (value: string | undefined) => {
      if (value !== undefined) {
        setSchemaJson(value)
        try {
          const schema = JSON.parse(value)
          setJsonError(null)
          if (formatType === 'json_schema') {
            onResponseFormatChange({
              type: 'json_schema',
              json_schema: {
                name: schemaName || 'response_schema',
                schema,
                strict: strictMode,
              },
            })
          }
        } catch (e) {
          setJsonError((e as Error).message)
        }
      }
    },
    [schemaName, strictMode, formatType, onResponseFormatChange]
  )

  // Handle schema name change
  const handleSchemaNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newName = e.target.value
      setSchemaName(newName)
      if (formatType === 'json_schema') {
        try {
          const schema = JSON.parse(schemaJson)
          onResponseFormatChange({
            type: 'json_schema',
            json_schema: {
              name: newName || 'response_schema',
              schema,
              strict: strictMode,
            },
          })
        } catch {
          // Keep current format if schema is invalid
        }
      }
    },
    [formatType, schemaJson, strictMode, onResponseFormatChange]
  )

  // Handle strict mode change
  const handleStrictModeChange = useCallback(
    (checked: boolean) => {
      setStrictMode(checked)
      if (formatType === 'json_schema') {
        try {
          const schema = JSON.parse(schemaJson)
          onResponseFormatChange({
            type: 'json_schema',
            json_schema: {
              name: schemaName || 'response_schema',
              schema,
              strict: checked,
            },
          })
        } catch {
          // Keep current format if schema is invalid
        }
      }
    },
    [formatType, schemaName, schemaJson, onResponseFormatChange]
  )

  // Check if provider supports response_format
  const isProviderSupported = provider !== 'anthropic'

  // Format summary for collapsed state
  const formatSummary =
    formatType === 'text'
      ? 'Text (default)'
      : formatType === 'json_object'
        ? 'JSON Object'
        : `JSON Schema: ${schemaName}`

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
                <FileJson2 className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-medium">Response Format</CardTitle>
                {formatType !== 'text' && (
                  <Badge variant="secondary">{formatSummary}</Badge>
                )}
                {jsonError && formatType === 'json_schema' && (
                  <Badge variant="destructive" className="gap-1">
                    <AlertCircle className="h-3 w-3" />
                    Invalid Schema
                  </Badge>
                )}
              </div>
            </div>
          </CardHeader>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="pt-0 space-y-4">
            {/* Provider warnings */}
            {!isProviderSupported && (
              <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-3 py-2 rounded-md">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>
                  Anthropic does not support response_format. Structured output will be
                  ignored for this provider.
                </span>
              </div>
            )}

            {/* Tools + response_format warning */}
            {hasTools && formatType !== 'text' && (
              <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-3 py-2 rounded-md">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>
                  Using both tools and structured output together may have unexpected
                  behavior. Consider using one or the other.
                </span>
              </div>
            )}

            {/* Format type selector */}
            <RadioGroup
              value={formatType}
              onValueChange={(v) => handleFormatTypeChange(v as FormatType)}
              disabled={disabled}
              className="flex gap-6"
            >
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="text" id="format-text" />
                <Label htmlFor="format-text" className="cursor-pointer">
                  Text (default)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="json_object" id="format-json" />
                <Label htmlFor="format-json" className="cursor-pointer">
                  JSON Object
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="json_schema" id="format-schema" />
                <Label htmlFor="format-schema" className="cursor-pointer">
                  JSON Schema
                </Label>
              </div>
            </RadioGroup>

            {/* JSON Object mode info */}
            {formatType === 'json_object' && (
              <p className="text-sm text-muted-foreground">
                The model will output valid JSON. You should instruct the model to produce
                JSON in your system message.
              </p>
            )}

            {/* JSON Schema mode */}
            {formatType === 'json_schema' && (
              <div className="space-y-4">
                {/* Schema name and strict mode */}
                <div className="flex items-end gap-4">
                  <div className="flex-1 space-y-2">
                    <Label htmlFor="schema-name">Schema Name (required)</Label>
                    <Input
                      id="schema-name"
                      value={schemaName}
                      onChange={handleSchemaNameChange}
                      placeholder="response_schema"
                      className="font-mono"
                      disabled={disabled}
                    />
                  </div>
                  <div className="flex items-center gap-2 pb-2">
                    <Switch
                      id="strict-mode"
                      checked={strictMode}
                      onCheckedChange={handleStrictModeChange}
                      disabled={disabled}
                    />
                    <Label htmlFor="strict-mode" className="cursor-pointer">
                      Strict mode
                    </Label>
                  </div>
                </div>

                {/* JSON Schema editor */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>JSON Schema Definition</Label>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handlePrettify}
                      disabled={disabled}
                      className="h-7"
                    >
                      Prettify
                    </Button>
                  </div>

                  {jsonError && (
                    <div className="flex items-center gap-2 text-sm text-destructive">
                      <AlertCircle className="h-3 w-3 flex-shrink-0" />
                      <span>Invalid JSON: {jsonError}</span>
                    </div>
                  )}

                  <div
                    className={cn(
                      'border rounded-md overflow-hidden',
                      jsonError && 'border-destructive'
                    )}
                  >
                    <Editor
                      height={200}
                      language="json"
                      theme={isDarkMode ? 'vs-dark' : 'light'}
                      value={schemaJson}
                      onChange={handleSchemaChange}
                      onMount={handleEditorMount}
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

                <p className="text-xs text-muted-foreground">
                  Define a JSON Schema to constrain the model's output. The model will
                  produce a response that matches your schema exactly.
                </p>
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}
