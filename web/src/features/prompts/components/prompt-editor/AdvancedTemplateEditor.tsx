'use client'

import { useState, useCallback, useEffect } from 'react'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { EyeIcon, Code2Icon, PlayIcon, Loader2Icon } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { TextTemplate, ChatTemplate, PromptType, TemplateDialect } from '../../types'
import { MonacoTemplateEditor } from './MonacoTemplateEditor'
import { DialectSelector } from './DialectSelector'
import { VariablePanel, VariableInputPanel, generateSampleValues } from './VariablePanel'
import { TemplatePreview, ValidationStatus, SyntaxErrorList } from './TemplatePreview'
import { useTemplateValidation } from '../../hooks/use-template-validation'
import { useTemplatePreview } from '../../hooks/use-template-preview'

interface AdvancedTemplateEditorProps {
  projectId: string
  type: PromptType
  value: TextTemplate | ChatTemplate
  onChange: (value: TextTemplate | ChatTemplate) => void
  initialDialect?: TemplateDialect
  readOnly?: boolean
  height?: string | number
  className?: string
}

/**
 * AdvancedTemplateEditor - Full-featured template editor with Monaco.
 *
 * Features:
 * - Monaco editor with syntax highlighting
 * - Dialect selection (Simple, Mustache, Jinja2)
 * - Real-time validation
 * - Live preview with sample values
 * - Variable extraction and display
 */
export function AdvancedTemplateEditor({
  projectId,
  type,
  value,
  onChange,
  initialDialect = 'auto',
  readOnly = false,
  height = 350,
  className,
}: AdvancedTemplateEditorProps) {
  const [dialect, setDialect] = useState<TemplateDialect>(initialDialect)
  const [previewVariables, setPreviewVariables] = useState<Record<string, unknown>>({})
  const [activeTab, setActiveTab] = useState<'edit' | 'preview'>('edit')

  // Get content as string for editor
  const content = type === 'text'
    ? (value as TextTemplate).content
    : JSON.stringify((value as ChatTemplate).messages, null, 2)

  // Validation hook
  const validation = useTemplateValidation({
    projectId,
    type,
    dialect,
    debounceMs: 500,
    enabled: true,
  })

  // Preview hook
  const preview = useTemplatePreview({
    projectId,
    type,
    dialect,
    debounceMs: 300,
    enabled: true,
  })

  // Handle content change
  const handleContentChange = useCallback(
    (newContent: string) => {
      let newValue: TextTemplate | ChatTemplate

      if (type === 'text') {
        newValue = { content: newContent }
      } else {
        try {
          const messages = JSON.parse(newContent)
          newValue = { messages }
        } catch {
          // Invalid JSON, keep as text for now
          newValue = value
        }
      }

      onChange(newValue)
      validation.validate(newValue)
    },
    [type, value, onChange, validation]
  )

  // Handle dialect change
  const handleDialectChange = useCallback(
    (newDialect: TemplateDialect) => {
      setDialect(newDialect)
    },
    []
  )

  // Handle preview
  const handlePreview = useCallback(() => {
    // Auto-generate sample values if not set
    let vars = previewVariables
    if (Object.keys(vars).length === 0 && validation.variables.length > 0) {
      vars = generateSampleValues(validation.variables)
      setPreviewVariables(vars)
    }
    preview.preview(value, vars)
  }, [value, previewVariables, validation.variables, preview])

  // Auto-update preview variables when new variables are detected
  useEffect(() => {
    if (validation.variables.length > 0) {
      setPreviewVariables((prev) => {
        const updated = { ...prev }
        for (const variable of validation.variables) {
          if (!(variable in updated)) {
            const samples = generateSampleValues([variable])
            updated[variable] = samples[variable]
          }
        }
        return updated
      })
    }
  }, [validation.variables])

  // Validate on mount
  useEffect(() => {
    validation.validate(value)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className={cn('flex flex-col gap-4', className)}>
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Label className="text-sm font-medium">Dialect</Label>
          <DialectSelector
            value={dialect}
            onChange={handleDialectChange}
            disabled={readOnly}
          />
        </div>
        <div className="flex items-center gap-2">
          <ValidationStatus
            isValid={validation.isValid}
            isValidating={validation.isValidating}
            errorCount={validation.errors.length}
            warningCount={validation.warnings.length}
          />
        </div>
      </div>

      <Separator />

      {/* Main content area */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'edit' | 'preview')}>
        <div className="flex items-center justify-between mb-2">
          <TabsList>
            <TabsTrigger value="edit" className="gap-1.5">
              <Code2Icon className="size-4" />
              Editor
            </TabsTrigger>
            <TabsTrigger value="preview" className="gap-1.5">
              <EyeIcon className="size-4" />
              Preview
            </TabsTrigger>
          </TabsList>
          {activeTab === 'preview' && (
            <Button
              size="sm"
              variant="outline"
              onClick={handlePreview}
              disabled={preview.isCompiling}
            >
              {preview.isCompiling ? (
                <Loader2Icon className="size-4 animate-spin mr-1.5" />
              ) : (
                <PlayIcon className="size-4 mr-1.5" />
              )}
              Compile
            </Button>
          )}
        </div>

        <TabsContent value="edit" className="mt-0">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {/* Editor - 2/3 width on large screens */}
            <div className="lg:col-span-2">
              <MonacoTemplateEditor
                value={content}
                onChange={handleContentChange}
                dialect={validation.detectedDialect || dialect}
                variables={validation.variables}
                errors={validation.errors}
                readOnly={readOnly}
                height={height}
              />
            </div>

            {/* Variables panel - 1/3 width on large screens */}
            <div className="space-y-4">
              <VariablePanel
                variables={validation.variables}
                isLoading={validation.isValidating}
                dialect={validation.detectedDialect || dialect}
              />

              {validation.errors.length > 0 && (
                <SyntaxErrorList errors={validation.errors} />
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="preview" className="mt-0">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Variable inputs */}
            <div>
              <VariableInputPanel
                variables={validation.variables}
                values={previewVariables}
                onChange={setPreviewVariables}
              />
            </div>

            {/* Preview output */}
            <div>
              <TemplatePreview
                compiled={preview.compiled}
                type={type}
                dialect={preview.detectedDialect || dialect}
                isLoading={preview.isCompiling}
                error={preview.error}
              />
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

/**
 * Simplified version for quick edits without full preview.
 */
interface SimpleTemplateEditorProps {
  projectId: string
  type: PromptType
  value: TextTemplate | ChatTemplate
  onChange: (value: TextTemplate | ChatTemplate) => void
  dialect?: TemplateDialect
  readOnly?: boolean
  height?: string | number
  className?: string
}

export function SimpleTemplateEditor({
  projectId,
  type,
  value,
  onChange,
  dialect = 'auto',
  readOnly = false,
  height = 200,
  className,
}: SimpleTemplateEditorProps) {
  // Get content as string for editor
  const content = type === 'text'
    ? (value as TextTemplate).content
    : JSON.stringify((value as ChatTemplate).messages, null, 2)

  // Validation hook
  const validation = useTemplateValidation({
    projectId,
    type,
    dialect,
    debounceMs: 500,
    enabled: true,
  })

  // Handle content change
  const handleContentChange = useCallback(
    (newContent: string) => {
      let newValue: TextTemplate | ChatTemplate

      if (type === 'text') {
        newValue = { content: newContent }
      } else {
        try {
          const messages = JSON.parse(newContent)
          newValue = { messages }
        } catch {
          // Invalid JSON, keep current value
          return
        }
      }

      onChange(newValue)
      validation.validate(newValue)
    },
    [type, onChange, validation]
  )

  // Validate on mount
  useEffect(() => {
    validation.validate(value)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        <Label className="text-sm font-medium">Template</Label>
        <ValidationStatus
          isValid={validation.isValid}
          isValidating={validation.isValidating}
          errorCount={validation.errors.length}
          warningCount={validation.warnings.length}
        />
      </div>

      <MonacoTemplateEditor
        value={content}
        onChange={handleContentChange}
        dialect={validation.detectedDialect || dialect}
        variables={validation.variables}
        errors={validation.errors}
        readOnly={readOnly}
        height={height}
      />

      <VariablePanel
        variables={validation.variables}
        isLoading={validation.isValidating}
        dialect={validation.detectedDialect || dialect}
      />
    </div>
  )
}
